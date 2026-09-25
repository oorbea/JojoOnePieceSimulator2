---
title: "Dev login bypass (2026-09-25)"
tags:
  - project
  - jojo-onepiece-simulator
  - auth
  - dev-workflow
---

# Dev login bypass (local-only)

Closes the gap flagged in [[auth_google_only_no_dev_bypass]]: local testing needed several
accounts logged in at once without going through Google each time, and without the old
two-tab trick (dead since [[session-token-storage-2026-09-05]] moved the refresh token into
a shared, browser-wide cookie).

## What shipped

- `POST /api/v1/auth/dev-login` (`AuthEndpoints.devLogin`) — only mounted when
  `AuthEndpoints.SetDevAuthBypass(true)` was called (`cmd/app/main.go`, from
  `cfg.DevAuthBypass`, env `DEV_AUTH_BYPASS`). Never Google, never a real credential:
  `AuthService.LoginDev(name, admin)` builds a synthetic `ports.GoogleIdentity`
  (`Subject: "dev:"+name`, `Email: name+"@dev.invalid"`) and reuses the same
  `findOrRegister`/`syncExisting` path a real login uses, so a repeat call with the same
  name re-authenticates the same user and can flip its role via the explicit `admin` flag.
- `.invalid` is RFC 2606-reserved, never resolvable — a dev account can never collide with a
  real Google account.
- `AuthService.Refresh` skips the `ADMIN_EMAILS` resync for `@dev.invalid` emails (they're
  never listed there) and instead preserves whatever `LoginDev` last set — otherwise every
  silent refresh would silently demote a dev-admin back to Regular.

## Defense in depth against ever reaching prod

Three independent layers, all in `apps/backend`:

1. **Boot guard** (`config.Load`): refuses to start if `DEV_AUTH_BYPASS=true` alongside a
   prod-shaped `AUTH_COOKIE_SECURE=true` (the default!) or a non-`localhost`/`127.0.0.1`
   `CORS_ALLOWED_ORIGINS` entry.
2. **`requireLocalRequest`** (`endpoints/local_only.go`): gates the route itself — TCP peer
   (`r.RemoteAddr`, deliberately not the XFF-aware `middleware.GetClientIP`) must be loopback
   or private-range, **and** the request must carry none of `X-Forwarded-For`, `Forwarded`,
   `X-Real-IP`, `CF-Connecting-IP`. Prod's NPM always appends to XFF (see
   [[csp-y-rate-limit-por-ip-2026-09-05]]'s `keyByClientIP` doc), so this is unreachable
   there. A failed check 404s, not 403s — doesn't reveal the route exists.
3. **Only ever set in `docker-compose.dev.yml`**: `DEV_AUTH_BYPASS`/`EXPO_PUBLIC_DEV_AUTH`
   never appear in the base compose file, `docker-compose.prod.yml`, or `.env.example`'s
   active lines — only as a commented-out, loudly-labelled example.

## Frontend: one session per tab, not per browser

The Google flow's refresh token lives in an HttpOnly cookie shared by the whole browser
profile — fine for one real account, useless for several dev accounts at once. Dev-login
instead:

- Never sets that cookie (`devLogin` skips `setRefreshCookie`, always returns the refresh
  token in the response body).
- Frontend keeps it in `sessionStorage` (`shared/api/dev-refresh-token.ts`, key
  `jops.dev_rt`) — scoped to one tab, so each tab can be a different dev account.
- `refresh.ts`'s `doRefresh` checks for a dev token first (web only) before falling back to
  the cookie-based flow — the real Google path is completely untouched.
- `query-provider.tsx`'s persisted-cache `buster` gets a `:userId` suffix for `@dev.invalid`
  sessions, so one dev tab's persisted React Query snapshot never rehydrates into a
  different dev tab's account (real accounts don't need this — one cookie, one user).
- Hidden route `app/dev-login.tsx`: redirects straight to `/login` unless
  `EXPO_PUBLIC_DEV_AUTH` was baked into the build (`docker-compose.dev.yml`'s build arg,
  default `"true"`). Never linked from anywhere in the app.
- UI (`DevLoginScreen`): name field + Regular/Admin role toggle (two `GlossButton`s as a
  radio group, `a11yRole="radio"`/`a11yChecked`) + recent-accounts quick-pick
  (`localStorage`, `dev-login-recent-accounts.ts`) + a link back to the real login.

## Gotchas hit during the live walkthrough

- **`Dockerfile.frontend` needs its own `ARG`/`ENV` per `EXPO_PUBLIC_*` var.** Adding
  `EXPO_PUBLIC_DEV_AUTH` to `docker-compose.dev.yml`'s `build.args` alone did nothing — the
  Dockerfile only forwards the build args it explicitly declares with `ARG ...` +
  `ENV ...=${...}`. Forgot this once; `/dev-login` kept redirecting to `/login` even with the
  compose arg resolved correctly (`docker compose config` showed it fine — the image build
  just never received it).
- **`CORS_ALLOWED_HEADERS` needs `X-Refresh-Token`/`X-Refresh-Token-Transport` for the web
  dev flow.** Both headers were native-only before this (no browser CORS preflight to
  satisfy on native). The symptom was brutal to diagnose: the browser's preflight `OPTIONS`
  returns `200` regardless, chi's cors middleware just silently omits every
  `Access-Control-Allow-*` header when a *requested* header isn't in the allow-list, so the
  actual `POST` never even reaches the Go backend (**zero log lines**) and Chrome reports a
  bare `TypeError: Failed to fetch` (surfaced upstream as a synthetic `503` in some tooling).
  Confirmed root cause by comparing an `OPTIONS` preflight to the already-working
  `/auth/google` route side by side, and by testing a raw `fetch()` from the page console
  with/without the custom header. Fixed in `config.go`'s `defaultCORSAllowedHeaders` **and**
  `deployments/.env.example` — but a pre-existing local `.env` (gitignored) still had the old
  3-header list and silently overrode the new default; `docker compose restart` does **not**
  reread `env_file` (only `up`/`recreate` does).
- **`FRONTEND_PORT` is `8081` on this machine's local `.env`, not the `3000` default** — see
  [[local_env_points_to_prod]]. `http://localhost:3000/dev-login` 404s outright; the
  running frontend is always on whatever `docker compose ps` actually shows.
- **Claude-in-Chrome: `find` + click-by-`ref` did not focus/type into the Tamagui
  `GlassField` input reliably** (typed text landed nowhere, field stayed empty, form
  submitted blank). A plain coordinate `left_click` at the field's on-screen position, read
  from a fresh `screenshot`, worked every time. Prefer coordinate clicks over element `ref`s
  for this design system's text inputs until this is root-caused.
- Not a Claude-in-Chrome safety guardrail — worth remembering next time a form submit
  mysteriously "fails" with no visible cause: check the actual CORS preflight
  (`Access-Control-Allow-Headers` on the `OPTIONS` response) before assuming automation was
  blocked.

## Verified live (Chrome, two tabs, 2026-09-25)

- Tab A `alice` (Regular) + Tab B `bob` (Admin) logged in simultaneously, same browser.
- `bob` gets the "Admin" nav item, `alice` doesn't.
- Full page reload on either tab keeps that tab's own session (sessionStorage survives
  reload, not just navigation).
- Logging `alice` out (`clearSession` → `/login`) left `bob` completely unaffected after
  reload — no shared-cookie bleed.
- Negative checks via `curl`: loopback → `201`; forged `X-Forwarded-For` → `404`; backend
  refuses to boot with `DEV_AUTH_BYPASS=true` + `AUTH_COOKIE_SECURE=true`.

## Still open / known limitation

Duplicating a tab copies its `sessionStorage`, so both copies would race to rotate the same
refresh token — the replay gets rejected and kills the whole family (both copies logged
out). Open a **new** tab at `/dev-login` for another account instead of duplicating one.

Related: [[auth_google_only_no_dev_bypass]] (superseded by this), [[auth-hardening-2026-09-02]],
[[session-token-storage-2026-09-05]], [[docker-setup]], [[backend-contract]]
