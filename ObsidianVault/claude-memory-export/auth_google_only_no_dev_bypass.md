---
name: auth-google-only-no-dev-bypass
description: "SUPERSEDED 2026-09-25 - a real local-only dev-login bypass now exists, see dev_auth_bypass_2026-09-25"
metadata: 
  node_type: memory
  type: project
  originSessionId: 7d1946d5-2c78-4cf8-aa3a-99ec19417412
  modified: 2026-09-25T09:00:00.000Z
---

**Superseded 2026-09-25** — the deferred fix this note called for is now built: `POST /api/v1/auth/dev-login`, gated behind `DEV_AUTH_BYPASS` (only ever set by `docker-compose.dev.yml`), `requireLocalRequest`, and a boot guard in `config.Load` that refuses to start if the flag is on alongside a prod-shaped `AUTH_COOKIE_SECURE`/`CORS_ALLOWED_ORIGINS`. `POST /auth/google` and the real `GoogleVerifier` wiring are untouched. Full writeup: [[dev-auth-bypass-2026-09-25]].

Kept for history only, below is the pre-2026-09-25 state:

`POST /auth/google` is the only login route and `main.go` always wires the real `GoogleVerifier` - the `fakeGoogleVerifier` used in `auth_service_test.go` is never constructed outside tests. There is no `AUTH_DEV_BYPASS`-style env var or similar.

**Why this mattered**: any live multi-account test (two-browser walkthroughs, etc.) needed the owner to actually log in with two real Google accounts. Not true anymore for local dev - see [[dev-auth-bypass-2026-09-25]] for the live-verified two-tab walkthrough using `/dev-login` instead.
