---
title: "Keyset pagination for the catalogue (2026-09-08)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - performance
---

# Catalogue pagination (2026-09-08)

Part of T3 in [[entrega-imagenes-red-lenta-2026-09-07]]. Shipped for
**Stand, DevilFruit and Stage** - all three now support the same opt-in
`?limit=`/`?cursor=` contract described below.

## Cursor keyset, not offset

`OFFSET` re-scans and skips/duplicates rows when an admin renames or creates
an item mid-pagination - not hypothetical for an admin-editable catalogue,
and renaming changes sort position because `powers.name` **is** the order
key. Keyset is free here: every list query already ends `ORDER BY p.name`,
and `powers.name` is `UNIQUE`, so there's no tie-break column to add.

## The `LIMIT`-inside-`base` bug this design prevents

`ListStandRows` has no recursive ancestor CTE (fine only because the full
catalogue is always in the result, so `buildStandsCollect`'s `byID` map
always has every ancestor). `FilterStandRows` does have the CTE. Naively
adding `LIMIT`/cursor to the **outer** `SELECT` of a filtered+paginated query
would truncate `matched = false` ancestor rows that a later page's
`evolvesFrom` needs - and `buildStandsLenient` **silently drops** any
matched stand whose ancestor isn't in the loaded set, just logging. A user
would see 23 of 24 cards, forever, with nothing in any error path.

Fix: the cursor predicate and `LIMIT` sit **inside** the `base` CTE, before
`chain`'s recursion runs. The recursion then always finds ancestors
regardless of where they'd sort - `PageStandRows`. Verified two ways:
1. Manually against real Postgres (`Seed Beta` evolving from `Seed Zeta`,
   where the ancestor name sorts *after* the descendant) - `page_limit=1`
   still returned both rows, `Seed Zeta` with `matched=false`.
2. `TestListStands_Page_WalkingCursorToExhaustion_EqualsUnpaginatedList` in
   `stand_endpoints_test.go` - walks the cursor to exhaustion, asserts every
   unpaginated item appears exactly once, and asserts `evolvesFrom` is
   populated at every page boundary.

`Go` requests `limit+1` rows and detects `HasMore` from the extra row - no
separate `COUNT` needed for that; `total` (page 1 only) is a real `COUNT(*)`
via `CountStandRows`, opt-out via `?total=false`.

## Envelope, and why it's opt-in

Generics don't survive `cmd/typegen` (`reflect.Type.Name()` on
`Page[StandResponse]` isn't a valid TS identifier, and Go type aliases don't
get their own reflect name) - so the envelope is a concrete
`StandPageResponse{PageInfo; Items []StandResponse}` per resource, exploiting
that `structFields` already flattens embedded structs the way
`encoding/json` does (precedent: `LobbyPreviewResponse`). Wire shape:
`{items, nextCursor?, total?}` - no nested `pageInfo` object.

**A request with neither `?limit=` nor `?cursor=` gets the legacy bare
array, byte-identical to before.** Not optional: a stale already-loaded tab
running old JS against a new backend would do `response.data.some(...)` on
an object instead of an array and go blank - the exact failure class this
repo already burned itself on twice (`admin-crud-cache-stale-sw.md`,
`etag-304-body-loss.md`). Locked with
`TestListStands_NoPageParams_StaysBareArray`.

## Cursor shape and the fingerprint

Opaque base64url of `{"v":1,"f":"<8-byte fingerprint>","k":{"name":"..."}}`.
The fingerprint is `sha256(canonical filter string + "|" + locale)[:8]`,
bound at issue time - replaying a cursor against a different filter
combination is a 400 (`DecodeCursor`'s fingerprint check), not
silently-wrong results. `TestListStands_Page_CursorFromDifferentFilters_Returns400`
locks this.

**Anti-drift fix, closed 2026-09-08:** `StandFilters`/`DevilFruitFilters`/
`StageFilters` each gained a `Canonical() string` method in `ports` -
the single source of truth for a filter set's canonical rendering,
consumed by both `cache/keys.go`'s `*FilterKey` functions and
`dto/pagination.go`'s `*FiltersFingerprint` functions. Before this, each
call site duplicated its own copy of the field list, exactly the trap
`admin-search-and-filters.md` already documents - a field added to a
*Filters struct without updating both was a silent gap (worst case: two
different filter combinations hashing to the same cache entry or cursor
fingerprint). `ports/canonical_test.go` locks it with a reflect-based
field-count assertion (`Canonical()`'s `|`-segment count must equal the
struct's `NumField()`) plus a per-field "changing this field changes the
output" check - so a field added without extending `Canonical()` fails
loudly in `go test` instead of compiling silently wrong.

## Caching

`Page`/`Count` are deliberate pass-throughs in
`cache/stand_repository.go` - never cached. Filters × locales × pages is a
combinatorial spray of entries each read once before the next admin write
invalidates the whole `stands:v2` namespace, and a paginated read is already
a cheap indexed `LIMIT` over a few hundred rows - the expensive read the
cache exists for (`all:<locale>`, a full un-paginated fetch) is exactly what
pagination replaces. ETag/304 needs zero code changes: `cacheHeaders`
already hashes the body per-request, and `limit`/`cursor` already fragment
the axios ETag cache key.

## Deviation from the plan: no new env vars

The plan called for `CATALOGUE_PAGE_SIZE`/`CATALOGUE_MAX_PAGE_SIZE`. Shipped
instead as Go constants (`defaultPageSize = 24`, `maxPageSize = 100`) in
`dto/pagination.go` - no operational reason for these to differ per deploy,
and every env var is a three-place manual step for the owner
(`.env`/`.env.example`/GitHub Variables) per
[[entrega-imagenes-red-lenta-2026-09-07]]'s standing constraint. Revisit if
a real need to tune page size per environment ever comes up.

## DevilFruit and Stage (shipped 2026-09-08)

`PageDevilFruitRows`/`CountDevilFruitRows` are a plain LIMIT/cursor over the
same filtered WHERE clause `FilterDevilFruitRows` uses - no recursive CTE,
no ancestor-truncation risk (DevilFruit has no `evolves_from`), so the
T3.6 bug class simply doesn't apply here.

`PageStageRows`/`CountStageRows` sort by the triple
`(manga, position, name)` - `UNIQUE (manga, name)` plus a fixed manga makes
the triple unique overall, since `position` alone can tie within a manga
(no `UNIQUE(manga, position)`, by design - admins can reorder). The cursor
predicate is a row-value comparison:

```sql
AND (
  sqlc.narg('after_manga')::manga IS NULL
  OR (s.manga, s.position, s.name) > (sqlc.narg('after_manga')::manga, sqlc.narg('after_position')::int, sqlc.narg('after_name')::text)
)
```

**The `::manga` cast trap, avoided by *not* casting anything.** `manga` is
left as `::manga` (its native enum type) on both sides of the row
comparison, never cast to `::text`. Postgres then compares it using its own
default btree opclass, which sorts by enum declaration order
(`CREATE TYPE manga AS ENUM ('JOJO', 'ONE_PIECE')` → JOJO before ONE_PIECE)
- identical to what a bare `ORDER BY s.manga, s.position, s.name` does. A
`::text` cast anywhere in the cursor predicate would silently switch to
alphabetical order and desync the cursor from the actual `ORDER BY`,
corrupting page boundaries with no error anywhere. Locked with
`TestListStages_Page_WalkingCursorToExhaustion_EqualsUnpaginatedList`, which
seeds stages sharing a manga with different `order` values (including a tie
broken by name) and asserts the paged walk lands in the exact declaration
order.

`ports.StagePageCursor{Manga, Position, Name}` carries all three fields
together (never partially) through the service/repository layers; the wire
cursor's decoded `k` is `dto.StageCursor{manga, position, name}` (JSON, not
the Go enum type - decoded back via `enums.ParseManga` before it reaches the
repository).

`DevilFruitPageResponse`/`StagePageResponse` follow `StandPageResponse`'s
shape exactly (`PageInfo` embedded, `items` array) - both registered in
`cmd/typegen/registry.go`'s `restTypes`, same as Stand.

## Frontend consumption (shipped 2026-09-08, Stand only)

`src/shared/hooks/use-paginated-catalogue.ts` is a generic `useInfiniteQuery`
wrapper: `usePaginatedCatalogue(queryKey, fetchPage, {hasPendingPicture})`
returns `{items, total, hasNextPage, isFetchingNextPage,
isFetchNextPageError, fetchNextPage, ...}` - `items` is every loaded page
flattened, `fetchPage(cursor, limit)` is the only per-resource piece. Keeps
the same native-only polling fallback (web uses `PictureEventsBridge`'s SSE
push instead) the three unpaginated hooks already have, generalized via the
`hasPendingPicture` predicate instead of a hardcoded `pictureStatus ===
'PENDING'` check.

**Query key must differ from the unpaginated one, even for identical
filters.** An infinite query caches `{pages, pageParams}`; a plain
`useQuery` caches a bare array. `standKeys.page(filters)` exists
specifically so the admin screen's full `useStands()` fetch and the public
catalogue's paginated fetch never collide under the same key when both
happen to be unfiltered - both still hang off `standKeys.all()`, so a
mutation's existing `allLocales` invalidation covers both without change.

**UI**: `stands-screen.tsx` gained an explicit "Cargar más" `GlossButton`
below the grid (not infinite scroll - `norma-teclado.md` requires it be
keyboard-reachable), all optional props (`hasNextPage`/`onLoadMore`/etc) so
the admin screen's usage (still full-fetch, no pagination) renders nothing
extra. States: has-more → button + `N / total`; loading → disabled +
spinner; a failed page → button becomes "Retry" (`tone="orange"`), already-
loaded pages stay on screen; exhausted → button gone, "That's all (total)".

**Focus management**: `StandCard`/`DevilFruitCard`/`StageCard` all gained
`forwardRef` targeting their main (detail) Pressable; each screen tracks
refs by id and moves focus to the first newly-appended card after a
successful load, so a keyboard user never lands on a button that moved or
unmounted. Cross-platform via `focusElement` (`shared/lib/a11y.ts`, closed
2026-09-08): web calls `.focus()` on the forwarded DOM node (RNW's
Pressable ref), native calls `AccessibilityInfo.setAccessibilityFocus` via
`findNodeHandle` - RN's `View` has no DOM-style `.focus()`, so the two
platforms need genuinely different calls, not just a guarded no-op. `null`/
unresolvable handles are silently skipped (best-effort, never on a user's
direct action). `shared/lib/__tests__/a11y.test.ts` covers the web branch
only - jest.mock('react-native', ...) hit a module-instance mismatch when
forcing the native branch from jsdom (findNodeHandle got mocked,
AccessibilityInfo.setAccessibilityFocus didn't, for reasons not worth the
time to chase for a 15-line helper); the native branch is simple enough to
read for correctness and gets the same manual on-device verification every
other `Platform.OS !== 'web'` branch in this codebase does.

Adopted first in the public Stand catalogue, then rolled out to
DevilFruit/Stage the same session (2026-09-08) - `DevilFruitCard`/`StageCard`
both gained the same `forwardRef`, `catalog-devil-fruits-container.tsx`/
`catalog-stages-container.tsx` both wired to
`getDevilFruitsPage`/`getStagesPage` + `usePaginatedCatalogue`. Admin
screens (all three) are unchanged - still the full unpaginated fetch, by
design (see T1.4's note on why the admin container needs the full set).

**Two real bugs this rollout caught, both fixed in the same commit:**
- `stages-screen.tsx` imports the Lucide `Map` icon
  (`@tamagui/lucide-icons-2`), which shadowed the global `Map` constructor -
  `new Map<string, View | null>()` for the card-ref tracking was silently
  constructing the icon component instead, so every `.set(...)` call threw
  `cardRefs.current.set is not a function`. Caught immediately by the new
  "Cargar más" tests (only 2 of 4 failed - the ones that actually appended a
  page and exercised the focus effect). Fixed by aliasing the icon import to
  `MapIcon`.
- `use-paginated-catalogue.ts`'s `queryKey` didn't include `limit` -
  `@tanstack/eslint-plugin-query`'s `exhaustive-deps` rule flagged it as a
  real bug (the queryFn closes over `limit`, but a different page size for
  the same filters needs its own cache entry, or a caller changing `limit`
  mid-session would silently keep serving the old size from cache). Also
  switched the hook's return from `{...query, items, total}` to explicit
  fields - spreading a TanStack Query result subscribes the caller to every
  internal field's changes (`no-rest-destructuring`).

## Still open

- Native accessibility focus after "Cargar más" (see above) - currently a
  no-op on native, not a crash, but not the real fix either.
