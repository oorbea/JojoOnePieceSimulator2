import { hasAllLoadouts, revealSlotKinds } from '@/features/game/lib/match-rules'
import type { LiveMatchState } from '@/features/game/stores/game-socket.store'
import type { GameSnapshot } from '@/features/game/types/game.types'
import type { Manga } from '@/shared/lib/zod'

// shouldReveal gates the sorteo on "an assignment frame arrived since the
// last reveal" AND "the snapshot has actually caught up" AND "the game is
// still ASSIGNING" - LOADOUTS_ASSIGNED is sent before its own
// pushCurrentState, so the frame alone is not proof the snapshot carries the
// new loadouts yet (see game-socket.store.ts's LOADOUTS_ASSIGNED case).
//
// The ASSIGNING check replaced an earlier "does the current round's index
// match what was assigned" check (fixed 2026-08-14, same pass as the sorteo
// redesign - see game-match-assignment-frontend.md). That check assumed a
// Round already existed the moment loadouts were assigned - true when
// OpenVoting ran in the same synchronous call as AssignLoadouts, which is
// exactly what changed: GameService.scheduleRevealDelay now holds the Game
// in ASSIGNING, with ZERO Rounds yet, for the whole sorteo. Gating on the
// round index made shouldReveal false for that entire window and only flip
// true once OpenVoting finally created the round - i.e. exactly when voting
// had already opened, the one moment the sorteo must NOT still be starting.
// snapshot.state is authoritative and needs no Round to exist.
export function shouldReveal(live: LiveMatchState, snapshot: GameSnapshot): boolean {
  if (live.assignmentSeq <= live.revealedAssignmentSeq) return false
  if (snapshot.state !== 'ASSIGNING') return false
  return hasAllLoadouts(snapshot)
}

// Reveal timing constants (ms), deliberately fixed regardless of which
// value a slot's draw actually landed on - even a NONE/PRIVATE floor value
// still plays out its full spin+hold. Mirrors the backend's reveal.go
// EXACTLY (RevealIntroMs/RevealSpinMs/RevealHoldScalarMs/RevealHoldBlockMs/
// RevealOutroMs) so both sides compute the identical total duration without
// sharing code - loosely modeled on the original terminal-based
// JoJoOnePiece_Simulator's loadingScreen/delay pacing (github.com/oorbea/
// JoJoOnePiece_Simulator).
export const REVEAL_INTRO_MS = 1100
export const REVEAL_SPIN_MS = 1650
export const REVEAL_HOLD_SCALAR_MS = 2500
export const REVEAL_HOLD_BLOCK_MS = 4000
export const REVEAL_OUTRO_MS = 3300

const BLOCK_SLOTS = new Set(['stand', 'devilFruit'])

export type RevealPhaseKind = 'intro' | 'spin' | 'land' | 'outro'

// A single tick of the sorteo timeline: 'intro' (nothing spinning yet,
// "preparados"), 'spin' (every participant's carril spins for slot N),
// 'land' (slot N's real value is visible and holding, before the next
// slot's spin starts), 'outro' (dissolve into the filled-in roster). Unlike
// the old per-participant reveal, this is poder-a-poder for EVERY
// participant at once - slot N always means the same slot for the whole
// lobby, so one global timeline (not one per participant) drives the
// overlay.
export type RevealPhase = { kind: RevealPhaseKind; slot?: number }

// revealPhases expands a lobby's mangas into the full sorteo timeline, each
// phase paired with how long it holds. A pure function of mangas alone
// (never of any actual loadout/draw), matching game.RevealDuration's own
// doc on the backend for why that's the point: both sides can compute the
// identical total without exchanging anything beyond mangas.
export function revealPhases(mangas: Manga[]): { phase: RevealPhase; durationMs: number }[] {
  const kinds = revealSlotKinds(mangas)
  const phases: { phase: RevealPhase; durationMs: number }[] = [
    { phase: { kind: 'intro' }, durationMs: REVEAL_INTRO_MS },
  ]
  kinds.forEach((kind, index) => {
    phases.push({ phase: { kind: 'spin', slot: index }, durationMs: REVEAL_SPIN_MS })
    phases.push({
      phase: { kind: 'land', slot: index },
      durationMs: BLOCK_SLOTS.has(kind) ? REVEAL_HOLD_BLOCK_MS : REVEAL_HOLD_SCALAR_MS,
    })
  })
  phases.push({ phase: { kind: 'outro' }, durationMs: REVEAL_OUTRO_MS })
  return phases
}

// revealDurationMs is the total sorteo duration for mangas - the same
// number GameService.scheduleRevealDelay uses server-side to delay
// OpenVoting (see game.RevealDuration). Exposed so a client can compute it
// independently (e.g. for a reconnecting client with no revealMs from a
// LOADOUTS_ASSIGNED frame to trust).
export function revealDurationMs(mangas: Manga[]): number {
  return revealPhases(mangas).reduce((sum, p) => sum + p.durationMs, 0)
}
