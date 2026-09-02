---
title: "Feature: SSE replaces picture-status polling (fixed/shipped 2026-08-10)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - feature
---

# Picture pipeline: SSE push instead of polling (2026-08-10)

## Why

[[admin-crud-cache-stale-sw]] and [[admin-panel-crud-ux-fixes]] fixed the data layer (Redis
invalidation, ETag, service worker), but Stand/Devil Fruit/avatar picture uploads still only
updated the UI once the browser tab lost and regained focus. Root cause: the poll-attempt counter
in `use-stands.ts`/`use-devil-fruits.ts`/`use-profile.ts` was `query.state.dataUpdateCount` -
cumulative for that query key's *whole lifetime*, persisted across reloads by
`PersistQueryClientProvider`. After enough refetches (trivially reached across a dev session, or
just normal admin use over time), it permanently exceeded `MAX_POLL_ATTEMPTS` and polling silently
died for good; only `refetchOnWindowFocus` (blur/focus) ever refetched again. Fixed the counter
first (moved it into a per-hook `useRef`, resets whenever nothing's pending), then the owner asked
to replace polling with Server-Sent Events outright.

## Design

**Backend** - new `internal/application/services/picture_event_hub.go`: a plain in-process
pub/sub (`map[chan PictureEvent]struct{}` + mutex, non-blocking `Publish` that drops for a
full/slow subscriber). No Redis pub/sub needed - single backend instance, no fan-out requirement.
`PictureWorker` (`picture_worker.go`) publishes `{Kind, SubjectID, PictureReady}` right after its
`UpdatePicture(READY)` call succeeds, and `{Kind, SubjectID, PictureFailed}` from `markFailed`
(which needed `job.Kind` threaded through - it previously only took an id). Two pre-existing
*silent* failure paths (a `PictureKeys` read error, or `UpdatePicture(READY)` itself erroring)
still leave DB status at PENDING with no event - unchanged from before, not fixed here, since
publishing FAILED without writing FAILED to the DB would desync the event from the record.

New `internal/infrastructure/api/endpoints/events_endpoints.go`: `GET /api/v1/events`, admin-only,
`text/event-stream`, 20s heartbeat comments (`: ping\n\n`) to keep idle proxies from timing it out.

**Auth tradeoff, deliberate**: `EventSource` cannot set custom headers, so this route accepts the
JWT via `?token=` query param (falls back to `Authorization` header) via a small helper local to
`events_endpoints.go` - *not* a change to the shared `RequireAuth` middleware, so no other route
gains query-param auth. The token lands in chi's `middleware.Logger` output and (in prod) Nginx
Proxy Manager's access log, and stays usable there for the whole of `JWT_TTL` — which is **24h**,
not the 8h this note claimed until 2026-09-02. `deployments/.env.example` ships `JWT_TTL=24h` and
`config.go`'s `defaultJWTTTL` is `24 * time.Hour`, so 24h is what is deployed; [[backend-contract]]
said 24h all along. The note was wrong, the value was not — see [[auth-hardening-2026-09-02]].
Still accepted for v1 as the standard/only practical way to authenticate `EventSource`, with the
longer window making the follow-up slightly more attractive than the old text implied: a harder
mitigation (a `POST /api/v1/events/ticket` minting a short-TTL single-use ticket just for the SSE
connection) is a reasonable next step if this ever needs hardening - not built now.

`router.go`: `middleware.Timeout(60s)` was global, which would have killed the stream at 60s.
Moved it from the top-level `r.Use` into two inner `r.Group`s (health/swagger, and the
auth+stands/devil-fruits/users group) so `/api/v1/events` alone is unbounded - relies on the
client's own reconnect, the heartbeat, and the app's root shutdown `ctx` (passed into
`NewEventsEndpoints` so the stream loop exits promptly on SIGINT/SIGTERM) instead.

**Frontend** - new `src/providers/picture-events-bridge.tsx`, mounted once in `app-providers.tsx`
(inside `QueryProvider`, sibling to `ErrorBoundary`/`ToasterMount`). Web-only (`Platform.OS`
guard) - React Native has no built-in `EventSource`; native keeps the ref-based polling as its
fallback, gated with `Platform.OS === 'web' ? undefined : pollInterval` in all three hooks.
Manual reconnect (not the browser's built-in retry): the connection URL bakes the token in at
connect time, so a native auto-retry would keep hitting a dead URL if the session token rotated
while disconnected. On `onerror`, closes and reconnects reading the *current* token fresh from
`useSessionStore.getState()`, same backoff shape as the old polling (2s→4s→8s...→30s) but no
permanent give-up - this is now the only notification path. On any reconnect that follows a prior
successful connection, fires a full `invalidateQueries` across all three query keys as a
safety-net resync, since the event stream has no durable replay log and could have missed events
while disconnected.

## Manual step still needed in production

Nginx Proxy Manager is GUI-managed on the host, no config file in this repo (per
[[cicd-deployment]]). Its `/api` location's *Advanced* tab needs, by hand, before this is trusted
in prod:

```
proxy_buffering off;
proxy_read_timeout 1h;
proxy_http_version 1.1;
proxy_set_header Connection '';
```

Without `proxy_buffering off`, NPM buffers the whole response and the browser never sees events
until the buffer fills or the connection closes. Local dev is unaffected - no reverse proxy in
`docker-compose.dev.yml`.

## Where things live

- Backend: `internal/application/services/picture_event_hub.go` (+ its test),
  `picture_worker.go`, `internal/infrastructure/api/endpoints/events_endpoints.go`, `router.go`,
  `cmd/app/main.go`.
- Frontend: `src/providers/picture-events-bridge.tsx`, `app-providers.tsx`,
  `use-stands.ts`/`use-devil-fruits.ts`/`use-profile.ts`.

See also [[admin-crud-cache-stale-sw]] (the service-worker fix from the same debugging session),
[[admin-panel-crud-ux-fixes]] (original picture-polling design this replaces),
[[cicd-deployment]] (NPM routing, why its config isn't in-repo).
