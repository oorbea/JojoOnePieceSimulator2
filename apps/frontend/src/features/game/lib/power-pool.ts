import type { DevilFruitResponse } from '@/features/devil-fruits'
import type { StandResponse } from '@/features/stands'
import type { PoolFilter } from '@/features/game/types/game.types'

// applyPoolFilter mirrors the backend's game.PoolFilter.Apply exactly (see
// pool_filter.go's AllowsStand/AllowsDevilFruit): an empty rarity/fruitType
// allowlist means "no restriction" on that axis, `banned` is always an
// exclusion by id regardless of the allowlists, and order is preserved.
// Frontend-only re-implementation - the backend applies its own copy
// server-side when actually drawing loadouts (game_service.go's
// beginRound); this one exists purely so the sorteo's CS-strip shows the
// lobby's REAL candidate pool instead of the full catalogue (owner
// request, 2026-09-25 playtest feedback: the old reel span the whole
// catalogue, never the pool the host actually configured).
export function applyPoolFilter(
  stands: StandResponse[],
  fruits: DevilFruitResponse[],
  filter: PoolFilter
): { stands: StandResponse[]; fruits: DevilFruitResponse[] } {
  const banned = new Set(filter.banned)
  const standRarities = new Set(filter.standRarities)
  const fruitRarities = new Set(filter.fruitRarities)
  const fruitTypes = new Set(filter.fruitTypes)

  return {
    stands: stands.filter(
      (s) => !banned.has(s.id) && (standRarities.size === 0 || standRarities.has(s.rarity))
    ),
    fruits: fruits.filter(
      (f) =>
        !banned.has(f.id) &&
        (fruitRarities.size === 0 || fruitRarities.has(f.rarity)) &&
        (fruitTypes.size === 0 || fruitTypes.has(f.fruitType))
    ),
  }
}
