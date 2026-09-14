import { useState } from 'react'

import { matchRecap } from '@/features/game/lib/game-result'
import type { GameSnapshot, GameViewer } from '@/features/game/types/game.types'

export type OutcomeCinematicKind = 'victory' | 'defeat'

export type OutcomeCinematicState = {
  visible: boolean
  kind: OutcomeCinematicKind | null
  /** True once the viewer has seen it and it's no longer auto-shown -
   * MatchResultScreen's "Revivir" button is only meaningful then. */
  canReplay: boolean
  /** Changes every time a fresh play-through starts (first show, or each
   * replay) - pass as OutcomeCinematic's React `key` so it always mounts
   * fresh instead of carrying stale skip-timer state from a previous
   * viewing (see that component's own doc comment on why it has no
   * `visible` prop of its own). */
  playToken: string
  /** Re-shows the cinematic on demand (the "Revivir" button). */
  replay: () => void
  /** Called by the cinematic itself, once, when it finishes or is skipped. */
  dismiss: () => void
}

// Decides whether the end-of-game cinematic should be showing right now,
// and drives its one-shot-then-replayable lifecycle - kept out of
// lobby-room-screen.tsx for the same reason match-rules.ts/game-result.ts
// are split out: this judgment call (when does the cinematic show, is it
// applicable at all) is worth unit-testing/reading on its own.
//
// An aborted game never gets one - see matchRecap/roundOutcome's own
// GAUNTLET-FALL precedent: there is neither a winner nor a loser to
// dramatize. Each viewer's own snapshot.state/matchRecap already differs by
// nothing except perspective (GAUNTLET is collective, VERSUS is per-seat via
// MatchOutcome.won) - see outcome-cinematic.tsx's own doc comment - so this
// hook needs no participantId beyond what `you` already carries.
export function useOutcomeCinematic(
  snapshot: GameSnapshot,
  you: GameViewer
): OutcomeCinematicState {
  const recap = matchRecap(snapshot, you)
  const eligible = snapshot.state === 'FINISHED' && !recap.aborted

  const kind: OutcomeCinematicKind | null = !eligible
    ? null
    : recap.mode === 'GAUNTLET'
      ? recap.squadSurvived
        ? 'victory'
        : 'defeat'
      : (() => {
          const won = recap.outcomes.find((o) => o.isSelf)?.won
          return won === null || won === undefined ? null : won ? 'victory' : 'defeat'
        })()

  // A rematch produces a new game id - the cinematic must be eligible to
  // show again for it, not stay dismissed from the previous match. Rather
  // than resetting seen/replaying in an effect or a render-time ref
  // mutation (both rejected by this project's stricter hooks lint - see
  // react-hooks/set-state-in-effect and react-hooks/refs), the state itself
  // remembers which game id it describes; a mismatch is simply treated as
  // "fresh" for this render, with no mutation of anything until dismiss/
  // replay is actually called.
  const initial = { gameId: snapshot.id, seen: false, replaying: false, plays: 0 }
  const [session, setSession] = useState(initial)
  const { seen, replaying, plays } = session.gameId === snapshot.id ? session : initial

  return {
    visible: kind !== null && (!seen || replaying),
    kind,
    canReplay: kind !== null && seen && !replaying,
    playToken: `${snapshot.id}:${plays}`,
    replay: () => setSession({ gameId: snapshot.id, seen, replaying: true, plays: plays + 1 }),
    dismiss: () => setSession({ gameId: snapshot.id, seen: true, replaying: false, plays }),
  }
}
