---
title: "Bug: admin CRUD looked broken in prod - stale service worker, not the data layer (fixed 2026-08-10)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - bug
  - fixed
---

# Bug: admin Stand/Devil Fruit CRUD invisible until a manual cache clear, in prod only

## Symptom

Creating/editing a Stand or Devil Fruit in the admin panel never showed up in the list -
identical-sounding symptom to [[etag-304-body-loss]] and [[admin-panel-crud-ux-fixes]]'s bug #1,
but reported again *after* both of those shipped (prod was on PR #15, which already contains both
fixes). Only reproduced in production, not confirmed in local dev.

## Investigation

Re-audited every cache layer top to bottom and found the data path already correct in the current
source:

- Redis (`apps/backend/internal/infrastructure/cache/stand_repository.go:43,155,168`): `Save`/
  `Delete`/`UpdatePicture` all call `invalidate()` (namespace-wide `INCR`) on success.
- HTTP ETag (`.../api/endpoints/cache_headers.go:76-90`): the ETag is a SHA-256 of the response
  body, so changed content structurally cannot produce a `304`.
- `CACHE_HTTP_MAX_AGE=0s` → `Cache-Control: private, no-cache` (set in the etag-304-body-loss fix).
- Frontend mutations (`use-stand-mutations.ts`/`use-devil-fruit-mutations.ts`) call `clearEtags()`
  before `invalidateQueries(allLocales)` (the admin-panel-crud-ux-fixes fix).

None of that survives a page reload - but the report says the symptom does, and specifically that
clearing the *browser* cache (not just reloading) fixes it. That points at whatever survives a
plain reload and *is* wiped by "clear cache":

1. **`apps/frontend/public/sw.js`** - the previous version cache-first'd every same-origin GET,
   including `index.html` and the hashed JS bundle, with no revalidation and no version bump since
   `jops-shell-v2`. A browser that had visited the site before PR #15 kept being served whatever
   `index.html`/bundle existed at that first visit, forever - i.e. the *browser* was still running
   pre-fix JS, even though the backend and the freshly-deployed frontend image were both correct.
   `docker compose up -d --build` on the server changes what nginx serves; it does nothing to a
   tab that already has an old shell cached client-side.
2. `PersistQueryClientProvider` (`src/providers/query-provider.tsx`) persisting the React Query
   cache to `localStorage`/AsyncStorage with no `buster`/`maxAge` - secondary risk, not confirmed
   as the actual trigger here, but the same class of bug (stale client-side state surviving a
   deploy) and cheap to close at the same time.

## Fix

- `apps/frontend/public/sw.js`: rewritten fetch strategy. Navigations/HTML now go **network-first**
  (fetch, cache on success, fall back to the cached shell only if the fetch itself fails - offline).
  Only `/_expo/static/**` (content-hashed, matches `nginx.frontend.conf`'s `immutable` rule) stays
  **cache-first**. Everything else passes through uncached (no more opportunistic caching of
  arbitrary same-origin GETs). `CACHE_NAME` bumped `v2` → `v3` so already-stuck clients evict their
  old cache on the next SW update cycle regardless.
- `src/providers/query-provider.tsx`: `persistOptions` now sets `maxAge: 24h` and
  `buster: env.EXPO_PUBLIC_BUILD_ID` - a persisted cache from a different build is discarded
  instead of rehydrated.
- `EXPO_PUBLIC_BUILD_ID` added end-to-end: `shared/config/env.ts` (zod schema, default `'dev'`),
  `Dockerfile.frontend`/`docker-compose.yml` (build arg, same mechanism as `EXPO_PUBLIC_API_URL`
  since these are inlined at build time), `.env.example`, and `.github/workflows/cd.yml` pins it to
  `${{ github.sha }}` right after the secrets/vars dump so every deploy gets a distinct value.

## Verification

Rebuilt the frontend image twice in local Docker with a temporary marker string change, using a
Playwright `launchPersistentContext` profile to keep one browser's SW/cache state across both
builds (simulating a tab left open across a deploy - the exact reported scenario):
1st build (no marker) → load, confirm `caches.keys()` returns `['jops-shell-v3']` → 2nd build
(marker added, no cache clear, no hard reload) → plain reload of the *same* profile → marker text
present, zero console errors. This is what the previous SW would have failed: it would have kept
serving the 1st build's cached `index.html`/bundle indefinitely.

Did not re-verify the Redis/ETag data path end-to-end against a real admin session - crafting a
throwaway admin JWT for that (there's no dev-login bypass in this app, only Google OAuth) was
blocked by the coding sandbox's auto-mode classifier as forging an auth credential, even for
localhost. Left unexercised on this pass; that layer is unchanged by this fix and was already
re-confirmed correct by code reading (see Investigation above) plus the two prior fix notes.

## Where things live

- `apps/frontend/public/sw.js`
- `apps/frontend/src/providers/query-provider.tsx`
- `apps/frontend/src/shared/config/env.ts`
- `deployments/docker/Dockerfile.frontend`, `deployments/docker-compose.yml`,
  `deployments/.env.example`
- `.github/workflows/cd.yml`

See also [[etag-304-body-loss]] and [[admin-panel-crud-ux-fixes]] (the two earlier, data-layer
fixes for a similarly-described bug), [[cicd-deployment]] (CD pipeline this touches),
[[docker-setup]] (compose/Dockerfile conventions).
