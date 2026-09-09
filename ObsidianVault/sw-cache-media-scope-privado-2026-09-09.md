---
title: "SW media cache excludes the private/signed scope + dev goes same-origin (2026-09-09)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - service-worker
  - performance
  - fixed
---

# SW media cache excludes the private/signed scope (2026-09-09)

> [!success] Status: DONE, shipped, verified in Docker

The vault's "still open: service worker media caching (`jops-img-v1`)" flag
was stale bookkeeping - that feature already shipped 2026-09-08
(`public/sw.js:28`, commit `5f373eb`, closed by `ed94841`). Chasing that flag
down surfaced a real defect in what shipped, which this note fixes, plus a
latent local-dev CSP problem this closes as a side effect.

## The defect: signed avatar URLs were churning out the public cache

`isImmutableImage` cache-first'd all of `/api/v1/media/**`, public
(Stand/DevilFruit/Stage) **and** private (User avatars) scope alike. Public
group ids are permanent - correct to cache forever. Private URLs are not:
`exp` is quantized to `MEDIA_PRIVATE_URL_TTL/2` windows, **12h by default**
(`dto.MediaURLBuilder.privateWindow`), so the same avatar's bytes get a
brand-new URL twice a day.

`trimImageCache` evicts by **insertion order** once `IMG_CACHE_MAX_ENTRIES`
(200) is exceeded - i.e. it evicts the *oldest* entries first. Every avatar
rotation orphans the previous URL and inserts a new one, and that churn
preferentially pushes out the permanently-valid *public* entries the cache
exists to protect. A 10-player lobby generates ~20 rotating entries/day;
against a 200-slot budget that's a real, ongoing eviction pressure on
catalogue art.

## Decision: exclude `/api/v1/media/p/...` from the SW cache (option A)

```js
function isImmutableImage(url) {
  return (
    url.origin === self.location.origin &&
    url.pathname.startsWith('/api/v1/media/') &&
    !url.pathname.startsWith('/api/v1/media/p/')
  )
}
```

Avatars fall through to the SW's passthrough (no `respondWith`) and are
handled by the browser's normal HTTP cache, which already gets
`private, max-age=..., immutable` from the backend
(`media_endpoints.go:190`) - so this loses nothing for a warm browser tab,
only cross-session/offline avatar caching, which this app doesn't otherwise
support anyway (it needs a live WS/API connection). See
[[media-proxy-content-addressed]]'s "Public vs private scope" section for
the URL shape.

**No `CACHE_NAME` bump.** The file's own norm says bump on any logic change,
but the reason to bump is to evict caches that might serve something stale
- this change can't do that, it only *stops* caching a scope. Bumping would
also be self-defeating here since the cache being protected is
`IMG_CACHE_NAME`, not the shell. Documented as a deliberate exception
directly in `sw.js`'s header comment. Existing `/p/...` entries already in
visitors' `jops-img-v1` are simply never matched again and age out via the
normal insertion-order cap - no active purge needed.

### Alternatives considered and rejected

- **Normalize the cache key** (strip `exp`/`sig`, cache under
  `/group/variant`): best hit rate, avatars stay cacheable offline - but the
  signature is the *only* authorization private media has, and Cache
  Storage has no TTL. Normalizing the key moves authority out of the URL the
  design deliberately put it in, for a marginal win on a resource this app
  doesn't need offline.
- **Two caches with separate budgets** (`jops-img-v1` + a small
  `jops-avatar-v1`): would preserve avatar caching without touching the auth
  model, but adds a third name the `activate` handler's allowlist must
  remember - exactly the footgun the file already warns about (forgetting a
  name there wipes every visitor's images on the next deploy).
- **Widen `MEDIA_PRIVATE_URL_TTL`**: would reduce churn (fewer rotations/day)
  but is a security tradeoff (how long a leaked avatar URL stays valid), not
  a caching one - shouldn't be decided as a side effect of a cache tuning
  pass.

## Side effect: local dev now same-origin for media (and a latent CSP bug closed)

`MEDIA_BASE_URL` in local dev pointed straight at the backend
(`http://localhost:8080/api/v1/media`) - cross-origin from the frontend's
`:3000`. Two consequences, both closed by this change:

1. The SW's `fetch` handler bails out early on any cross-origin request
   (`sw.js:139-141`), so `cacheFirstImage` was **dead code in local dev** -
   this whole feature only ever ran in prod, and had never been verified
   live before this tanda.
2. The frontend's CSP is `img-src 'self' data: blob: https:` - an absolute
   `http://localhost:8080` URL doesn't match `'self'` (different origin) or
   `https:` (it's plain `http:` in dev). This was silently masked because
   most local rows still resolve through R2's presigned `https:` fallback
   (the media backfill, `cmd/mediabackfill`, has never been run - see
   [[media-proxy-content-addressed]]'s "Still open").

Fixed by making dev same-origin instead of matching the CSP to dev's
cross-origin shape: `nginx.frontend.conf.template` gained an
`/api/v1/media/` location that proxies to `backend:8080` (deliberately using
`resolver 127.0.0.11 valid=10s` + a `set` variable rather than a literal
`proxy_pass http://backend:8080` - a literal upstream is DNS-resolved at
nginx startup, and this container can start before the backend's name
exists, which would crash-loop it), and `deployments/.env`'s
`MEDIA_BASE_URL` is now `http://localhost:3000/api/v1/media` - absolute (so
native, which has no "relative to the page" concept, still works), but same
origin as the frontend as far as the browser and the SW are concerned.
Deliberately **no `add_header` inside that location block** - the file's own
IMPORTANT note says a location's `add_header` replaces every inherited one,
so the fix is to declare none and let the server block's CSP/security
headers apply unmodified.

This location is inert in prod: NPM routes `/api` straight to `backend:8080`
before the frontend container ever sees the request
(`docker-compose.prod.yml`).

See [[media-proxy-content-addressed]]'s new "MEDIA_BASE_URL / orígenes"
section for the full origin-by-environment table.

## Tests

`apps/frontend/src/test/__tests__/sw.test.ts` (new) - exercises the real
`public/sw.js` loaded from disk via `new Function('self', 'caches', 'fetch',
src)` (chosen over `vm.runInNewContext`, which lacks a shared `URL`/`Promise`
identity, and over assigning jsdom globals directly, which doesn't work -
`window.self`/`location` aren't assignable there). A fake `CacheStorage`
backed by `Map`s preserves insertion order, which `trimImageCache` depends
on. Cache names are discovered from behavior rather than hardcoded, so a
future `CACHE_NAME` bump doesn't need the test edited.

Covers: `activate` preserves exactly the two live cache names and deletes
everything else (the expensive failure mode - wiping every visitor's images
on a deploy); `/api/v1/media/p/...` is never intercepted (this tanda's
actual change); public media still cache-firsts both ways (miss and hit);
cross-origin/non-GET guards; the `trimImageCache` eviction cap; the
`/_expo/static/` shell-asset path; and `networkFirst`'s three branches.

Verified in Docker per [[norma-verificacion-docker]]: backend
`go build`/`go vet`/`go test ./...` clean (comment-only change in
`dto/media.go`), `make types-check-docker` shows zero contract drift,
frontend `pnpm typecheck && pnpm lint && pnpm jest` green - 63 suites / 1221
tests.

**Live browser verification, done (2026-09-09, `local-up` + `claude-in-chrome`)**:
`FRONTEND_PORT=8081` in this local `.env`, not 3000 - the running frontend was at
`localhost:8081`, and `MEDIA_BASE_URL` had to (and does) match that same origin.
`cmd/mediabackfill` had never been run before (see
[[media-proxy-content-addressed]]'s "Still open"), so there was no real
group id to test against - ran it live (`docker exec ... mediabackfill`, no
`-dry-run`) as part of this verification, which incidentally closes that
other open item too: 3 rows backfilled (2 DevilFruit + 1 User avatar), 24
skipped (already had a group or nothing to backfill from).

- Confirmed `curl http://localhost:8081/api/v1/media/<real-group>/card.webp`
  returns a real 200 `image/webp` with
  `Cache-Control: public, max-age=31536000, immutable` through the new nginx
  proxy - same body/headers as hitting the backend on `:8080` directly.
- Confirmed the SPA-fallback-swallows-media-404 risk this used to have is
  gone: the same nonexistent-group media path now 404s with
  `application/json` (matching the backend verbatim) instead of 200'ing with
  `index.html`, while an unrelated non-media route still correctly 200s with
  `text/html` (SPA fallback intact for everything else).
- In a real Chrome tab (`claude-in-chrome`), with the SW `activated`:
  `fetch()`-ing the real public group URL populated `caches.open('jops-img-v1')`
  with exactly that URL; `fetch()`-ing a synthetic `/api/v1/media/p/...` URL
  (401, invalid signature - the *scope exclusion* doesn't care whether the
  signature would've been valid) did **not** appear in any cache. `caches.keys()`
  returned exactly `['jops-shell-v4', 'jops-img-v1']`, matching the unit
  tests' name-discovery assumption.
- Re-fetching the same public URL showed `transferSize: 0` with a real
  response body and a nonzero `workerStart` in the Resource Timing entry -
  the standard signature of "served by the service worker, network never
  touched" (a CSP or cache miss would instead show a nonzero transfer size
  or a failed/opaque response).
- No CSP violations in the console for either fetch (both came back with
  real status codes, not a blocked/opaque response) - confirms the
  same-origin fix actually closes the latent `img-src` problem, not just in
  theory.

Related: [[media-proxy-content-addressed]], [[admin-crud-cache-stale-sw]]
(the prior SW incident and the rules it produced - network-first
navigations, only content-hashed paths cache-first, bump `CACHE_NAME` on
logic changes unless there's a documented reason not to),
[[entrega-imagenes-red-lenta-2026-09-07]], [[docker-setup]],
[[csp-y-rate-limit-por-ip-2026-09-05]], [[norma-verificacion-docker]].
