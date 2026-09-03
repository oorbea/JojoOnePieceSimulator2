---
title: "Feature: stream connection tickets for SSE + game WebSocket (2026-09-03)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - feature
  - auth
  - decision
---

# Stream connection tickets (2026-09-03)

Closes accepted risk #1 from [[auth-hardening-2026-09-02]]: the SSE stream (`GET /api/v1/events`)
and the game WebSocket (`GET /api/v1/games/{id}/ws`) used to authenticate via the full 24h access
token in the query string (`?token=<jwt>`), because neither `EventSource` nor the browser
`WebSocket` can set an `Authorization` header. That token sat in chi's `middleware.Logger` output
and, in prod, Nginx Proxy Manager's access log, for the whole of `JWT_TTL`. Both
[[picture-events-sse]] and [[game-realtime-transport]] named the fix ahead of time: a short-lived,
single-use ticket minted just for the connection. This tanda builds it.

## Design decisions

**Opaque token in a server-side store, not a second JWT.** Single-use requires a burn store no
matter what (a signed token can't be un-issued). Once the store exists, a signature buys nothing —
the store lookup already authenticates the token, and the stored record carries `UserID`/`Role`/
purpose/resource as native Go values instead of re-parsed claims. A second JWT would mean two token
formats, a `purpose` claim, a second TTL knob on `ITokenIssuer`, *and* the store — strictly more
moving parts for the same guarantee. Concretely: 32 bytes from `crypto/rand`,
`base64.RawURLEncoding` → 43 URL-safe characters.

**No IP binding.** In prod the app sits behind Nginx Proxy Manager and `router.go` installs
`middleware.ClientIPFromRemoteAddr`, which resolves to *the proxy's* IP for every caller — IP
binding would be uniformly worthless there while breaking legitimate clients whose IP changes
between mint and connect (mobile carrier hand-off, VPN). Short TTL + single-use + purpose/resource
binding is the whole value; IP adds risk (breakage) with no protection in this deployment. If the
router ever moves to `ClientIPFromXFF` (trusting the proxy's forwarded-for header), IP binding
becomes cheap and meaningful and could be added then.

**Ships both a Redis store and an in-memory one**, mirroring `gamestore`'s split exactly
(`internal/infrastructure/streamticket/` for memory + reaper, `streamticket/redis/` for the
Redis-backed one). `REDIS_URL` unset ⇒ memory, same as the game store and the cache layer. A
30-second ticket in the in-memory store loses nothing on restart — every open stream is already torn
down on shutdown (`e.ctx.Done()`), so the client just re-mints on its next reconnect.

**TTL 30s** (`STREAM_TICKET_TTL`, default), long enough for a slow mobile mint round-trip plus the
upgrade, short enough that a ticket leaked via an access log is already dead by the time anyone
could reuse it — the residual risk this leaves is 30s/single-use/purpose+resource-scoped, versus a
24h/multi-use/full-authority JWT before.

**Two mint endpoints, not one generic one**: `POST /api/v1/events/ticket` (behind `RequireAuth` +
`RequireAdmin`) and `POST /api/v1/games/{id}/ws-ticket` (behind `RequireAuth`, plus the exact same
`GetGame` + `resolveParticipant` check `serveWS` itself does). The purpose is implicit in the path,
the game id comes from `chi.URLParam`, and no new wire enum was needed through `cmd/typegen` (a
generic endpoint would have needed a `purpose` field in a new DTO). A real side benefit: a
non-participant now gets an ordinary axios-readable `403 FORBIDDEN` / `404 GAME_NOT_FOUND` from the
mint, instead of a WebSocket handshake failure the browser cannot inspect at all.

**`?token=` removed outright, not kept as a fallback** (owner decision) — the query param only
accepts `?ticket=`; `Authorization: Bearer` keeps working everywhere it did (curl, native clients,
tests). A stale deployed frontend still sending `?token=<jwt>` falls through to a plain 401.

## Where things live

- `internal/domain/ports/stream_ticket.go` — `TicketPurpose` (`events`/`game-ws`), `StreamTicket`,
  `IStreamTicketStore{Issue, Redeem}`. `ports/errors.go` — `ErrTicketInvalid` (distinct from
  `ErrUnauthenticated` so logs/tests can tell "no credential" from "bad ticket", both map to 401).
- `internal/infrastructure/streamticket/{memory,reaper}.go` — `MemoryStore` (single critical
  section: `Redeem` deletes unconditionally *before* checking expiry, which is what makes it
  single-use under concurrency), `Reaper` (mirrors `gamestore.Reaper`).
- `internal/infrastructure/streamticket/redis/{store,wire}.go` — `SET NX PX ttl` for `Issue`, a Lua
  `GET`+`DEL` script for the atomic burn on `Redeem` (not `GETDEL`: needs Redis ≥6.2, and every
  other adapter here is already Lua-shaped). Key: `jojo:stream-ticket:<token>`.
- `internal/infrastructure/api/endpoints/middleware.go` — `authenticateStream(r, issuer, tickets,
  purpose, resource)` replaces the two verbatim-duplicated `authenticate`/`authenticateWS` helpers
  that used to live on `EventsEndpoints`/`GameEndpoints`.
- `events_endpoints.go`'s `mintTicket`, `game_endpoints.go`'s `mintWSTicket`. `game_ws_endpoints.go`'s
  `serveWS` now runs `ParseGameID` *before* auth (the ticket's `Resource` check needs the id), and
  the ticket is burned before `websocket.Accept` — a rejected origin therefore costs the ticket, but
  the client just re-mints on its next attempt.
- `apierr.StreamTicketInvalid` (`STREAM_TICKET_INVALID`, 401). **Only observable on the mint
  responses** (ordinary `UNAUTHENTICATED`/`FORBIDDEN`/`GAME_NOT_FOUND`) — `EventSource.onerror` and a
  failed WS handshake expose no status/body to the browser, so no client behavior should ever be
  designed around this code on the stream paths themselves.
- `dto.StreamTicketResponse{Ticket, ExpiresAt}`, registered in `cmd/typegen/registry.go`.
- Config: `STREAM_TICKET_TTL` (default 30s), `STREAM_TICKET_REAP_INTERVAL` (default 1m, memory-store
  only), `RATE_LIMIT_TICKET_PER_USER` (default 60/window, keyed by user id like the other tiers).
  `RATE_LIMIT_GLOBAL_PER_IP` raised 120→240: every connect now costs two requests (mint + stream)
  instead of one, and in prod every caller collapses to NPM's IP under `ClientIPFromRemoteAddr` — see
  `ratelimit.go`'s own doc comment, which already anticipated switching to `ClientIPFromXFF` for a
  real per-user key; that switch is its own future tanda, not folded into this one.
- Frontend: `src/shared/api/stream-tickets.ts` (`mintEventsTicket`/`mintGameSocketTicket`),
  `features/game/lib/socket-url.ts` (query param renamed `token`→`ticket`),
  `features/game/stores/game-socket.store.ts`, `providers/picture-events-bridge.tsx`.

## The async `connect()` race (the actual risk in this tanda)

Both the socket store's `connect()` and the SSE bridge's `connect()` went from synchronous
(open-immediately) to `async` (mint-then-open). That opens a window that didn't exist before: a
`detach()`/`attach()`-to-another-game, or a role change, arriving while a mint is still in flight.

Both got the same fix — a module-level generation counter bumped on teardown (`closeSocket()` for
the store, `disconnect()`'s `cancelledRef` for the bridge), captured before the `await`, checked
after it before touching any state. `game-socket.store.ts` additionally needed a `pendingConnect`
flag: unlike the old synchronous `connect()`, `attach()`'s `if (!socket) connect(gameId)` guard
alone can no longer tell "already connecting" from "not connected" once there's an await in between,
so a second `attach()` call for the *same* game while a mint is in flight would otherwise fire a
second, concurrent mint.

**Mint-failure policy** (both call sites): 401 → `closed` (the axios response interceptor already
cleared the session, nothing to retry); 403/404 → terminal `closed` (this is a real behavior fix —
before tickets, a non-participant's WS handshake 403 and a non-admin's SSE 403 both retried on the
backoff curve forever, since the browser couldn't see *why* the connection failed); anything else
(429, 5xx, network) → the existing reconnect backoff, which now re-mints on its next attempt.

## Verification

Backend: `go build ./... && go vet ./... && go test ./... -race` (Docker, `-race` needs
`CGO_ENABLED=1` + gcc — the norm's `backend-test` image doesn't have cgo by default, so this needed
`apk add gcc musl-dev` on top of the usual `norma-verificacion-docker.md` recipe). New coverage:
`streamticket/memory_test.go` (single-use under concurrent `Redeem`, expiry, entropy/format,
reap), `streamticket/redis/store_test.go` (same matrix, `TEST_REDIS_URL`-or-skip), new
`events_endpoints_test.go` (mint auth, stream auth including the `?token=` regression guard and the
wrong-purpose case), extended `game_ws_endpoints_test.go`'s sibling `game_endpoints_test.go` with
real `websocket.Dial` connect tests (valid/reused/cross-game/wrong-purpose/old-token/Bearer-still-
works). Frontend: `game-socket.store.test.ts` (every pre-existing test needed a `flush()` after
`attach()`/`retryNow()` since the mint is now an awaited microtask; new cases for 401/403/network-
error/the `pendingConnect` guard), new `socket-url.test.ts` + `socket-url.unconfigured.test.ts`, new
`picture-events-bridge.web.test.tsx` (jest.config.js's `logic` project testMatch extended to cover
`src/providers/__tests__/**/*.web.test.tsx`, mirroring the existing hooks pattern). All green:
backend `go test ./...` clean including `-race`; frontend `pnpm typecheck && pnpm lint && pnpm
test:ci` clean.

Contracts regenerated: `make types-docker` (host `go run` blocked by Windows App Control, per
[[feedback_backend_tests_via_docker]]) + `make swagger` via the same `typegen` container — both
`/events/ticket` and `/games/{id}/ws-ticket` now have swagger annotations.

Related: [[picture-events-sse]], [[game-realtime-transport]], [[auth-hardening-2026-09-02]],
[[backend-contract]], [[norma-verificacion-docker]], [[contratos-tipos-generados]].
