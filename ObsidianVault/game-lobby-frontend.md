---
title: "Feature: lobby UI + WebSocket client (frontend, 2026-08-12)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - feature
  - gameplay
---

# Lobby UI + WebSocket client (2026-08-12)

## Status

**See [[game-lobby-todo]] for the actionable remaining-work checklist.** This note is
background/rationale only.

Shipped a working, tested slice of the lobby feature end to end: create/join/browse/room screens,
a real native-WebSocket client, i18n in all three locales. See [[game-lobby-management]] for the
backend half this wires against. `tsc --noEmit` clean, full jest suite green (34 suites / 301
tests, including 9 new socket-store tests run against a real Redis-free in-memory fake WebSocket).

## What shipped

- New feature `src/features/game/` (types, `api/`, `lib/`, `stores/`, `hooks/`,
  `components/{containers,presentational}`), barrel in `index.ts` per [[frontend-stack]]'s
  convention.
- **Socket store** (`stores/game-socket.store.ts`): a zustand store with the live `WebSocket`
  instance kept as a **module-level ref**, not store state (not serializable, and only one
  connection should exist regardless of mount count) - refcounted the same way the backend's own
  `connRegistry` is. Reconciliation rule enforced in code, not just convention: `STATE` **replaces**
  `snapshot` wholesale; every other frame (`PLAYER_JOINED`, `TEAM_CHANGED`, ...) only pushes to a
  bounded `feed`, never touches `snapshot`. `VOTE_CAST` touches neither. `socketFactory` is
  injectable, which is what makes the store testable with a hand-rolled `FakeWebSocket` in the
  `logic` jest project with zero network and zero jest module-mock of `WebSocket` itself.
- Reconnect: same exponential backoff curve as `picture-events-bridge.tsx`'s SSE (2s→4s→8s...
  capped 30s, extracted to `lib/backoff.ts` so both the store and its test read one definition).
  Token for the socket URL is read **fresh from `useSessionStore.getState()` on every (re)connect**,
  never captured once - same reasoning as the SSE bridge (a token rotated after re-login while
  disconnected must not get baked into a stale reconnect URL).
- `PLAYER_KICKED` targeting self sets a `terminal: {kind:'KICKED'}` the container reacts to by
  cleaning up and routing back to `/play` - the same terminal handling as `GAME_ABORTED`/
  `GAME_FINISHED`.
- Query cache: `gameKeys.detail(id)` is seeded by REST mutations (create/join) and used as the
  fallback query when the socket isn't `open` (`useGameDetail`), but the room screen reads live
  state from the socket store directly, not from TanStack Query - the socket is authoritative while
  connected. `query-provider.tsx`'s `dehydrateOptions.shouldDehydrateQuery` now excludes the
  `'games'` cache segment outright, so a live lobby/game snapshot never gets written to
  AsyncStorage and rehydrated stale on next app launch.
- 5 screens under `app/(app)/play/`: hub, create, join (code input + preview), browse (public
  lobby list), room (`[id].tsx`). Home's channel grid gained an unlocked "Play" tile (was all-locked
  placeholders before); nav bar gained a `/play` item.
- Lobby room: join-code card (copy/share via RN core `Share` + web `navigator.clipboard`) - **no new
  dependency added**, `expo-clipboard` was in the original plan but turned out unnecessary), Versus
  team columns with tap-to-switch (host or self only, mirrors `game.SwitchTeam`'s own rule), Gauntlet
  squad roster, host-only kick/transfer-host icon buttons per row (not a "..." menu - simpler, and
  every action already needs its own `ConfirmSheet`), lock toggle, Start button whose disabled
  reason renders as its own text node under it (`t(gate.reasonKey, gate.params)` from
  `lib/lobby-rules.ts` - mirrors `game.Game.Start`'s checks, first-failing-wins).

## Deliberately cut from this pass (scope call, not forgotten)

- **No config-edit UI in the room** (mangas/team-size/voting-window/pool-filter editing after
  creation) - the backend's `PATCH /games/{id}/config` and `UPDATE_CONFIG` WS command exist and are
  tested, but no frontend form calls them yet. Create-lobby has the full form; editing an existing
  lobby's config is not wired.
- **No power-pool restriction UI** (rarity/fruit-type chips, banlist search) anywhere - `PoolFilter`
  types exist end to end but the create form never sets a non-empty one.
- **No drag-to-move-player** - team switching is tap-only (self) / icon-button (host moving
  someone, via the same tap target). Fully accessible, just not the fancier `PanResponder` drag from
  the original plan.
- **No in-match UI at all** (rounds, voting countdown, loadout cards, tiebreak) - out of scope for
  this tanda by design, but the socket store's frame handling already has stub cases for
  `VOTING_OPENED`/`ROUND_RESOLVED`/etc. ready for that tanda to fill in.
- Join code input is a plain `GlassField`, not the fancier 6-cell UI variant floated in planning -
  functional, reuses an existing primitive, cut for time.

## Gotchas hit

- **`.expo/types/router.d.ts` (expo-router's typed-routes output) is stale and does not regenerate
  from `expo export`** - only a running `expo start` dev server keeps it current. New routes
  (`/play`, `/play/create`, etc.) needed `as never` casts at every `router.navigate`/`router.replace`
  call site, the same escape hatch `app-shell-container.tsx` already used for its own dynamic nav
  array. Not a real type hole - the routes are real and Metro resolves them; opening `expo start`
  once will regenerate the `.d.ts` and these casts become unnecessary (safe to leave, though - it's
  the same pattern already established in the codebase).
- The `copy.test.ts` em/en-dash guard caught this feature's own copy on the first pass
  (`"Public — anyone can..."` in all three locale files) - fixed to a colon. Confirms the guard
  actually works end to end, including on freshly-added catalogs.
- `jest.config.js`'s two-project split needed a matching pair of edits (`logic.testMatch` +
  `native.testPathIgnorePatterns`) to cover `src/features/*/lib/__tests__` and
  `src/features/*/stores/__tests__` - a new pure-logic test under a feature folder would otherwise
  silently run in the `native` project instead (wrong preset, would still probably pass but for the
  wrong reasons, per [[frontend-stack]]'s documented trap).

Related: [[game-lobby-management]], [[gameplay-game-modes]], [[frontend-stack]],
[[frontend-responsive-frutiger-aero]], [[i18n-multi-language]], [[zettelkasten-workflow]].
