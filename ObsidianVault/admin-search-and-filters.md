---
title: "Search + filters for Stands, Devil Fruits, Stages (2026-08-11)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - feature
  - admin
---

# Server-side search + filter bars (2026-08-11)

Stands and Devil Fruits admin screens had zero search/filter UI, even though
`StandFilters`/`DevilFruitFilters` (rarity, stats, fruitType, evolvesFrom)
were fully wired end to end (types, api, keys, backend parsing/SQL) and
simply never populated by any container. Stages had just shipped a search
bar the previous tanda, but it filtered client-side in memory.

## Status

Done. Backend + frontend, all three entities, tests green (`go test ./...`,
`-tags vips` and `-tags integration` in Docker, `pnpm test` both jest
projects), swagger regenerated, i18n parity verified.

## Decision: search moved server-side, Stages included

The owner's ask started as "add search+filters to Stands/DevilFruits like
Stages has", but since Stages' search was backend-less (`?q=` didn't exist
anywhere), the owner explicitly redirected mid-tanda: **add `?q=` to the
backend and make Stages use it too**, rather than keep two different search
strategies (client vs server) across three near-identical screens.

## Backend: `Search *string` joins the existing filter structs

`ports.StandFilters`/`DevilFruitFilters` each gained a `Search *string`
field - not a separate parameter, so it rides the existing `hasFilters`
flag and `Filter(...)` path (and the Redis decorator) for free.

**Stage was refactored to match the powers pattern.** It used to parse
`?manga=` inline in the handler and call a dedicated `ListByManga` repo
method - fine for one filter, but doesn't compose with a second. New
`ports.StageFilters{Manga, Search}` + `dto.StageFiltersFromQuery` +
`db/query/stages.sql`'s `FilterStageRows` (replacing `ListStagesByManga`),
same `sqlc.narg(...) IS NULL OR ...` shape as
`FilterStandRows`/`FilterDevilFruitRows`. `StageService.StagesByManga` became
`FilterStages`; `IStageRepository.ListByManga` became `Filter`.

**Two traps worth remembering for the next filter added to any of these:**

1. **Cache keys are hand-serialized, not derived from the struct.**
   `standFilterKey`/`devilFruitFilterKey` (`internal/infrastructure/cache/keys.go`)
   build a canonical string field-by-field in a fixed order. Add a field to
   `ports.XFilters` and forget to add it here, and two different searches
   share one Redis slot - stale/wrong results with no error anywhere. Caught
   this in review, not from a test failing on its own; added
   `TestXRepository_Filter_SearchDifferentiatesCacheKey` for both Stand and
   DevilFruit specifically to pin this down going forward.
2. **`FilterStandRows`'s `WHERE` lives inside a recursive CTE (`base`), not
   the outer `SELECT`.** Stand's evolvesFrom-ancestor-chain query means the
   translations LATERAL join for the outer `SELECT` doesn't exist yet at
   filter time - had to add a second, separate `LEFT JOIN LATERAL` inside
   `base` just to filter on `tr.description`, and must NOT add that join's
   columns to `base`'s `SELECT` list, since the recursive `chain` CTE does
   `SELECT * FROM base UNION SELECT p2...` and any extra column breaks that
   union's column correspondence. `FilterDevilFruitRows`/`FilterStageRows`
   have no such recursion, so their `Search` addition was a one-line `AND`.

**ILIKE input is escaped in Go, not SQL**: `escapeLikePattern` (new helper in
`internal/infrastructure/repositories/power_translations.go`, shared by all
three repos via `searchPtr`) escapes `\`, `%`, `_` before the term reaches
`'%' || term || '%' ESCAPE '\'`. Without this a search for `%` returns the
whole table and `_` acts as a single-char wildcard - both would look like
"search is broken" bugs to an admin, not obviously security-shaped ones.

## Frontend

- **`shared/hooks/use-debounced-value.ts`** (new): generic `useDebouncedValue<T>`,
  300ms default. Search-as-you-type would otherwise fire one request per
  keystroke now that search is server-side.
- **`shared/components/presentational/filter-disclosure.tsx`** (new): generic
  collapsible "more filters" panel (badge count + clear-all), used only by
  Stands today (6 stats + evolvesFrom would otherwise permanently occupy
  screen space) but lives in `shared/` since Devil Fruits/Stages could grow
  into it later.
- **Stands**: search + rarity always visible; 6 stats + evolvesFrom behind
  `FilterDisclosure`. **Bug this would have introduced, caught before
  shipping**: the "Evolves From" filter and the create/edit form's own
  "Evolves From" picker used to derive their options from the same `stands`
  list the grid renders. Once that list became filterable, applying any
  filter would silently narrow the form's parent-Stand picker too - a
  Stand's real evolution parent could disappear from the dropdown just
  because an admin was mid-search for something else. Fixed by adding a
  second, unfiltered `useStands()` call (`allStands`) that feeds both the
  filter's own option list and the form's picker; the filtered call only
  feeds the grid. Same reasoning applies to the "picture failed" toast
  watcher - moved from the filtered list to `allStands`, or a Stand whose
  picture fails while hidden by a filter would never surface the toast.
- **Devil Fruits**: search + rarity + fruitType, all always visible (3
  controls, no disclosure needed).
- **Stages**: dropped its `useMemo` client-side name/description filter
  entirely; `q` now joins `manga` in the same server-side filters object.
- Empty-state copy: added `emptyFilteredTitle` next to `emptyTitle` for all
  three entities (was already the pattern Stages introduced, just doing it
  right this time - `stands`/`devilFruits` didn't have the split before,
  and Stages' `emptyTitle` had been overwritten with filtered wording last
  tanda since it had no unfiltered state to distinguish from).

## Tests

Backend: per-entity endpoint tests for `?q=` (name match, description
match, no match; Stage's `?manga=&q=` combined - proves AND not OR), a
cache-key differentiation test per entity, and integration tests for ILIKE
case-insensitivity + locale-resolved description + `%`/`_` literal
handling.

Frontend: `filter-disclosure.test.tsx`, `use-debounced-value.test.tsx`
(rendered through a tiny `Text` probe component + fake timers - this
codebase's `renderHook` from `@testing-library/react-native` resolved to a
broken duplicate install (multiple pnpm-hoisted copies), so probe-component
+ `render()`/`act()` was used instead, matching every other test in the
suite anyway), and `stands-screen.test.tsx`/`devil-fruits-screen.test.tsx`
covering the search field, the FilterDisclosure's collapse, and both empty
states.

## Known follow-up, not done this tanda

~~None new. The Redis-cache-decorator gap for `stage_repository.go` (flagged
in [[stages_admin_crud_2026-08-11]]) still stands - `FilterStageRows` is a
live Postgres round trip on every admin request, same as before.~~

> [!done] Closed 2026-09-02
> `cache.StageRepository` now decorates the Stage adapter read-through:
> `Filter`/`List`/`FindByID` are cached per locale, `Stages` per manga, and
> every write flushes the whole `stages` namespace. Details and design
> decisions in [[stage-redis-cache-2026-09-02]].

Related: [[game-stage-content]], [[admin-panel-crud-ux-fixes]],
[[picture-events-sse]], [[stage-redis-cache-2026-09-02]].
