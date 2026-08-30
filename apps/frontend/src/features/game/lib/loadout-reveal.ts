import { hasAllLoadouts, revealSlotKinds, type LoadoutSlotKind } from '@/features/game/lib/match-rules'
import type { LiveMatchState } from '@/features/game/stores/game-socket.store'
import type { GameSnapshot } from '@/features/game/types/game.types'
import type { Manga, RevealSpeed } from '@/shared/lib/zod'

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

// Reveal timing constants (ms). Rewritten 2026-08-30 (owner request) to
// reproduce the pacing of the original terminal-based
// JoJoOnePiece_Simulator's loadingScreen/delay (github.com/oorbea/
// JoJoOnePiece_Simulator's main.cc/loadingScreen.cc) per participant per
// slot, instead of one flat per-slot hold for the whole lobby at once - see
// ObsidianVault/game-match-assignment-frontend.md for the redesign this
// supersedes. Mirrors the backend's reveal.go EXACTLY (same names,
// mechanically translated) so both sides compute the identical total
// duration without sharing code.
export const REVEAL_INTRO_MS = 1100
export const REVEAL_PLAYER_INTRO_MS = 1500
export const REVEAL_NARRATOR_MS = 1200
export const REVEAL_SPIN_BASE_MS = 1100
export const REVEAL_HOLD_SCALAR_MS = 2500
export const REVEAL_HOLD_FRUIT_MS = 5000
export const REVEAL_HOLD_STAND_MS = 10000
export const REVEAL_HOLD_EMPTY_MS = 2500
export const REVEAL_PLAYER_OUTRO_MS = 800
export const REVEAL_OUTRO_MS = 3300

// REVEAL_SPEED_MULTIPLIER mirrors enums.RevealSpeed's own Multiplier()
// method (reveal_speed.go) - every reveal timing constant above is scaled
// by this before use.
export const REVEAL_SPEED_MULTIPLIER: Record<RevealSpeed, number> = {
  NORMAL: 1.3,
  RELAXED: 1.6,
  SWIFT: 1.0,
}

export type RevealPhaseKind =
  | 'intro'
  | 'playerIntro'
  | 'narrator'
  | 'spin'
  | 'land'
  | 'playerOutro'
  | 'outro'

// A single tick of the sorteo timeline. `participant` indexes into the
// reveal's own player order (join order, i.e. snapshot.participants);
// `slot` indexes into revealSlotKinds(mangas). Both are undefined during
// the lobby-wide 'intro'/'outro' phases.
export type RevealPhase = {
  kind: RevealPhaseKind
  participant?: number
  slot?: number
}

// RevealPlayer is the minimal per-participant shape the timeline needs -
// mirrors the backend's game.RevealPlayer exactly, since a Stand/DevilFruit
// slot holds far longer when it actually lands a power than when it lands
// NONE (V1's delay(10)/delay(5) vs delay(2.5)).
export type RevealPlayer = { hasStand: boolean; hasDevilFruit: boolean }

// revealSpinCycles mirrors the backend's game.RevealSpinCycles bit for bit
// (FNV-1a 32, same byte layout) - deterministic, not actually random, so a
// client's spin-cycle count for a given (game, round, participant, slot)
// always agrees with what a fresh page load / another client would compute
// for the exact same reveal. Reproduces V1's generateRandomNumber(1, 2)
// flavour (1 or 2 loadingScreen cycles) without needing the server to ever
// send it. Keep both in sync.
export function revealSpinCycles(
  gameId: string,
  roundIndex: number,
  participantIndex: number,
  slot: number
): number {
  const bytes: number[] = []
  for (let i = 0; i < gameId.length; i++) {
    // gameId is always canonical ASCII (hex + hyphens), so charCodeAt is
    // already the UTF-8 byte value - no need for a full TextEncoder here.
    bytes.push(gameId.charCodeAt(i) & 0xff)
  }
  bytes.push(roundIndex & 0xff, participantIndex & 0xff, slot & 0xff)

  let hash = 0x811c9dc5 // FNV-1a 32 offset basis
  for (const b of bytes) {
    hash ^= b
    // 32-bit unsigned multiply by the FNV prime (16777619), Math.imul keeps
    // it within a 32-bit result the way Go's uint32 multiplication does.
    hash = Math.imul(hash, 0x01000193) >>> 0
  }
  return hash % 2 === 0 ? 1 : 2
}

// BLOCK_SLOTS are the two slots with their own art + stat grid (see
// slotHoldMs) - kept as a Set for the same O(1) membership check the
// backend's slotHoldMs switch expresses as a type switch.
const BLOCK_SLOTS = new Set<LoadoutSlotKind>(['stand', 'devilFruit'])

function slotHoldMs(slot: LoadoutSlotKind, player: RevealPlayer): number {
  if (slot === 'stand') return player.hasStand ? REVEAL_HOLD_STAND_MS : REVEAL_HOLD_EMPTY_MS
  if (slot === 'devilFruit') return player.hasDevilFruit ? REVEAL_HOLD_FRUIT_MS : REVEAL_HOLD_EMPTY_MS
  return REVEAL_HOLD_SCALAR_MS
}

// revealTimeline expands a lobby's reveal into the full jugador-por-jugador
// sequence: a lobby-wide intro, then for every player in turn a "Le toca a
// X" beat, then for every slot that player's turn plays a narrator line,
// a spin, and a hold on the landed value, then a short outro beat before
// the next player's turn - finally a lobby-wide outro. Mirrors the
// backend's game.RevealDuration structurally (see reveal.go) so both sides
// walk the identical sequence of beats, scaled by speed's own multiplier
// (REVEAL_SPEED_MULTIPLIER).
export function revealTimeline(
  gameId: string,
  roundIndex: number,
  mangas: Manga[],
  players: RevealPlayer[],
  speed: RevealSpeed
): { phase: RevealPhase; durationMs: number }[] {
  const slots = revealSlotKinds(mangas)
  const scale = REVEAL_SPEED_MULTIPLIER[speed] ?? REVEAL_SPEED_MULTIPLIER.NORMAL
  const scaled = (ms: number) => ms * scale

  const phases: { phase: RevealPhase; durationMs: number }[] = [
    { phase: { kind: 'intro' }, durationMs: scaled(REVEAL_INTRO_MS) },
  ]

  players.forEach((player, pi) => {
    phases.push({ phase: { kind: 'playerIntro', participant: pi }, durationMs: scaled(REVEAL_PLAYER_INTRO_MS) })
    slots.forEach((slot, si) => {
      phases.push({
        phase: { kind: 'narrator', participant: pi, slot: si },
        durationMs: scaled(REVEAL_NARRATOR_MS),
      })
      const cycles = revealSpinCycles(gameId, roundIndex, pi, si)
      phases.push({
        phase: { kind: 'spin', participant: pi, slot: si },
        durationMs: scaled(REVEAL_SPIN_BASE_MS * cycles),
      })
      phases.push({
        phase: { kind: 'land', participant: pi, slot: si },
        durationMs: scaled(slotHoldMs(slot, player)),
      })
    })
    phases.push({ phase: { kind: 'playerOutro', participant: pi }, durationMs: scaled(REVEAL_PLAYER_OUTRO_MS) })
  })

  phases.push({ phase: { kind: 'outro' }, durationMs: scaled(REVEAL_OUTRO_MS) })
  return phases
}

// revealDurationMs is the total sorteo duration for a reveal - the same
// number GameService.scheduleRevealDelay computes server-side (via
// revealDurationFor) to delay OpenVoting. Exposed so a client can compute
// it independently (e.g. for a reconnecting client with no revealMs from a
// LOADOUTS_ASSIGNED frame to trust).
export function revealDurationMs(
  gameId: string,
  roundIndex: number,
  mangas: Manga[],
  players: RevealPlayer[],
  speed: RevealSpeed
): number {
  return revealTimeline(gameId, roundIndex, mangas, players, speed).reduce(
    (sum, p) => sum + p.durationMs,
    0
  )
}
