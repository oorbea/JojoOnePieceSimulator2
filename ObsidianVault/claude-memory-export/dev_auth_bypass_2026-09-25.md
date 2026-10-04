---
name: dev-auth-bypass-2026-09-25
description: "Local-only dev login shipped (POST /auth/dev-login) - several .invalid-domain test accounts, one per browser tab, defense-in-depth against prod; two gotchas that block Claude-in-Chrome testing until fixed"
metadata:
  node_type: memory
  type: project
  modified: 2026-09-25T07:54:38.118Z
  originSessionId: e2708676-617b-4fbf-96c5-24d45bc8e3d6
---

`POST /api/v1/auth/dev-login` bypasses Google entirely for local dev: synthetic `<name>@dev.invalid` accounts, `admin` flag sets role explicitly, reuses the real login's find-or-register path. Gated by `DEV_AUTH_BYPASS` (only in `docker-compose.dev.yml`) + `requireLocalRequest` (loopback/private, no proxy headers) + a `config.Load` boot guard that refuses to start if the flag is on alongside a prod-shaped `AUTH_COOKIE_SECURE`/`CORS_ALLOWED_ORIGINS`. Frontend keeps the dev refresh token in `sessionStorage` (per-tab), not the shared cookie, so several dev accounts are live at once in separate tabs of the same browser. Hidden route `/dev-login`, only rendered when `EXPO_PUBLIC_DEV_AUTH` was baked into the build. Full detail: [[dev_auth_bypass_2026-09-25]] in the vault (same filename, `ObsidianVault/dev-auth-bypass-2026-09-25.md`).

**Why this matters**: replaces [[auth_google_only_no_dev_bypass]]/[[two_tab_two_account_browser_testing]] as the way to get 2+ independent live sessions for testing - no more asking the owner to log in real Google accounts by hand.

**Two things that block Claude-in-Chrome from testing this app's forms/auth autonomously, both hit and fixed 2026-09-25**:

1. **`CORS_ALLOWED_HEADERS` must list every custom header a web request sends**, not just the ones already proven to work. Missing one makes the browser's preflight `OPTIONS` come back `200` with **zero** `Access-Control-Allow-*` headers (chi's cors middleware silently omits them, doesn't error) - the real request then never reaches the Go backend at all (zero log lines) and Chrome just reports `TypeError: Failed to fetch` (looked like a synthetic `503` upstream, and briefly looked like a Claude-in-Chrome safety block on an auth-shaped POST - it wasn't). Diagnose by comparing the failing route's `OPTIONS` preflight response to a working route's side by side, or by running `fetch()` directly from the page console via `javascript_tool` with/without the suspect header.
2. **A local, gitignored `deployments/.env` can silently shadow a config default you just changed** - `CORS_ALLOWED_HEADERS` (and anything else) set there overrides `config.go`'s default, and `docker compose restart` does **not** reread `env_file`; only `up`/recreate does. Always `up -d <service>` after editing `.env`, never just `restart`.
3. **Claude-in-Chrome's `find` + click-by-`ref` did not reliably focus/type into this app's Tamagui `GlassField` inputs** - typed text landed nowhere, form submitted blank. A plain coordinate `left_click` at the field's on-screen position (read from a fresh `screenshot`) worked every time. Not root-caused; just use coordinates for this design system's text inputs.

**How to apply**: when testing any form/auth flow on this repo with Claude-in-Chrome, use coordinate clicks for text inputs, and if a POST silently "fails" with no server log at all, check the CORS preflight response headers before suspecting anything else (automation block, network issue, etc.).
