import { currentRound, hasAllLoadouts, loadoutSlots } from '@/features/game/lib/match-rules'
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

// A single tick of the reveal sequence: either "flip this participant's
// card face-up" (slot === -1) or "show one more of that card's loadout
// slots, in loadoutSlots' draw order" (slot === the slot's index). Flat
// across every participant, in revealOrder, so use-loadout-reveal.ts can
// walk one global timeline instead of per-card timers.
export type RevealStep = { participantId: string; slot: number }

// revealSteps expands revealOrder into the full poder-a-poder timeline: for
// each participant in turn, a flip step followed by one step per loadout
// slot they actually have (loadoutSlots already gates by
// snapshot.config.mangas - a JoJo-only lobby never gets a physicalForm/
// devilFruit/haki step, and vice versa). Assumes every included participant
// already has a loadout, which shouldReveal's hasAllLoadouts check
// guarantees by the time this is called.
export function revealSteps(snapshot: GameSnapshot, selfParticipantId: string): RevealStep[] {
  const order = revealOrder(snapshot, selfParticipantId)
  const byId = new Map(snapshot.participants.map((p) => [p.id, p]))
  const steps: RevealStep[] = []

  for (const participantId of order) {
    steps.push({ participantId, slot: -1 })
    const loadout = byId.get(participantId)?.loadout
    if (!loadout) continue
    const slotCount = loadoutSlots(loadout, snapshot.config.mangas).length
    for (let slot = 0; slot < slotCount; slot++) {
      steps.push({ participantId, slot })
    }
  }

  return steps
}

// flipStepMs/slotStepMs are the delay for a card-flip step vs. a
// single-slot reveal step, both zeroed entirely under reduced motion (an
// instant reveal, not a fast one). flipStepMs matches LoadoutCard's own
// Reanimated flip duration (450ms) so the next step never fires mid-flip.
export function flipStepMs(reducedMotion: boolean): number {
  return reducedMotion ? 0 : 450
}

export function slotStepMs(reducedMotion: boolean): number {
  return reducedMotion ? 0 : 220
}
