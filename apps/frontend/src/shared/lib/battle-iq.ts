import { MANUAL_RULES } from '@/shared/contracts/rules'

// BattleIQ is a JoJo-only Loadout stat: a raw 0-255 score, mirroring
// apps/backend/.../game/battle_iq.go. The backend sends only the number
// (see GameLoadoutResponse.BattleIQ); the WAIS-IV classification band is
// deliberately derived here instead, the same way the frontend already
// derives the haki-set summary from raw haki levels rather than the
// backend sending one.
//
// i18n note: the category labels live under `enums.battleIQCategory.*` in
// every locale, even though there is no backend enum behind them (every
// other `enums.*` namespace mirrors a real generated enum) - this is the
// one deliberate exception, recorded here so the keys already line up if
// the backend ever exposes the band itself.
export type BattleIQCategory = (typeof MANUAL_RULES.battleIQ)[number]['key']

// BATTLE_IQ_BANDS is the inclusive upper bound of every band except the
// last, in ascending order. Generated from the backend's own band table
// (contracts/rules.ts <- game.BattleIQ.Range), so it cannot drift from
// BattleIQ.Band()'s thresholds.
const BATTLE_IQ_BANDS: { max: number; key: BattleIQCategory }[] = MANUAL_RULES.battleIQ.map(
  (band, i, all) => ({ max: i === all.length - 1 ? Infinity : band.hi, key: band.key })
)

export function battleIQCategory(score: number): BattleIQCategory {
  for (const band of BATTLE_IQ_BANDS) {
    if (score <= band.max) return band.key
  }
  return 'VERY_SUPERIOR'
}

export function battleIQCategoryKey(score: number): string {
  return `enums.battleIQCategory.${battleIQCategory(score)}`
}

// formatBattleIQ returns "<score> · <category label>", or null when score
// is undefined (absent - no JoJo manga in the lobby). Callers must never
// substitute "0" for null: 0 is itself a legitimate, absurd score.
export function formatBattleIQ(
  t: (key: string) => string,
  score: number | undefined
): string | null {
  if (score === undefined) return null
  return `${score} · ${t(battleIQCategoryKey(score))}`
}
