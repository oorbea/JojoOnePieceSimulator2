---
title: "Bug: frontend lost the body on every 304, showing empty Stand/Devil Fruit lists"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - bug
  - fixed
---

# Bug: conditional GET dropped the response body on a 304 (fixed 2026-08-09)

## Symptom

Creating a Stand or Devil Fruit in the admin channels appeared to succeed (toast fired), but the
list screen kept rendering its empty state ("no Stands yet") as if nothing had been created.
Reported alongside an unrelated visual bug: the "add skill" `GlossButton` in `SkillsField`
rendered as a tiny dot instead of a `+` — see the `GlossButton` fix note in
[[frontend-responsive-frutiger-aero]].

## Root cause

`apps/frontend/src/shared/api/etag.ts` only ever remembered the **ETag** for a URL
(`getKnownEtag`/`rememberEtag`), never the **body** it was issued for. The request interceptor
(`interceptors.ts`) attached `If-None-Match` from that map on every GET; the backend correctly
answered `304 Not Modified` with no body (`cache_headers.go`, see [[backend-contract]] Caching
section). `apps/frontend/src/shared/api/client.ts` treats `304` as a success status
(`validateStatus`), so axios handed the caller `response.data === ''`.

Every list screen does `stands ?? []` — `''` is not nullish, so it passed straight through as
`stands`, and `stands.length === 0` rendered the empty state. React Query's own cache was correct;
the interceptor layer silently threw the real data away underneath it on the very next 304.

Secondary defect in the same map: it was keyed by URL only, not by query params or locale. A
filtered list and an unfiltered list (or the same list read in two different `Accept-Language`s)
could serve each other's cached ETag despite the backend's own
`Vary: Authorization, Accept-Language`.

## Fix

- `etag.ts` now caches `{etag, data}` together, keyed by `locale|url|sortedParams` — mirrors the
  backend's own locale-scoped Redis keys (`id:<locale>:<uuid>`, `all:<locale>`).
- `interceptors.ts` only sends `If-None-Match` when a full cache entry (etag *and* body) exists for
  that key; on a `304` response it rewrites `response.status` to `200` and `response.data` to the
  cached body before the caller ever sees it.
- Backend: `internal/config/config.go`'s `defaultCacheHTTPMaxAge` dropped from `30s` to `0`
  (→ `Cache-Control: private, no-cache`) so the browser's own HTTP cache can no longer serve a
  pre-write body for up to 30s independently of this bug — ETag/304 revalidation is unaffected.
- Screens (`stands-screen.tsx`, `devil-fruits-screen.tsx`) now also render a distinct error state
  (`isError` + retry) instead of falling into the same empty-state branch on a failed GET — the
  empty state and "couldn't load" state were previously indistinguishable, which is why this took
  investigation to tell apart from the SQL/cache-invalidation causes it was first suspected to be.

## Also considered and ruled out

- `ListStandRows`/`ListDevilFruitRows` SQL: `LEFT JOIN LATERAL` on `power_translations`, not an
  inner join — a missing translation row cannot drop a Stand/Devil Fruit from the list.
- Redis read-through cache: `Save` invalidates the whole namespace via a generation `INCR`
  (`cache/redis/cache.go`) before the request returns; correct.
- React Query mutation `onSuccess`: invalidates the unlocalized `standKeys.allLocales` /
  `devilFruitKeys.allLocales` prefix, which prefix-matches every locale branch; correct.
