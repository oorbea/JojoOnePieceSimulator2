import { useEffect, useRef, useState } from 'react'

import {
  revealTimeline,
  seekRevealTimeline,
  type RevealPhaseKind,
  type RevealPlayer,
} from '@/features/game/lib/loadout-reveal'
import type { GameParticipant } from '@/features/game/types/game.types'
import type { Manga, RevealSpeed } from '@/shared/contracts/enums'
import { serverNow } from '@/shared/lib/server-clock'

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
  /** The reveal timer's own arm instant - LOADOUTS_ASSIGNED's own
   * revealStartedAt (via the socket store's live.revealStartedAt), or a
   * STATE-adopted game.revealStartedAt on reconnect. Paired with
   * revealEndsAt so a client that mounts the reveal late (a slow device, a
   * laggy STATE, a catalog still loading) seeks its local timeline to
   * wherever the server's own window says "now" actually is, instead of
   * always starting over at phase 0. null has the same "not known yet"
   * meaning as revealEndsAt's own null (see seekRevealTimeline's fallback). */
  revealStartedAt: number | null
  /** True while the server's own snapshot.state is still ASSIGNING - the
   * authority that actually ends a reveal. A client's own local timeline is
   * only ever a rehearsal of the server's real pacing (see the module doc);
   * the moment the server has actually moved the game on to SUMMARY/VOTING,
   * this flips false and the reveal stops immediately, however far its own
   * timers still had left to run. Without this, a client whose local
   * timeline ran even slightly long (a stalled JS thread, a scale
   * mismatch) kept animating the sorteo well after voting had already
   * opened underneath it. */
  stillAssigning: boolean
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
  /** How much every phase's local duration is being stretched/squeezed to
   * fit the server's actual reveal window - PowerRoulette's own spinMs must
   * be multiplied by this too (see reveal-stage.tsx), or its animation
   * drifts out of step with the rest of the timeline the moment the two
   * ever disagree (a slow device, a rejoined tab, ...). 1 while nothing is
   * known to scale against. */
  scale: number
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
  revealStartedAt,
  stillAssigning,
}: Params): Result {
  const mangasKey = mangas.slice().sort().join(',')
  const players: RevealPlayer[] = participants.map((p) => ({
    hasStand: !!p.loadout?.stand,
    hasDevilFruit: !!p.loadout?.devilFruit,
    hasArmamentHaki: p.loadout?.armamentHaki !== undefined && p.loadout.armamentHaki !== 'NONE',
    hasObservationHaki:
      p.loadout?.observationHaki !== undefined && p.loadout.observationHaki !== 'NONE',
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
  const [runScale, setRunScale] = useState(1)
  const timers = useRef<ReturnType<typeof setTimeout>[]>([])
  const startedRef = useRef<string | null>(null)

  const clearTimers = () => {
    timers.current.forEach(clearTimeout)
    timers.current = []
  }

  // seek is where the timeline should ALREADY be right now - not always
  // phase 0. Computed here (during render, gated by the same
  // `runKey !== seededKey` one-shot as the phaseIndex reset below) so the
  // very first paint of a genuinely new run already shows the right spot
  // instead of flashing phase 0 first and correcting a tick later. Reading
  // serverNow() here is the same accepted pattern the effect below already
  // documents for its own once-per-run reads - it only ever runs once per
  // actual run, guarded the same way setPhaseIndex(0) always was.
  const seek =
    active && runKey !== seededKey
      ? seekRevealTimeline(phases, localTotalMs, revealStartedAt, revealEndsAt, serverNow())
      : null

  // Resets the phase index the moment a genuinely new reveal starts, during
  // render rather than inside an effect (React's own recommended pattern for
  // "reset state when a derived key changes") - lands before paint and
  // avoids react-hooks/set-state-in-effect's cascading-render warning for an
  // unconditional setState in an effect body.
  if (seek) {
    setSeededKey(runKey)
    setRunScale(seek.scale)
    if (phases.length > 0) {
      setPhaseIndex(seek.startIndex)
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

    // Read once per run, right as it starts - serverNow() in the render body
    // would make this impure per-render (the seek above is the one accepted
    // exception, gated identically). A revealEndsAt already in the past
    // (e.g. a laggy reconnect) collapses every remaining timer to 0ms and
    // the reveal finishes instantly - correct: the server is about to open
    // SUMMARY/VOTING regardless, so there is nothing left to animate toward.
    const { scale, startIndex, offsetIntoPhaseMs } = seekRevealTimeline(
      phases,
      localTotalMs,
      revealStartedAt,
      revealEndsAt,
      serverNow()
    )

    // The phase we're seeking INTO fires after its own remaining time
    // (durationMs - however far in we already are); every phase after that
    // fires after its own full scaled duration, same as before.
    let elapsedMs = 0
    phases.forEach((p, index) => {
      if (index < startIndex) return
      const durationMs = p.durationMs * scale
      elapsedMs += index === startIndex ? Math.max(0, durationMs - offsetIntoPhaseMs) : durationMs
      const timer = setTimeout(() => {
        setPhaseIndex((current) => Math.max(current, index + 1))
        if (index === phases.length - 1) setRevealing(false)
      }, elapsedMs)
      timers.current.push(timer)
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps -- startedRef (keyed on runKey) guards re-entry; phases/revealStartedAt/revealEndsAt/markRevealed are read fresh on the run they gate, not meant to re-trigger it on their own
  }, [active, runKey])

  // Only clears pending timers on unmount - deliberately not tied to
  // `active`/`runKey` changing (that's the bug described above). Any timer
  // still pending when stillAssigning flips false below fires into a no-op
  // (revealing is already false by then) rather than being cancelled eagerly
  // - simpler than a second effect, and harmless.
  useEffect(() => clearTimers, [])

  // Server authority overrides local pacing, adjusted during render rather
  // than an effect (same accepted pattern as the seed reset above - React's
  // own "adjust state when a derived condition changes", not a setState-in-
  // effect cascade): the instant the game is no longer ASSIGNING, this
  // reveal is over, full stop, however far its own timers still had left to
  // run - see stillAssigning's doc.
  if (revealing && !stillAssigning) {
    setRevealing(false)
  }

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
      scale: 1,
      skip: () => {},
    }
  }

  const current = phases[Math.min(phaseIndex, phases.length - 1)]
  const participantIndex = current.phase.participant ?? -1
  const slotIndex = current.phase.slot ?? -1
  const totalSlots = current.phase.totalSlots ?? 0
  return {
    isRevealing: true,
    phase: current.phase.kind,
    participantIndex,
    slotIndex,
    totalSlots,
    scale: runScale,
    skip,
  }
}
