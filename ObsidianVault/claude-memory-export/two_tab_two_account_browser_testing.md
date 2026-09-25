---
name: two-tab-two-account-browser-testing
description: "STALE (pre session-token-storage-2026-09-05) plus SUPERSEDED 2026-09-25 - real accounts still share one refresh cookie per profile, but dev-login accounts (dev-auth-bypass-2026-09-25) each get a real independent per-tab session"
metadata: 
  node_type: memory
  type: project
  originSessionId: 7d1946d5-2c78-4cf8-aa3a-99ec19417412
  modified: 2026-09-25T09:00:00.000Z
---

**Stale as of [[session-token-storage-2026-09-05]]**: the access token moved out of `localStorage` into memory-only, backed by a rotating HttpOnly refresh cookie (web) - the mechanism below (`localStorage` token swap) no longer applies to real Google logins. That cookie is still shared per Chrome profile, so two *real* accounts still can't be independently live in two tabs the way this note originally described.

**Superseded for the multi-account use case, 2026-09-25**: for local dev, `/dev-login` ([[dev-auth-bypass-2026-09-25]]) now gives each tab a genuinely independent session via `sessionStorage` (not `localStorage`) - no swap risk, `navigate` is safe to use again on a dev-login tab. Use that instead of this trick whenever real Google accounts aren't specifically required.

Kept for history only, below is the pre-2026-09-05 state (real accounts, `localStorage` token):

This app's web auth token lived in `localStorage` (`secure-storage.ts`), shared per-origin across all tabs of the same Chrome profile - two tabs could NOT hold two different tokens in storage at once.

**The trick that works**: have the owner log in tab A first and let it fully load (token now read into memory, WS already open) - only then log in tab B. Tab B's login overwrites the shared `localStorage` value, but tab A never re-reads storage unless it *navigates* again, so both tabs stay independently authenticated and live for the rest of the session.

**The mistake that breaks it**: calling the browser tool's `navigate` action on a tab that already has a session established. `navigate` is always a full page load, which re-reads whatever token currently sits in `localStorage` (the other tab's, if it logged in more recently) - silently swapping which account that tab is acting as. Confirmed happening on this exact repo when `navigate` was used to move a host tab to `/play/create`.

**How to apply**: once both tabs are correctly authenticated, drive all further navigation in either tab via in-app link/button clicks (`find` + `computer left_click` on an in-app nav element) or same-origin SPA routing, never the `navigate` tool - it re-authenticates the tab as whichever account most recently wrote to `localStorage`. If a re-login is ever needed mid-session, it requires the owner to do it again (see [[auth_google_only_no_dev_bypass]]).
