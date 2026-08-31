---
title: "Feature: game realtime transport (2026-08-11)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - feature
  - gameplay
---

# Game realtime transport (2026-08-11)

Closes the second half of [[gameplay-application-layer]]'s "still not built" list: a native
WebSocket endpoint plus the plain HTTP routes for the game feature. Before this, `GameService` was
built and tested but literally unreachable — `cmd/app/main.go` constructed it and threw it away
(`_ = gameService`). See [[game-lobby-persistence]] for the store/persistence half that shipped
alongside this.

## Status

Done, backend-only. No frontend WebSocket client and no admin UI for `/stages` yet — see
[[gameplay-application-layer]]'s updated "still not built" section.

## Library: `github.com/coder/websocket`, not socket.io

Native WebSocket, `context`-native API (`conn.Read(ctx)`) matching the existing SSE precedent
(`events_endpoints.go` holds the app's root ctx so the stream dies on shutdown), zero transitive
deps. The frontend has `socket.io-client` installed but unwired — it stays that way until a later
tanda migrates the client onto this protocol (`EXPO_PUBLIC_SOCKET_URL` should then point at
`ws://.../api/v1`, and the client builds
`${EXPO_PUBLIC_SOCKET_URL}/games/${id}/ws?token=${accessToken}`).

## Protocol (`internal/infrastructure/api/dto/game_ws.go`)

Envelopes: `ClientCommand{type, requestId?, payload?}` / `ServerFrame{type, requestId?, payload?}`.
`requestId` is opaque, only used to correlate an `ERROR` frame back to the command that caused it.

**Confirmed reusable on the frontend (2026-08-31)**: `useGameSocketStore`'s `send()` returns the
`requestId` it stamps on the outgoing command, and `lastError.requestId` on the store surfaces it
back. `LobbyRoomContainer`'s config-edit save is the first caller to actually correlate the two
(attributing an `ERROR` frame to its own in-flight `UPDATE_CONFIG` rather than showing a generic
toast) - see [[game-lobby-frontend]]'s "Config-edit panel hardened" section. Worth reusing for any
future command with its own submit-state UI rather than re-deriving the pattern.

**Commands** (each maps to exactly one `GameService` method): `LEAVE`, `ADD_BOT`, `REMOVE_BOT`,
`START`, `ABORT`, `VOTE`, `RESYNC`. Deliberately **not** commands: `CreateGame`/`JoinByCode`/
`GetGame*` (plain HTTP — you can't join a game over a socket you can only open by already being a
participant), `Disconnect`/`Reconnect` (socket lifecycle only, so a client can never forge its own
or anyone else's presence), `CloseVotingWindow` (server-internal timer + early-close paths only —
reachable from a command, any player could guillotine a vote).

**Frames**: `STATE` (the full snapshot), one per domain event reusing `entities/game/events.go`'s
wire names verbatim (`PLAYER_JOINED`, `VOTING_OPENED`, `VOTE_CAST`, `ROUND_RESOLVED`,
`GAME_FINISHED`, ...), `ERROR` (reuses `errorCode(err)` so the frontend's `errors.<CODE>` i18n
lookup works over the socket with zero new mapping), `RESYNC_REQUIRED`.

**Event + snapshot rule**: after any state-changing event, the connection sends a fresh `STATE`.
The client never reconstructs state from deltas, a hub-dropped event self-heals on the next one,
and ordering races are transient since `STATE` always arrives last and always wins. `VOTE_CAST`,
`GAME_FINISHED`, and `GAME_ABORTED` are the exceptions — the first is high-frequency and
self-describing, the latter two fire right as/after `GameService.finalizeLocked` deletes the game
from the store, so a fresh `STATE` fetch would just race `ErrGameNotFound`.

## Votes hidden until a round resolves (the owner's explicit call)

- `VOTE_CAST` carries only `{roundIndex, votesCast, voters}` over the wire — no participant, no
  option — even though the domain event carries both. `votesCast`/`voters` are an anonymous
  human-vote-progress count (connected humans only, bots excluded), added 2026-08-26 — see
  [[game-vote-buttons-2026-08-26]]. The domain itself was **not** changed to hide who/what voted;
  this is a transport-only redaction, and the two counters leak strictly less than
  `GameRoundResponse`'s own `votedParticipantIds` already does while a round is live.
- In `GameStateResponse`, a round still voting exposes `votedParticipantIds` (who voted, not what)
  plus the caller's own `you.vote`; once `Result` is set the full `votes` map is revealed. Both
  derive from the new `Ballot.Votes()` getter (see [[game-lobby-persistence]]).
- **Exception, added 2026-08-28** (see [[game-round-result-2026-08-28]]): `Round.TiedVotes` — the
  ballot as it stood the instant a tie opened a revote, captured right before `Ballot.Reset()` wipes
  it — is revealed via `GameRoundResponse.tiedVotes` **while the round is still live** (state
  `TIEBREAK`, no `Result` yet). This is a deliberate owner call to show what tied before the revote
  replaces it, not a change to the "hidden while live" rule for the *current* ballot - `VOTE_CAST`
  itself stays exactly as anonymous as before.
- **Loadouts stay fully public**, by contrast: in Gauntlet you're voting whether the squad survives
  a Stage, in Versus which team wins the round — judging either without seeing the powers in play
  is arbitrary, and the domain's own bot-voting (`Game.optionScores`) already assumes full
  information across every team.

## Connection lifecycle (`game_ws_endpoints.go`)

- `GET /api/v1/games/{id}/ws?token=<jwt>` (also accepts `Authorization: Bearer` first) — copies
  `events_endpoints.go`'s query-param-auth pattern verbatim, same reasoning (browser `WebSocket`,
  like `EventSource`, can't set headers). Pre-upgrade errors are plain JSON via `handleError`.
- `originPatterns()` strips the scheme off `CORSConfig.AllowedOrigins` for
  `AcceptOptions.OriginPatterns` (which wants `host[:port]`, not full origin URLs). Empty list ⇒
  same-origin only, mirroring CORS's own deny-by-default.
- `resolveParticipant` scans `g.Participants()` for the caller's `UserID` — the same scan
  `GameService.JoinByCode` already does. No match ⇒ `ports.ErrForbidden` (403), reused as-is rather
  than a new code.
- `connRegistry` reference-counts sockets per `(GameID, ParticipantID)` — two browser tabs are two
  sockets for one participant; `Reconnect` only fires on the first, `Disconnect` only on the last.
- Subscribes to `GameEventHub` **before** anything else that could race a concurrent mutation, so
  no event can slip between the initial snapshot and the subscription.
- One write-pump goroutine (`coder/websocket` allows only one writer) draining an `outbound`
  channel plus a `heartbeatInterval` (20s, shared const with SSE) protocol-level `Ping`; one
  read-pump loop dispatching commands. A malformed command gets an `ERROR` frame and the connection
  **stays open** — only a genuine protocol violation (read-limit exceeded) closes it.
- **Backpressure**: `GameEventHub.Publish` silently drops for a full subscriber buffer (8) with no
  signal. Mitigated by (1) the forwarder draining straight into `outbound` with no I/O in between,
  so the hub buffer rarely fills; (2) the event+snapshot rule self-healing a drop; (3) if
  `outbound` itself (32) is ever full, the connection is closed with `StatusPolicyViolation` and
  the client reconnects to a fresh snapshot — safer than serving a silently stale view.
- On close: cancel the connection's ctx (also tied to the app's root ctx, so shutdown kills every
  open socket), wait for both goroutines, unsubscribe, then — only if this was the *last* socket for
  that participant — call `Disconnect` with a **fresh** background context (the request's is
  already dead, but `Disconnect` still has to run: it can reassign host, abort the game, or
  early-close a vote).

## Mounting: `/games` sits outside the `Timeout(60s)` group

`router.go` already excludes `/events` from the per-group 60s timeout for the same reason. chi
can't mount two handlers with different middleware on the same pattern, so `GameEndpoints.Routes`
applies its **own** `Timeout`+`RequireAuth` to the REST sub-group (`POST /`, `POST /join`,
`GET /{id}`, `GET /by-code/{id}`) while leaving `GET /{id}/ws` bare; `router.go` mounts the whole
`/games` router alongside `/events`, outside its own Timeout group.

## HTTP routes (`game_endpoints.go`)

`POST /games`, `POST /games/join`, `GET /games/{id}`, `GET /games/by-code/{code}` — everything that
must be reachable without a live socket (creation, resume). Both `GET`s **403 a non-participant**
(same `resolveParticipant` as the socket), so a guessed UUID or a leaked join code can't dump a
lobby's roster/loadouts to a stranger — `by-code` is therefore a *resume* route, not a pre-join
preview (`GET /games/preview?code=` is deferred, see below).

## Stage catalog admin routes {#stages}

`GET /stages` (optional `?manga=`), `GET /stages/{id}` open to any authenticated user;
`POST/PUT/DELETE /stages` admin-only (`RequireAdmin`), inside the normal `Timeout` group. See
[[game-lobby-persistence]] for the schema.

## New error codes

`STAGE_NOT_FOUND` (404), `STAGE_ALREADY_EXISTS` (409), plus three sentinels that already existed
in the domain but had no status/code branch before this tanda — `EMPTY_TEAM_NAME`,
`INVALID_PARTICIPANT_KIND`, `INVALID_SQUAD_VERDICT` (all 400, all previously fell through to 500
`INTERNAL`) — and `UNKNOWN_COMMAND` (400, new WS-only sentinel). All in lockstep across
`error_codes.go`/`errors.go`, translated in all three locales.

## Tests

`internal/infrastructure/api/endpoints` package tests (existing suite, extended) cover the new
routes via `httptest` against the real router with fake token issuers; `go vet`/`go build ./...`
clean. Verified live against the actual dev stack after rebuilding the backend container: `/stages`
and `/games` mount correctly (401 unauthenticated without a token, 405 for an undefined method),
migrations `00008`/`00009` applied cleanly on startup, and the full `-tags integration` suite
(including the new `stage_repository_test.go`/`game_history_test.go`) passes against the live
Postgres/Redis containers via `make test-integration-docker`.

## Deferred / not built here

- `GET /games/preview?code=` — a pre-join preview (mode/team size/player count without full
  roster/loadouts).
- `DELETE /games/{id}/participants/me` — leaving without a live socket (today `LEAVE` is WS-only).
- A `Seq` counter on `services.GameEvent` so a hub drop could be *detected* by the transport instead
  of only self-healed by the next `STATE`.
- Frontend WebSocket client, game screens, admin CRUD screens for stages.

Related: [[gameplay-application-layer]], [[game-lobby-persistence]], [[gameplay-domain-design]],
[[picture-events-sse]], [[ADR]], [[backend-contract]].
