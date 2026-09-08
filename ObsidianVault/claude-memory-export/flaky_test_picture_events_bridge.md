---
name: flaky-test-picture-events-bridge
description: "picture-events-bridge.web.test.tsx's network-error backoff case intermittently double-counts mintEventsTicket in CI"
metadata: 
  node_type: memory
  type: project
  originSessionId: cb96e461-84f4-4e3b-b275-6321e0513c6b
  modified: 2026-09-03T09:10:28.565Z
---

`src/providers/__tests__/picture-events-bridge.web.test.tsx` test "a network error minting a
ticket backs off and re-mints" occasionally fails in CI with `mockMintEventsTicket` called 2x
instead of 1x right after the initial `render()+flush()`, before any `jest.advanceTimersByTime`.
Seen on PR #38 (2026-09-03), unrelated to that PR's diff (didn't touch this file) — rerunning the
failed job (`gh run rerun <id> --failed`) turned it green with no code change.

**Why:** root cause not yet diagnosed — component's `connect()` effect (see
`picture-events-bridge.tsx`) is single-invoke with no StrictMode in tests, so the double call isn't
explained yet. Added in the `stream-connection-tickets-2026-09-03` tanda ([[stream-connection-tickets-2026-09-03]] in ObsidianVault).

**How to apply:** if this test fails again in CI on a PR that doesn't touch
`picture-events-bridge.tsx`/its test, treat as this known flake — rerun the job first before
digging in. If it starts failing repeatedly (not just once), worth actually root-causing the
timer/microtask race instead of keep-rerunning.
