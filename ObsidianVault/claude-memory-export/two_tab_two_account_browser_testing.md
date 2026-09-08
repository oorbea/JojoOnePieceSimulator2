---
name: two-tab-two-account-browser-testing
description: "How to keep two real logged-in sessions alive in two tabs of the same Chrome profile for this app, and the mistake that breaks it"
metadata: 
  node_type: memory
  type: project
  originSessionId: 7d1946d5-2c78-4cf8-aa3a-99ec19417412
  modified: 2026-09-02T09:25:51.812Z
---

This app's web auth token lives in `localStorage` (`secure-storage.ts`), shared per-origin across all tabs of the same Chrome profile - two tabs can NOT hold two different tokens in storage at once.

**The trick that works**: have the owner log in tab A first and let it fully load (token now read into memory, WS already open) - only then log in tab B. Tab B's login overwrites the shared `localStorage` value, but tab A never re-reads storage unless it *navigates* again, so both tabs stay independently authenticated and live for the rest of the session.

**The mistake that breaks it**: calling the browser tool's `navigate` action on a tab that already has a session established. `navigate` is always a full page load, which re-reads whatever token currently sits in `localStorage` (the other tab's, if it logged in more recently) - silently swapping which account that tab is acting as. Confirmed happening on this exact repo when `navigate` was used to move a host tab to `/play/create`.

**How to apply**: once both tabs are correctly authenticated, drive all further navigation in either tab via in-app link/button clicks (`find` + `computer left_click` on an in-app nav element) or same-origin SPA routing, never the `navigate` tool - it re-authenticates the tab as whichever account most recently wrote to `localStorage`. If a re-login is ever needed mid-session, it requires the owner to do it again (see [[auth_google_only_no_dev_bypass]]).
