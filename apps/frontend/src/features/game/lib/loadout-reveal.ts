import { currentRound, hasAllLoadouts } from '@/features/game/lib/match-rules'
import type { LiveMatchState } from '@/features/game/stores/game-socket.store'
import type { GameSnapshot } from '@/features/game/types/game.types'

// revealOrder lists participant ids in the order their loadout card should
// flip face-up, with the caller's own id always moved to the end (self
// revealed last, for the "climax" reveal). VERSUS groups by team (iterating
// snapshot.teams so the two columns fill in team order); GAUNTLET just uses
// the roster order as-is. Pure and deterministic - no randomness, no
// dependency on render state.
export function revealOrder(snapshot: GameSnapshot, selfParticipantId: string): string[] {
  const ids: string[] =
    snapshot.mode === 'VERSUS'
      ? snapshot.teams.flatMap((team) =>
          snapshot.participants.filter((p) => p.teamId === team.id).map((p) => p.id)
        )
      : snapshot.participants.map((p) => p.id)

  const withoutSelf = ids.filter((id) => id !== selfParticipantId)
  return ids.includes(selfParticipantId) ? [...withoutSelf, selfParticipantId] : withoutSelf
}

// shouldReveal gates the reveal on BOTH "an assignment frame arrived since
// the last reveal" AND "the snapshot has actually caught up" - LOADOUTS_
// ASSIGNED is sent before its own pushCurrentState, so the frame alone is
// not proof the snapshot carries the new loadouts yet (see game-socket.store
// .ts's LOADOUTS_ASSIGNED case).
export function shouldReveal(live: LiveMatchState, snapshot: GameSnapshot): boolean {
  if (live.assignmentSeq <= live.revealedAssignmentSeq) return false
  if (!hasAllLoadouts(snapshot)) return false
  if (live.assignedRoundIndex === null) return true
  const round = currentRound(snapshot)
  return round !== null && round.index === live.assignedRoundIndex
}

// revealStepMs is the delay between each card flipping, zeroed entirely
// under reduced motion (an instant reveal, not a fast one).
export function revealStepMs(reducedMotion: boolean): number {
  return reducedMotion ? 0 : 550
}
