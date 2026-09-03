import { useEffect, useRef, useState } from 'react'

import {
  revealTimeline,
  type RevealPhaseKind,
  type RevealPlayer,
} from '@/features/game/lib/loadout-reveal'
import type { GameParticipant } from '@/features/game/types/game.types'
import type { Manga, RevealSpeed } from '@/shared/contracts/enums'

type Params = {
  gameId: string
  roundIndex: number
  mangas: Manga[]
  /** Ordered exactly like the backend's Game.Participants() (join order) -
   * the reveal plays them in this order, one full turn each. */
  participants: GameParticipant[]
  speed: RevealSpeed
  active: boolean
  markRevealed: () => void
  /** Sends the REVEAL_READY command - the server-side half of the
   * synchronized skip (owner decision, 2026-08-30): once every connected
   * human has called this, GameService cuts the pending reveal timer short
   * for everyone, not just the caller's own client. */
  sendRevealReady: () => void
  /** The ASSIGNING phase's authoritative close instant - LOADOUTS_ASSIGNED's
   * own closesAt (via the socket store's live.revealEndsAt), or a
   * STATE-adopted game.revealEndsAt on reconnect - since
   * GameService.scheduleRevealDelay is what actually decides when voting
   * opens. The local timeline is scaled to fit whatever time remains until
   * this instant rather than trusting a transported duration, so hub-
   * delivery latency degrades pacing instead of desyncing "reveal looks
   * done" from "voting is actually open". null before any frame carrying
   * it has arrived (e.g. a client still on the very first STATE fetch). */
  revealEndsAt: number | null
}

type Result = {
  isRevealing: boolean
  phase: RevealPhaseKind
  /** Index into `participants` - whose turn is currently playing. -1
   * during the lobby-wide 'intro'/'outro'. */
  participantIndex: number
  /** The slot currently spinning/holding for the current participant's
   * turn, index into playerSlots(mangas, that participant) - -1 outside a
   * slot phase. Varies per participant (see playerSlots' doc), so
   * `totalSlots` below is carried per-phase, not a single lobby constant. */
  slotIndex: number
  totalSlots: number
  /** Marks this client's own human ready to skip ahead - sends
   * REVEAL_READY over the socket AND ends this client's own local
   * animation immediately, so a lone player never has to sit through their
   * own already-acknowledged reveal waiting on the server's timer. */
  skip: () => void
}

// Drives the sorteo overlay: jugador-por-jugador (owner request,
// 2026-08-30 - see ObsidianVault/game-match-assignment-frontend.md for the
// all-lanes-at-once design this supersedes), paced by
// revealTimeline(...) and scaled to fit the time remaining until
// revealEndsAt (LOADOUTS_ASSIGNED's own authoritative closesAt) so a
// constants drift between backend and frontend degrades the pacing rather
// than desyncing "reveal done" from "voting actually open" (the backend's
// own timer, not this hook, is what truly gates OpenVoting).
//
// The bug this hook's predecessor had (fixed 2026-08-14, see
// game-match-assignment-frontend.md): the scheduling effect returned
// `clearTimers` as its cleanup, keyed on `[active, ...]`. But
// `markRevealed()` flips `active` back to `false` on the very next render
// (it catches `revealedAssignmentSeq` up to `assignmentSeq` in the store) -
// so React ran that cleanup one render after scheduling, cancelling every
// timer before the first one could ever fire. Fix, preserved here: this
// effect never returns a cleanup tied to `active` flipping - pending timers
// are only cleared (a) right before a genuinely NEW sequence schedules its
// own timers, and (b) on unmount, via a separate effect with an empty
// dependency array.
export function useLoadoutReveal({
  gameId,
  roundIndex,
  mangas,
  participants,
  speed,
  active,
  markRevealed,
  sendRevealReady,
  revealEndsAt,
}: Params): Result {
  const mangasKey = mangas.slice().sort().join(',')
  const players: RevealPlayer[] = participants.map((p) => ({
    hasStand: !!p.loadout?.stand,
    hasDevilFruit: !!p.loadout?.devilFruit,
    hasArmamentHaki: p.loadout?.armamentHaki !== undefined && p.loadout.armamentHaki !== 'NONE',
    hasObservationHaki: p.loadout?.observationHaki !== undefined && p.loadout.observationHaki !== 'NONE',
    hasConquerorHaki: p.loadout?.conquerorHaki !== undefined && p.loadout.conquerorHaki !== 'NONE',
  }))
  const playersKey = players
    .map((p) =>
      [p.hasStand, p.hasDevilFruit, p.hasArmamentHaki, p.hasObservationHaki, p.hasConquerorHaki]
        .map((b) => (b ? 1 : 0))
        .join('')
    )
    .join(':')
  const phases = revealTimeline(gameId, roundIndex, mangas, players, speed)
  const localTotalMs = phases.reduce((sum, p) => sum + p.durationMs, 0)
  // runKey is an absolute epoch value (stable across renders), never a
  // Date.now()-derived one - the actual "how much time is left" read only
  // happens once, inside the scheduling effect below, at the moment a run
  // is genuinely (re)started.
  const runKey = `${gameId}:${roundIndex}:${mangasKey}:${playersKey}:${speed}:${revealEndsAt ?? 'local'}`

  const [phaseIndex, setPhaseIndex] = useState(0)
  const [revealing, setRevealing] = useState(false)
  const [seededKey, setSeededKey] = useState<string | null>(null)
  const timers = useRef<ReturnType<typeof setTimeout>[]>([])
  const startedRef = useRef<string | null>(null)

  const clearTimers = () => {
    timers.current.forEach(clearTimeout)
    timers.current = []
  }

  // Resets the phase index the moment a genuinely new reveal starts, during
  // render rather than inside an effect (React's own recommended pattern for
  // "reset state when a derived key changes") - lands before paint and
  // avoids react-hooks/set-state-in-effect's cascading-render warning for an
  // unconditional setState in an effect body.
  if (active && runKey !== seededKey) {
    setSeededKey(runKey)
    if (phases.length > 0) {
      setPhaseIndex(0)
      setRevealing(true)
    }
  }

  useEffect(() => {
    if (!active || startedRef.current === runKey) return
    startedRef.current = runKey

    markRevealed()
    // Clear any timers left over from a previous sequence before scheduling
    // this one - NOT returned as this effect's cleanup (see the file-level
    // comment above for why that distinction is the actual fix).
    clearTimers()

    if (phases.length === 0) return

    // Read once per run, right as it starts - Date.now() in the render body
    // would make this impure per-render, which is exactly what the
    // exhaustive-deps disable below already documents for phases/scale.
    // remainingMs === 0 (a closesAt already in the past, e.g. a laggy
    // reconnect) collapses every timer to 0ms and the reveal finishes
    // instantly - correct: the server is about to open SUMMARY/VOTING
    // regardless, so there is nothing left to animate toward.
    const remainingMs = revealEndsAt !== null ? Math.max(0, revealEndsAt - Date.now()) : null
    const scale = remainingMs !== null && localTotalMs > 0 ? remainingMs / localTotalMs : 1

    let elapsedMs = 0
    phases.forEach((p, index) => {
      elapsedMs += p.durationMs * scale
      const timer = setTimeout(() => {
        setPhaseIndex((current) => Math.max(current, index + 1))
        if (index === phases.length - 1) setRevealing(false)
      }, elapsedMs)
      timers.current.push(timer)
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps -- startedRef (keyed on runKey) guards re-entry; phases/scale/markRevealed are read fresh on the run they gate, not meant to re-trigger it on their own
  }, [active, runKey])

  // Only clears pending timers on unmount - deliberately not tied to
  // `active`/`runKey` changing (that's the bug described above).
  useEffect(() => clearTimers, [])

  const skip = () => {
    sendRevealReady()
    clearTimers()
    setPhaseIndex(phases.length)
    setRevealing(false)
  }

  if (!revealing) {
    return {
      isRevealing: false,
      phase: 'outro',
      participantIndex: -1,
      slotIndex: -1,
      totalSlots: 0,
      skip: () => {},
    }
  }

  const current = phases[Math.min(phaseIndex, phases.length - 1)]
  const participantIndex = current.phase.participant ?? -1
  const slotIndex = current.phase.slot ?? -1
  const totalSlots = current.phase.totalSlots ?? 0
  return { isRevealing: true, phase: current.phase.kind, participantIndex, slotIndex, totalSlots, skip }
}
