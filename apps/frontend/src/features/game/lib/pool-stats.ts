import type { Manga } from '@/shared/lib/zod'

// Pure, presentation-agnostic helpers behind §4's power-pool hardening pass
// (see ObsidianVault/game-lobby-todo.md §4 and game-lobby-frontend.md's UX
// passes for why this is banlist-only, no rarity/fruit-type whitelist).
// Deliberately a leaf module: no React, no component imports. `PoolItem` is
// a structural subset of `BannableItem` (fields/banlist-field.tsx) rather
// than importing that type, so this stays importable from both jest
// projects (lib→component would be a backwards dependency edge for a
// components/ file).
export type PoolItem = {
  id: string
  name: string
  kind: 'STAND' | 'DEVIL_FRUIT'
}

export type PoolCounts = {
  STAND: { total: number; remaining: number }
  DEVIL_FRUIT: { total: number; remaining: number }
}

// Per-kind total/remaining split over items, given a flat banned-id list.
// Set-based lookup so duplicate banned ids don't double-count and banned
// ids absent from the catalog (e.g. a stale ban left over after an admin
// deletes a power) are silently ignored. remaining never goes negative.
export function computePoolCounts(items: PoolItem[], banned: string[]): PoolCounts {
  const bannedSet = new Set(banned)
  const counts: PoolCounts = {
    STAND: { total: 0, remaining: 0 },
    DEVIL_FRUIT: { total: 0, remaining: 0 },
  }
  for (const item of items) {
    const bucket = counts[item.kind]
    bucket.total += 1
    if (!bannedSet.has(item.id)) {
      bucket.remaining += 1
    }
  }
  return counts
}

export type PoolShortfall = {
  manga: Manga
  kind: 'STAND' | 'DEVIL_FRUIT'
  remaining: number
  required: number
}

// Mirrors the backend rule in checkPoolSufficiency
// (apps/backend/.../game_service.go): JOJO's selected power manga needs
// enough Stands, ONE_PIECE's needs enough Devil Fruits, only for mangas
// actually selected in powerMangas. Deliberately stricter than the
// backend, which checks actual seated occupancy - the UI has no live
// roster to read here, so it uses the configured teamSize (capacity) as
// `required` instead; see power-pool-fields.tsx's warning banner for where
// that's surfaced. Returns [] whenever catalogKnown is false so a still-
// loading catalog (items.length === 0 while queries are in flight) never
// renders a false "insufficient pool" alarm.
export function poolShortfalls(
  counts: PoolCounts,
  powerMangas: Manga[],
  required: number,
  catalogKnown: boolean
): PoolShortfall[] {
  if (!catalogKnown) return []
  const shortfalls: PoolShortfall[] = []
  if (powerMangas.includes('JOJO') && counts.STAND.remaining < required) {
    shortfalls.push({ manga: 'JOJO', kind: 'STAND', remaining: counts.STAND.remaining, required })
  }
  if (powerMangas.includes('ONE_PIECE') && counts.DEVIL_FRUIT.remaining < required) {
    shortfalls.push({
      manga: 'ONE_PIECE',
      kind: 'DEVIL_FRUIT',
      remaining: counts.DEVIL_FRUIT.remaining,
      required,
    })
  }
  return shortfalls
}

export type PoolSearchResult = {
  results: PoolItem[]
  total: number
}

// NFD-normalize + lowercase, stripping combining diacritics - same
// approach as this codebase's admin search (see admin_search_filters
// vault note), reimplemented here since no shared helper existed yet.
function foldForSearch(value: string): string {
  return value
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
}

// Case/diacritic-insensitive multi-token AND search over items, excluding
// already-banned ids. Every whitespace-split token of the (trimmed) query
// must match somewhere in the folded item name, in any order. Returns up
// to `limit` results plus the true total match count so the UI can show
// "showing N of M". An empty/whitespace-only query returns nothing (no
// results, total 0) - browsing the full catalog isn't this field's job.
export function searchPoolItems(
  items: PoolItem[],
  banned: string[],
  query: string,
  limit = 8
): PoolSearchResult {
  const trimmed = query.trim()
  if (!trimmed) return { results: [], total: 0 }

  const bannedSet = new Set(banned)
  const tokens = foldForSearch(trimmed).split(/\s+/).filter(Boolean)

  const matches = items.filter((item) => {
    if (bannedSet.has(item.id)) return false
    const foldedName = foldForSearch(item.name)
    return tokens.every((token) => foldedName.includes(token))
  })

  return { results: matches.slice(0, limit), total: matches.length }
}
