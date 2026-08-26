import type { LiveMatchState } from '@/features/game/stores/game-socket.store'
import type { GameLoadout, GameRound, GameSnapshot } from '@/features/game/types/game.types'
import type { Manga } from '@/shared/lib/zod'

// currentRound mirrors "the round a client should be looking at right now" -
// the backend never sends a round index explicitly for this, it's simply
// the last element of Rounds (see dto.NewGameStateResponse - rounds is
// append-only per Round.Index).
export function currentRound(snapshot: GameSnapshot): GameRound | null {
  const rounds = snapshot.rounds
  return rounds.length > 0 ? rounds[rounds.length - 1] : null
}

// hasAllLoadouts is true once every participant who could plausibly still be
// in the match has a loadout assigned. Bots always count (never
// disconnected); a disconnected human is excluded since the backend may
// still be waiting on their reconnect/replacement without ever handing them
// a loadout.
export function hasAllLoadouts(snapshot: GameSnapshot): boolean {
  return snapshot.participants
    .filter((p) => p.connected !== false)
    .every((p) => !!p.loadout)
}

// LoadoutSlotKind is every distinct thing a loadout card can reveal, in the
// exact order the owner asked for (2026-08-14): PhysicalForm -> Stand ->
// DevilFruit -> FruitMastery -> Hamon -> ArmamentHaki -> ObservationHaki ->
// ConquerorHaki -> Spin - the same order LoadoutBuilder.Build draws them in
// (apps/backend/.../loadout_builder.go), so the reveal walks the draw order
// a player actually experienced. 'stand'/'devilFruit' are the two big art
// blocks (rendered specially by LoadoutCard); every other slot is a scalar
// chip.
export type LoadoutSlotKind =
  | 'physicalForm'
  | 'stand'
  | 'devilFruit'
  | 'fruitMastery'
  | 'hamon'
  | 'armamentHaki'
  | 'observationHaki'
  | 'conquerorHaki'
  | 'spin'

export type LoadoutSlot = { key: LoadoutSlotKind; i18nKey?: string; value?: string }

const SLOT_ORDER: LoadoutSlotKind[] = [
  'physicalForm',
  'stand',
  'devilFruit',
  'fruitMastery',
  'hamon',
  'armamentHaki',
  'observationHaki',
  'conquerorHaki',
  'spin',
]

function hasSlotManga(key: LoadoutSlotKind, jojo: boolean, onePiece: boolean): boolean {
  switch (key) {
    case 'stand':
    case 'hamon':
    case 'spin':
      return jojo
    default:
      return onePiece
  }
}

// revealSlotKinds lists which slot KINDS a reveal shows for a lobby playing
// mangas, in draw order - mirrors the backend's game.RevealSlots exactly
// (apps/backend/.../reveal.go), gated the same way loadoutSlots is (see its
// doc). Pure function of mangas alone, independent of any actual loadout -
// this is what lets the reveal overlay (which shows every participant's
// slot at once, before any card exists) and the backend's own delay timer
// agree on how many slots there are without exchanging anything beyond
// mangas itself. Kept in sync with SLOT_ORDER/hasSlotManga below.
export function revealSlotKinds(mangas: Manga[]): LoadoutSlotKind[] {
  const jojo = mangas.includes('JOJO')
  const onePiece = mangas.includes('ONE_PIECE')
  return SLOT_ORDER.filter((key) => hasSlotManga(key, jojo, onePiece))
}

// loadoutSlots lists every slot a loadout card should render, gated by
// which manga(s) the lobby actually plays - a One Piece-only lobby never
// gets a stand/spin/hamon slot (PRIVATE/NONE is just LoadoutBuilder's
// unreached default, not a real "you got nothing" result), and a JoJo-only
// lobby never gets physicalForm/devilFruit/fruitMastery/haki. Without this
// gate a single-manga lobby's card showed every One Piece stat pinned at
// its floor value (PRIVATE has no NONE member) alongside a real Stand,
// which reads as "the game only gave me a stand" instead of "physical form
// doesn't apply to this lobby". Unlike an earlier pass, a NONE spin/hamon/
// fruitMastery is still INCLUDED (with its NONE value) rather than omitted -
// that's what makes a gated manga's slot COUNT depend only on mangas, never
// on the actual draw (see revealSlotKinds' doc for why that matters), and a
// "you didn't draw this one" is itself worth revealing during the sorteo.
// TraitChip renders the NONE case as a dimmed chip.
export function loadoutSlots(loadout: GameLoadout, mangas: Manga[]): LoadoutSlot[] {
  const jojo = mangas.includes('JOJO')
  const onePiece = mangas.includes('ONE_PIECE')
  const slots: LoadoutSlot[] = []

  for (const key of SLOT_ORDER) {
    if (!hasSlotManga(key, jojo, onePiece)) continue
    switch (key) {
      case 'physicalForm':
        slots.push({ key, i18nKey: 'game.match.trait.physicalForm', value: loadout.physicalForm })
        break
      case 'stand':
        slots.push({ key })
        break
      case 'devilFruit':
        slots.push({ key })
        break
      case 'fruitMastery':
        slots.push({ key, i18nKey: 'game.match.trait.fruitMastery', value: loadout.fruitMastery })
        break
      case 'hamon':
        slots.push({ key, i18nKey: 'game.match.trait.hamon', value: loadout.hamon })
        break
      case 'armamentHaki':
        slots.push({ key, i18nKey: 'game.match.trait.armamentHaki', value: loadout.armamentHaki })
        break
      case 'observationHaki':
        slots.push({ key, i18nKey: 'game.match.trait.observationHaki', value: loadout.observationHaki })
        break
      case 'conquerorHaki':
        slots.push({ key, i18nKey: 'game.match.trait.conquerorHaki', value: loadout.conquerorHaki })
        break
      case 'spin':
        slots.push({ key, i18nKey: 'game.match.trait.spin', value: loadout.spin })
        break
    }
  }

  return slots
}

// voteProgress reports how many connected humans have voted in the current
// round versus how many are eligible to - the frontend's own definition of
// exactly what the backend's Game.humanVoteProgress computes, so a
// reconnecting client with no live frame yet renders the same numbers a
// connected one already has. `total` prefers live.voters (from the latest
// VOTE_CAST, absolute per LiveMatchState's doc); when that's not known yet
// (no frame has landed for this round - true right after a reconnect, or in
// the instant between VOTING_OPENED and the first VOTE_CAST) it falls back
// to counting connected, non-bot participants off the snapshot directly.
// `cast` follows the same rule via live.votesCast, falling back to
// round.votedParticipantIds intersected with that same connected-human set
// (bots are never in votedParticipantIds's *meaning*, but the intersection
// guards it anyway in case a bot's id ever appeared there).
export function voteProgress(snapshot: GameSnapshot, live: LiveMatchState): { cast: number; total: number } {
  const round = currentRound(snapshot)
  const connectedHumanIds = new Set(
    snapshot.participants.filter((p) => p.kind !== 'BOT' && p.connected !== false).map((p) => p.id)
  )
  const frameIsForCurrentRound = round !== null && live.votingRoundIndex === round.index

  const total = frameIsForCurrentRound && live.voters !== null ? live.voters : connectedHumanIds.size

  if (frameIsForCurrentRound && live.votesCast !== null) {
    return { cast: live.votesCast, total }
  }
  const cast = round ? round.votedParticipantIds.filter((id) => connectedHumanIds.has(id)).length : 0
  return { cast, total }
}

// secondsUntil clamps to zero rather than going negative once a countdown
// has passed its closesAt - the server, not the client, is authoritative on
// when voting actually closes.
export function secondsUntil(closesAtMs: number | null, nowMs: number): number | null {
  if (closesAtMs === null) return null
  return Math.max(0, Math.round((closesAtMs - nowMs) / 1000))
}
