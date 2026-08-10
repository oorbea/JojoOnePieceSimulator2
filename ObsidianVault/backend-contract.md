---
title: Backend API contract
tags:
  - project
  - jojo-onepiece-simulator
  - backend
---

# Backend API contract

Go/chi backend, `apps/backend`. Base path `/api/v1`, port 8080 (`apps/backend/internal/config/config.go:187`).

## Auth

- `POST /api/v1/auth/google` `{idToken}` → `{accessToken, tokenType:"Bearer", expiresAt, user{id,email,username,completeName,avatar,avatarThumb,avatarStatus,role,language}}`
  (2026-08-04: `picture` was replaced by `avatar`/`avatarThumb`/`avatarStatus` — see [[user-profile-feature]]. 2026-08-06: `language` added — see [[i18n-multi-language]])
- 201 first login, 200 returning.
- **No refresh token, no cookies.** 24h JWT. 401 anywhere → force re-login, no silent refresh.
- All other routes: `Authorization: Bearer <token>` required.
- Writes additionally require `role === "ADMIN"`.
- **Frontend gotcha (bit us once, 2026-08-01)**: axios `baseURL` (`EXPO_PUBLIC_API_URL`) already includes `/api/v1` — feature API calls must use bare paths (`/auth/google`, not `/api/v1/auth/google`) or you get a silent double-prefixed 404. See [[frontend-stack]] for the full incident.
- **Web OAuth flow is full-page redirect, NOT popup** (`use-google-auth.ts`): `expo-auth-session`'s popup flow breaks because `accounts.google.com` sends its own strict COOP header, permanently severing `window.opener`. Web builds `response_type=id_token` redirect URLs by hand, round-trips `state`/`nonce` via `sessionStorage`, and reads the `id_token` back out of the URL hash on return. The nginx `Cross-Origin-Opener-Policy: same-origin-allow-popups` header (`docker-setup.md`) predates this and is now unused by the web path — kept in case native/popup flow returns.

## Known test flake (fixed 2026-08-04) — don't re-derive

`internal/infrastructure/auth/jwt_issuer_test.go`'s `TestJWTIssuer_Parse_RejectsTamperedSignature`
used to tamper a token by replacing its **last** base64url char (`token[:len(token)-1] + "x"`). An
HS256 signature is 32 bytes → 43 base64url chars, and the *last* char only carries 4 significant bits
(its low 2 bits are decoder-ignored padding) — so ~1 in 16 CI runs, the "tampered" signature decoded
to the exact same bytes as the original and `Parse` correctly accepted it, failing the test on a
real backend behavior that was never actually broken. This is what failed `develop` CI
(`Backend (test + vips)`) on 2026-08-04 — `jwt_issuer.go` itself was fine
(`WithValidMethods`/`WithIssuer`/`WithExpirationRequired` all correct). Fixed by flipping the
signature's **first** char instead (6 significant bits, any change alters the decoded bytes).
Verified with `go test -count=300 -run TestJWTIssuer ./internal/infrastructure/auth/`.

## Caching

Read routes emit `ETag` + `Cache-Control: private` + `Vary: Authorization, Accept-Language` (Accept-Language added 2026-08-06, see [[i18n-multi-language]]), honor `If-None-Match` → `304`. (`apps/backend/internal/infrastructure/api/endpoints/cache_headers.go`) `max-age` defaults to `0` since 2026-08-09 (was `30s`) — see [[etag-304-body-loss]] for why.

Backend also has a caching layer for Stand repo + picture storage (ETag/Cache-Control) and background image processing with `pictureStatus` tracking, returning `202 Accepted` on picture upload. The Redis read-through decorator's cache keys include locale (`id:<locale>:<uuid>`, `all:<locale>`, ...) — a write still invalidates the whole namespace, correctly dropping every locale's entries together.

## Locale (i18n)

`GET /stands`, `GET /devil-fruits` (list + detail) resolve `description`/`skills` per `Accept-Language` (or `?lang=` override), falling back `ca-ES → es-ES → en-GB`. Response shape unchanged. Admin-only `GET .../{id}/translations` returns every locale at once; `POST`/`PUT` bodies take a `translations` map keyed by locale instead of flat `description`/`skills` fields (`en-GB` mandatory). `name` is never translated. See [[i18n-multi-language]] for the full decision record.

## JSON shape

camelCase. Errors: `{error, details?[]}`.

## Enums (strings)

- rarity: `COMMON|RARE|EPIC|LEGENDARY`
- stand stats: `E|D|C|B|A|INFINITE|NULL`
- fruitType: `PARAMECIA|ZOAN|LOGIA|SPECIAL_PARAMECIA|ANCIENT_ZOAN|MYTHICAL_ZOAN`
- role: `REGULAR|ADMIN`
- pictureStatus: `NONE|PENDING|READY|FAILED`

## Users (added 2026-08-04)

See [[user-profile-feature]] for the full design. Summary:
- `GET/PATCH /users/me`, `PATCH/DELETE /users/me/picture`, `DELETE /users/me` — id always comes from the JWT claims, never from path/body. `PATCH /users/me` accepts only `{username}`; email/role/completeName in the body → 400 (unknown-field rejection).
- `GET /users/{id}` (any authenticated caller) → `PublicUserResponse` (`id,username,completeName,avatar,avatarThumb`), never email/role.
- Admin-only: `GET /users` (paginated), `PATCH /users/{id}` (username), `PATCH /users/{id}/role`, `DELETE /users/{id}` — self-demotion/self-delete via these routes is blocked (403), and the last remaining admin can't be demoted/deleted (409).

## CORS

Deny-all unless `CORS_ALLOWED_ORIGINS` set explicitly (`config.go:230`). Frontend origin must be added — see [[docker-setup]].

## Websockets

None exist yet. The Gauntlet/Versus game domain ([[gameplay-game-modes]], [[gameplay-domain-design]])
plus its application layer ([[gameplay-application-layer]], added 2026-08-10) are both built and
tested, fully wired in `main.go` (`GameService`, an in-memory game store, a per-game event hub, a
voting-window timer) — but **nothing calls into any of it**: no HTTP routes, no DTOs, no websocket
transport. Voting/loadout assignment is designed assuming websockets, backed by lobby state in
Redis; both are still the explicit next step. Until they land, still do not build real-time features
against this backend.

## Domain entities seen in backend history

DevilFruit (CRUD, filtering, picture handling), Stand (with cache layer), pictures (background processing + status).

Related: [[frontend-stack]]
