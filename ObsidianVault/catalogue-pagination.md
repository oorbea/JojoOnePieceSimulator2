---
title: "Keyset pagination for the catalogue (2026-09-08)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - performance
  - in-progress
---

# Catalogue pagination (2026-09-08)

Part of T3 in [[entrega-imagenes-red-lenta-2026-09-07]]. Shipped for **Stand
only** so far - DevilFruit/Stage still return the full unpaginated list.

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

**Known duplication, deliberate for now:** `standFiltersCanonical` in
`dto/pagination.go` renders `ports.StandFilters` in the same fixed field
order `cache/keys.go`'s `standFilterKey` already does, but as a second,
separately-maintained copy - the anti-drift fix the original plan called
for (`Canonical()` on `StandFilters` itself, consumed by both) hasn't
landed. A field added to `StandFilters` without updating both places doesn't
break correctness (worst case: a wrong-but-consistent fingerprint, still
caught by DecodeCursor's equality check) but does mean two filter
combinations could theoretically collide to the same fingerprint. Follow-up.

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

## Still open

- DevilFruit/Stage pagination (same shape, `PageDevilFruitRows`/
  `PageStageRows` - Stage needs the `::manga` cast trap the original plan
  called out, since `ORDER BY s.manga` sorts by enum declaration order but
  `s.manga::text` sorts alphabetically, and the cursor comparison must match
  whichever `ORDER BY` actually uses).
- Frontend: no `use-paginated-catalogue.ts` hook, no "Cargar más" UI yet -
  `GET /stands` is paginatable but nothing calls it that way.
- The `Canonical()` anti-drift refactor mentioned above.
