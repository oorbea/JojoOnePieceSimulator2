---
name: frontend_test_infra
description: "Frontend (apps/frontend) has jest-expo test infra since 2026-08-04 — jest@29 pinned, two-project split, gotchas documented in vault"
metadata: 
  node_type: memory
  type: project
  originSessionId: 63a84f44-0a74-4e04-9aa1-0adf7f65015d
  modified: 2026-08-04T17:23:21.951Z
---

Frontend (`apps/frontend`) went from **zero tests** to a working jest-expo suite (2026-08-04, same
pass as the [[frontend-responsive-frutiger-aero]]-linked layout fixes — full gotcha list lives in the
repo's ObsidianVault `frontend-stack.md`, not duplicated here). `pnpm test` / `pnpm test:ci` run it;
CI's `Frontend` job now runs typecheck+lint+test.

**Why:** none of the layout bugs found in that pass (dead `gap` in `GlassPanel`, inverted `zIndex`,
over-constrained absolute in `ChannelBar`) were catchable by `tsc`/eslint alone — needed actual
component rendering to fail.

**How to apply:** before adding a frontend test, read `jest.config.js`'s two project definitions
(`logic` = jsdom, non-rendering only; `native` = react-test-renderer, everything that renders) and
pick the matching one — don't default to one without checking. Must-pin `jest@^29` (not `^30`) to
match `jest-expo@57`'s own dependency range, or the very first test run throws
`this._moduleMocker.clearMocksOnScope is not a function`. `@testing-library/react-native`'s `render()`
is async here — always `await renderWithProviders(...)` (`src/test/render.tsx`) or queries fail with a
confusing "render function has not been called" instead of a clear error.
