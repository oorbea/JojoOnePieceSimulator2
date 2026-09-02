---
title: "Auth: test coverage + doc drift cleanup"
tags:
  - project
  - jojo-onepiece-simulator
  - auth
  - testing
  - decision
---

# Auth — closing the test gap and the doc drift (2026-09-02)

**Zero behaviour changes.** Nothing in the real auth flow was touched: no handler, no middleware,
no service, no config. This tanda added tests around what already existed, and corrected
documentation that was actively lying about it.

## Why this tanda existed

A review of the auth surface found **no exploitable hole**. Google login has been built
end-to-end and verified live in production since [[auth-login-implementation]]:

- Google ID token verified via `google.golang.org/api/idtoken`, with `aud` pinned to
  `GOOGLE_CLIENT_ID`.
- Our own access token is HS256, and `JWTIssuer.Parse` validates the algorithm
  (`jwt.WithValidMethods`, so no `alg=none` downgrade), the issuer and the expiry
  (`jwt.WithExpirationRequired`).
- `config.go` refuses to boot with a `JWT_SECRET` under `minJWTSecretLen` (32 chars).
- `POST /auth/google` has its own tighter per-IP rate-limit tier, separate from the global one.
- `RequireAuth` / `RequireAdmin` gate everything else; `ADMIN_EMAILS` is what promotes a caller.

What *was* missing: tests at the one point that touches Google's library, tests for the middlewares
in isolation, tests for the web state/nonce round trip — and three vault notes plus two READMEs
claiming auth "isn't built yet".

## Tests added

### `apps/backend/internal/infrastructure/auth/google_verifier_test.go`

`GoogleVerifier.Verify` was the only untested file in the package (`jwt_issuer.go` already had
`jwt_issuer_test.go`). Covers: wrong audience, expired token, unsupported algorithms
(`none`/`HS256`/`RS512`/empty), and seven shapes of malformed token — each asserting the error
wraps `ports.ErrInvalidGoogleToken` (which `endpoints/errors.go` maps to 401) and that the returned
`GoogleIdentity` is the zero value.

> [!warning] The limitation, written down so it isn't re-derived
> `Verify` calls the **package-level** `idtoken.Validate`, which reads a package-global
> `defaultValidator`. There is no seam to inject a fake HTTP client — the library's own
> `NewValidator` does accept `option.WithHTTPClient`, but `GoogleVerifier` can't reach it without
> changing its construction, which is a behaviour change this tanda deliberately did not make.
>
> **Consequence:** the happy path needs Google's real signing keys, so it is unreachable offline —
> and with it the `payload.Claims` → `ports.GoogleIdentity` mapping and the `Subject` copy. Faking
> it would mean standing up a fake OIDC/JWKS server or monkey-patching the library, both of which
> test the fake rather than our code, so neither was done.
>
> **What saved the tanda:** reading `validate.go` in `google.golang.org/api@v0.293.0` showed the
> `aud` and `exp` comparisons sit *above* the RS256/ES256 switch that fetches certs. So the
> audience check — the one this backend's security actually rests on — is fully covered, offline,
> with no network call. That is the check worth having.
>
> **Also noticed:** this version of `idtoken` does **not** check the `iss` claim at all. Not a
> hole (the issuer is pinned implicitly by the signature having to verify against Google's own
> JWKS), but it means an "issuer rejected" test here would assert behaviour the library lacks.
> Don't write one and don't "fix" it.
>
> A `stubGoogleVerifier` in the same file writes down the port contract the happy path must
> satisfy — in particular that `EmailVerified` is *reported* by the verifier and *enforced* by
> `AuthService`, so a verifier normalising it to `true` would silently break that check. It is
> labelled a specification test, not coverage of `google_verifier.go`.

### `apps/backend/internal/infrastructure/api/endpoints/auth_endpoints_test.go`

Follows the package's existing `*_endpoints_test.go` style (fakes over mocks, real service, real
`NewRouter`), reusing `fakeUserRepo` / `fakePictureStorage` and adding a `fakeGoogleVerifier` plus a
`loginTokenIssuer` that can actually `Issue` (the package's shared `fakeTokenIssuer` can't). Covers:

- **The 200-vs-201 split** — first login registers (201) and persists the user, second login of the
  same Google account is 200 and creates no duplicate. The frontend branches on this.
- `ADMIN_EMAILS` promoting on the *first* login too, matched case-insensitively with surrounding
  whitespace trimmed.
- **DTO validation** — nine bad bodies (missing/empty/null `idToken`, non-JSON, empty, wrong type,
  unknown field, array, truncated) all 400, each also asserting the verifier was called **zero**
  times: a rejected body must never reach the service.
- Rejected token → 401 with the generic `unauthenticated` body (no leak of *why*) and nobody
  registered; unverified email → 400 and nobody registered.
- That `POST /auth/google` still answers without a bearer token — a regression putting it behind
  `RequireAuth` would lock everyone out.

### `apps/backend/internal/infrastructure/api/endpoints/middleware_test.go`

`RequireAuth`/`RequireAdmin` were only ever exercised *through* other endpoints. Now driven
directly: ten unusable `Authorization` headers → 401 (including `bearer` lowercase,
`Beareruser-token` with no space, and `Bearer ` with nothing after it); a genuinely expired token
minted by a real `JWTIssuer` with a negative TTL and parsed by a live one → 401; valid token →
passes with the right `UserID`/`Role` on the context; `RequireAdmin` → 403 for `REGULAR`, passes
for `ADMIN`, and **fails closed** when mounted without `RequireAuth` in front of it (no claims on
the context must mean 403, not a free pass).

### `apps/frontend/src/features/auth/hooks/__tests__/use-google-auth.web.test.ts`

The `.web.test.ts` suffix routes this to the jsdom **logic** project (see [[frontend-stack]] and
`jest.config.js`) — it needs a real `document`, `sessionStorage`, `atob` and
`Platform.OS === 'web'`. `buildWebAuthUrl`, `decodeJwtPayload` and the redirect effect are all
unexported, so everything is driven through the hook via the same probe shape
`use-roving-group.web.test.tsx` established.

Covers the authorize URL (client id, `redirect_uri` being origin+pathname only, `response_type`,
`scope`, `prompt`, and state/nonce matching what was stored), the full redirect return (state and
nonce both matching → `postGoogleAuth` + `setSession`, hash cleared from history, state/nonce
burned so a replay can't reuse them), and every rejection path: mismatched state, absent state,
**no stored state at all** (what a pasted id_token looks like), mismatched nonce, and a token with
no nonce claim. Plus a table driving `decodeJwtPayload` across payload lengths 0..3 mod 4 and a
payload whose base64 contains `+`/`/`, since the decoder re-pads and swaps by hand.

> [!tip] jsdom gotchas worth keeping
> - `window.location` is an own **accessor** property on jsdom's window and is *configurable*
>   (jsdom's `utils.define` copies object-literal getter descriptors verbatim), so it can be
>   swapped for a plain object via `Object.defineProperty`. That is the only way to both read
>   `origin`/`pathname` deterministically and observe `assign()`: jsdom's real `Location` refuses
>   cross-origin navigation and cannot be spied on. Same trick works for `history` and `crypto`.
> - Every name a `jest.mock` factory closes over must start with `mock` —
>   `babel-plugin-jest-hoist` lifts the call above the imports and rejects any other out-of-scope
>   reference.
> - The hook defers its redirect work into an async IIFE that awaits further promises, so the flush
>   helper yields to a real `setTimeout(0)` inside `act`, not a single `await Promise.resolve()`
>   (which only drains one link of the chain per call).

## Doc drift fixed (before → after)

| Where | Before | After |
|---|---|---|
| [[overview]] "Flagged / not done" | "Google OAuth client IDs blank in `.env.example` — auth feature not built yet." | Struck through and corrected. Wrong on **both** halves: auth is built and prod-verified, and the blank ids are deliberate (both `.env.example` headers say so; real values live in GitHub secrets). |
| [[ADR]] "Known weak spots" | "Auth/session flow is incomplete — Google login exists, rest of the session/auth lifecycle is WIP." | Struck through and corrected, with the three *accepted* risks named so they aren't re-read as missing work. |
| `claude-memory-export/project_context.md` | "Auth/session flow incomplete — Google login works, rest of session lifecycle is WIP." | Struck through and corrected: what was missing was test coverage and accurate docs, not the feature. |
| [[picture-events-sse]] | "...given `JWT_TTL=8h`." | **24h.** `deployments/.env.example` ships `JWT_TTL=24h`, `config.go`'s `defaultJWTTTL` is `24 * time.Hour`, and [[backend-contract]] said 24h all along. The **note** was wrong, the value was not — `.env.example` was left untouched. |
| `apps/backend/README.md` | Pointed at `.env.example` (read as `apps/backend/.env.example`, which does not exist). | Points at `deployments/.env.example`, and says explicitly that the whole stack shares one. |
| `deployments/README.md` | Listed only CD-only credentials (Tailscale/SSH). | Adds a table for `GOOGLE_CLIENT_ID` / `JWT_SECRET` / `ADMIN_EMAILS` — the ≥32-char floor, the "must equal `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID`" trap, and that an empty `ADMIN_EMAILS` makes the admin UI unreachable. |

## CI placeholder raised

`.github/workflows/ci.yml`'s `compose-config` job set `JWT_SECRET: ci-placeholder` — **15
characters**, under `config.go`'s `minJWTSecretLen` of 32. Harmless today (the job only runs
`docker compose config`, which never boots the app) but a trap the day that job starts anything for
real. Raised to `ci-placeholder-not-a-real-secret-32chars` (40 chars, obviously fake) with a
comment saying why the length matters.

## The three accepted risks — re-confirmed, deliberately unchanged

None of these were touched, and none should be "fixed" by a future pass that rediscovers them
without reading the reasoning first.

1. **JWT in the query string for SSE/WS.** `EventSource` cannot set custom headers, full stop.
   `?token=` is scoped to a helper local to `events_endpoints.go` — the shared `RequireAuth`
   middleware never gained query-param auth, so no other route is affected. The game WebSocket does
   the same and tries `Authorization: Bearer` first. Reasoning in [[picture-events-sse]] and
   [[game-realtime-transport]]. The ticket-minting follow-up is noted there, not built.
2. **Session in `localStorage` on web.** `expo-secure-store` throws on web, so
   `src/shared/lib/secure-storage.ts` branches to `localStorage` — see [[frontend-stack]]. Also the
   reason the two-tab two-account testing trick works at all
   ([[game-round-result-live-walkthrough-2026-09-02]]).
3. **No refresh tokens.** 24h JWT, no cookies; a 401 anywhere forces a re-login with no silent
   refresh. Stated in [[backend-contract]]. Rotating `JWT_SECRET` therefore logs everyone out —
   now also called out in `deployments/README.md`.

## Verification

Backend: `go vet ./internal/infrastructure/auth/... ./internal/infrastructure/api/endpoints/...`
clean (compiles both new test files). Full `go test` was **not** run here — the Docker stack was
shared with two other tandas running in parallel and the project norm
([[norma-verificacion-docker]]) forbids host-native `go test` on this machine anyway. Frontend
tests were not run for the same reason plus no `node_modules` in this worktree. Centralised
verification was left to the orchestrator.

Related: [[auth-login-implementation]], [[backend-contract]], [[frontend-stack]],
[[picture-events-sse]], [[game-realtime-transport]], [[cicd-deployment]]
