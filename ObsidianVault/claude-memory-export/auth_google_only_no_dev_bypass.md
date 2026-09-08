---
name: auth-google-only-no-dev-bypass
description: "Auth is real Google OAuth only, no dev-mode bypass wired anywhere runnable - two-account live testing needs the owner to log in both accounts manually"
metadata: 
  node_type: memory
  type: project
  originSessionId: 7d1946d5-2c78-4cf8-aa3a-99ec19417412
  modified: 2026-09-02T09:25:42.081Z
---

`POST /auth/google` is the only login route and `main.go` always wires the real `GoogleVerifier` - the `fakeGoogleVerifier` used in `auth_service_test.go` is never constructed outside tests. There is no `AUTH_DEV_BYPASS`-style env var or similar.

**Why this matters**: any live multi-account test (two-browser walkthroughs, etc.) needs the owner to actually log in with two real Google accounts - Claude cannot create or use test accounts alone (entering third-party credentials is out of bounds), and cannot fake a token server-side without a code change.

**How to apply**: when a task needs 2+ distinct live sessions, ask the owner up front to log in each browser tab themselves rather than trying to work around it (e.g. via localStorage tricks alone - see [[two_tab_two_account_browser_testing]] for what that trick can and can't do). If this becomes a recurring need, the concrete fix (not yet built, owner explicitly deferred it once) is a dev-only env var wiring a fake verifier for `local-up`, never for prod.
