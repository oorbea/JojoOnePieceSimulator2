import type { GameLoadout, GameRound, GameSnapshot } from '@/features/game/types/game.types'

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

export type LoadoutTrait = { key: string; i18nKey: string; value: string }

// loadoutTraits lists the six scalar loadout fields in a stable order.
// spin/hamon/fruitMastery are omitted entirely when NONE - a real "didn't
// draw this ability" state already communicated by the absent stand/
// devilFruit card, not worth a redundant chip. Haki/physicalForm have no
// NONE member (PRIVATE is their floor), so they're never filtered.
export function loadoutTraits(loadout: GameLoadout): LoadoutTrait[] {
  const traits: LoadoutTrait[] = []
  if (loadout.physicalForm) {
    traits.push({ key: 'physicalForm', i18nKey: 'game.match.trait.physicalForm', value: loadout.physicalForm })
  }
  traits.push({ key: 'armamentHaki', i18nKey: 'game.match.trait.armamentHaki', value: loadout.armamentHaki })
  traits.push({ key: 'observationHaki', i18nKey: 'game.match.trait.observationHaki', value: loadout.observationHaki })
  traits.push({ key: 'conquerorHaki', i18nKey: 'game.match.trait.conquerorHaki', value: loadout.conquerorHaki })
  if (loadout.spin !== 'NONE') {
    traits.push({ key: 'spin', i18nKey: 'game.match.trait.spin', value: loadout.spin })
  }
  if (loadout.hamon !== 'NONE') {
    traits.push({ key: 'hamon', i18nKey: 'game.match.trait.hamon', value: loadout.hamon })
  }
  if (loadout.fruitMastery !== 'NONE') {
    traits.push({ key: 'fruitMastery', i18nKey: 'game.match.trait.fruitMastery', value: loadout.fruitMastery })
  }
  return traits
}

// secondsUntil clamps to zero rather than going negative once a countdown
// has passed its closesAt - the server, not the client, is authoritative on
// when voting actually closes.
export function secondsUntil(closesAtMs: number | null, nowMs: number): number | null {
  if (closesAtMs === null) return null
  return Math.max(0, Math.round((closesAtMs - nowMs) / 1000))
}
