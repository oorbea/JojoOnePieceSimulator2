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

// loadoutSlots lists every slot a loadout card should render, gated by
// which manga(s) the lobby actually plays - a One Piece-only lobby never
// gets a stand/spin/hamon slot (PRIVATE/NONE is just LoadoutBuilder's
// unreached default, not a real "you got nothing" result), and a JoJo-only
// lobby never gets physicalForm/devilFruit/fruitMastery/haki. Without this
// gate a single-manga lobby's card showed every One Piece stat pinned at
// its floor value (PRIVATE has no NONE member) alongside a real Stand,
// which reads as "the game only gave me a stand" instead of "physical form
// doesn't apply to this lobby". spin/hamon/fruitMastery are additionally
// omitted when NONE even inside a gated manga - a real "didn't draw this
// ability" result, already communicated by the stand/devilFruit block
// itself. Haki/physicalForm have no NONE member (PRIVATE is their floor),
// so within a One Piece lobby they're never filtered.
export function loadoutSlots(loadout: GameLoadout, mangas: Manga[]): LoadoutSlot[] {
  const jojo = mangas.includes('JOJO')
  const onePiece = mangas.includes('ONE_PIECE')
  const slots: LoadoutSlot[] = []

  for (const key of SLOT_ORDER) {
    switch (key) {
      case 'physicalForm':
        if (onePiece) slots.push({ key, i18nKey: 'game.match.trait.physicalForm', value: loadout.physicalForm })
        break
      case 'stand':
        if (jojo) slots.push({ key })
        break
      case 'devilFruit':
        if (onePiece) slots.push({ key })
        break
      case 'fruitMastery':
        if (onePiece && loadout.fruitMastery !== 'NONE') {
          slots.push({ key, i18nKey: 'game.match.trait.fruitMastery', value: loadout.fruitMastery })
        }
        break
      case 'hamon':
        if (jojo && loadout.hamon !== 'NONE') {
          slots.push({ key, i18nKey: 'game.match.trait.hamon', value: loadout.hamon })
        }
        break
      case 'armamentHaki':
        if (onePiece) slots.push({ key, i18nKey: 'game.match.trait.armamentHaki', value: loadout.armamentHaki })
        break
      case 'observationHaki':
        if (onePiece) slots.push({ key, i18nKey: 'game.match.trait.observationHaki', value: loadout.observationHaki })
        break
      case 'conquerorHaki':
        if (onePiece) slots.push({ key, i18nKey: 'game.match.trait.conquerorHaki', value: loadout.conquerorHaki })
        break
      case 'spin':
        if (jojo && loadout.spin !== 'NONE') {
          slots.push({ key, i18nKey: 'game.match.trait.spin', value: loadout.spin })
        }
        break
    }
  }

  return slots
}

// secondsUntil clamps to zero rather than going negative once a countdown
// has passed its closesAt - the server, not the client, is authoritative on
// when voting actually closes.
export function secondsUntil(closesAtMs: number | null, nowMs: number): number | null {
  if (closesAtMs === null) return null
  return Math.max(0, Math.round((closesAtMs - nowMs) / 1000))
}
