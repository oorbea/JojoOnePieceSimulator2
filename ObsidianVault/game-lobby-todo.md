---
title: "TODO: finish the game lobby feature"
tags:
  - project
  - jojo-onepiece-simulator
  - todo
  - gameplay
---

# TODO: finish the game lobby feature

**Read this note first** whenever the owner says something like "finish the lobbies", "acaba las
lobbies", "continue the lobby feature", etc. It is the single entry point - it tells you what's
done, what's uncommitted, and exactly what's left, each item with concrete files and patterns so
nothing needs re-deriving. Background/rationale for what's already shipped lives in
[[game-lobby-management]] (backend) and [[game-lobby-frontend]] (frontend) - read those for *why*,
this note for *what's left*.

## 0. Current state (2026-08-13) - DONE, closed out

- §1-§4 all committed on `develop` (commits 5aa8766, 2a82c8c, 8ab0356, a6a3ef5, 1ddead1, 9b88996).
- Backend: `go build ./...`, `go vet ./...`, `go test ./...` clean, including new
  `game_endpoints_test.go`/`game_ws_endpoints_test.go` (§2).
- Frontend: `tsc --noEmit` clean, `pnpm run test:ci` green (34 suites / 305 tests), including
  the config-edit panel (§3) and power-pool restriction UI (§4).
- **2026-08-13 UX pass**: §4's rarity/fruit-type whitelist chips were removed from the UI (owner
  found them confusing, only banning is needed) - see [[game-lobby-frontend]]'s "UX pass" section.
  `PowerPoolFields` is banlist-only now; the backend `PoolFilter` type is untouched. `tsc --noEmit`
  clean, `pnpm run test:ci` green (36 suites / 313 tests) after this pass.
- **Remaining, not part of "finish the lobbies"**: §5 (drag-to-move player - **owner made this
  mandatory 2026-08-28, no longer optional polish**, see §5 below), §6 (in-match UI - separate,
  larger future tanda).

## 1. Commit the existing work

Nothing has been committed. Before adding anything new, stage and commit what's already done and
verified (backend lobby-management commit, frontend lobby-UI commit - or one combined commit, ask
the owner which they prefer if unclear). Do not let new work pile up uncommitted on top of this.

## 2. Backend: HTTP/WS endpoint tests (gap, not forgotten)

**Why it's missing**: `game_endpoints_test.go` and `game_ws_endpoints_test.go` don't exist **at
all**, not even for routes that predate this tanda. Building one needs a local fake set
(`IUserRepository`, `IStageRepository`, `IGamePowerPool`, `IAssignmentWeights`, `ITiebreaker`)
mirroring what `apps/backend/internal/application/services/game_service_test.go` already has in
the `services_test` package - nothing reusable exists in the `endpoints_test` package today.

**What to do**: create `apps/backend/internal/infrastructure/api/endpoints/game_endpoints_test.go`.
- Reuse `idgen.UUIDGenerator[T]{}` (real, `apps/backend/internal/infrastructure/idgen`) for id
  generation instead of a fake - it's already dependency-free and used exactly this way elsewhere.
- Reuse `gamestore.NewMemoryGameStore()` (real, not a fake) as the `ports.IGameStore` - no need to
  fake this either.
- You DO need to write local fakes for `ports.IUserRepository`, `ports.IStageRepository`/
  `ports.IStageCatalog`, `ports.IGamePowerPool`, `ports.IAssignmentWeights`, `ports.ITiebreaker` -
  copy the shape from `game_service_test.go`'s `fakeStageCatalog`/`fakeGamePowerPool`/
  `fakeAssignmentWeights`/`fakeTiebreaker` (same file, `internal/application/services/`), just
  re-declared in the `endpoints_test` package since Go doesn't let you import test-only types
  across packages.
- Follow `stand_endpoints_test.go`'s `fakeTokenIssuer` (already exists in the `endpoints_test`
  package, reusable as-is) and its `doRequest`/`httptest.NewRecorder` pattern.
- Priority coverage (the genuinely new surface from this tanda): `GET /games/public` (200, no
  `code`/`participants` in body), `GET /games/preview?code=` (200 for non-participant, 404 unknown
  code, works for a PRIVATE lobby), `POST /games/{id}/join` (403 `LOBBY_PRIVATE` on a private
  lobby, 409 `LOBBY_LOCKED` on a locked one), `PATCH /games/{id}/config` (host-only, 403 otherwise).
- `game_ws_endpoints_test.go`: at minimum, table-test `dispatch`'s new cases
  (`SWITCH_TEAM`/`MOVE_PLAYER`/`KICK`/`TRANSFER_HOST`/`SET_LOCK`/`UPDATE_CONFIG`) and the
  kicked-participant-closes-own-socket behavior in `forwardEvents`.

## 3. Frontend: config-edit UI (host edits an existing lobby's settings)

**Backend is ready**: `PATCH /games/{id}/config` (REST) and `UPDATE_CONFIG` (WS command) both
exist and are tested (see `apps/backend/internal/application/services/game_service.go`'s
`EditLobbyConfig`, `apps/backend/internal/infrastructure/api/dto/game_ws.go`'s
`UpdateConfigPayload`). **Nothing on the frontend calls either yet.**

**What to do**, all in `apps/frontend/src/features/game/`:
- Add `updateConfig` to `hooks/use-game-commands.ts` sending `CLIENT_COMMAND` type
  `UPDATE_CONFIG` is missing from `types/game-ws.types.ts`'s `CLIENT_COMMAND` object - add it
  there first (mirror the others), then add the sender in `use-game-commands.ts`.
- Build a `lobby-config-panel.tsx` presentational component reusing the exact same field controls
  already built for `create-lobby-screen.tsx` (mode picker, manga toggles, the `Stepper` helper
  currently private to that file - promote it to its own file, e.g.
  `components/presentational/fields/number-stepper.tsx`, and import it from both screens instead of
  duplicating).
- Host-only: editable. Non-host: same layout, read-only (render the values as plain `GlowText`,
  no interactive controls) - `lobby-room-screen.tsx` already receives `you.isHost`, thread it
  through.
- Wire into `lobby-room-container.tsx`: local form state seeded from `snapshot.config`, submit
  calls `commands.updateConfig(...)`. Since `UPDATE_CONFIG`'s payload is a **full replacement**
  (mirrors `CreateGameRequest`), build the whole payload from current + edited fields, don't try to
  send a partial patch.
- Add it to `lobby-room-screen.tsx` between the roster and the `StartBar`, collapsed by default via
  the existing `FilterDisclosure` primitive (`shared/components/presentational/filter-disclosure.tsx`)
  so it doesn't dominate the screen on mobile.
- i18n: reuse the existing `game.create.*` keys where the label is identical (mangas, teamSize,
  votingSeconds, privacy, allowBots) - add a small `game.config.*` set only for what's config-panel
  specific (`title`, `readOnly`, `hostOnly`, `saving`, `saved`) in all three locale files, same
  pattern as every other `game.*` addition (identical keys across `en-GB`/`es-ES`/`ca-ES`, no
  em/en dashes - `copy.test.ts` will catch a violation immediately, it already did once this tanda).

## 4. Frontend: power-pool restriction UI (rarity/fruit-type/banlist)

**Backend is ready**: `PoolFilterPayload` (`apps/backend/internal/infrastructure/api/dto/game_ws.go`)
accepts `standRarities`/`fruitRarities`/`fruitTypes`/`banned` and is already part of
`CreateGameRequest` and `UpdateConfigPayload`. **Frontend types exist**
(`features/game/types/game.types.ts`'s `PoolFilter`) **but no form ever sets a non-empty one.**

**What to do**:
- `fields/power-pool-fields.tsx`: a `FilterDisclosure` with rarity chips (reuse
  `raritySchema`'s values from `shared/lib/zod.ts`: `COMMON|RARE|EPIC|LEGENDARY`) and fruit-type
  chips (`fruitTypeSchema`'s values), same toggle-chip pattern as the manga selector in
  `create-lobby-screen.tsx`.
- `fields/banlist-field.tsx`: a `GlassField` search box over Stands/DevilFruits. **Cross-feature
  import must go through each feature's barrel** (`@/features/stands`, `@/features/devil-fruits`) -
  today those barrels export nothing (`index.ts` is `export {}`), so widen them first to export
  their list hook (`useStands`/`useDevilFruits`) and response type, per
  `src/features/README.md`'s rule.
- Wire into both `create-lobby-screen.tsx`/`create-lobby-container.tsx` and the config-edit panel
  from §3 - same component, same state shape (`PoolFilter`), both screens just hold it in local
  state and pass it through on submit.
- i18n: `game.pool.*` (title, rarities, fruitTypes, banlist, banlistSearch, banlistEmpty, banned,
  removeBan, clearAll) in all three locales.

## 5. Frontend: drag-to-move-player (MANDATORY as of 2026-08-28, not optional anymore)

**Owner decision (2026-08-28)**: this is no longer optional polish - it's required, and it must
work on **both desktop (mouse drag) and mobile (touch drag)**, not just one platform. Do not ship
it desktop-only or behind a breakpoint gate that disables it on mobile.

Team switching already works via tap (tap own row to request a move; host taps another player's
row action to move them) - that tap-based path must stay as the accessible primary path regardless
(keyboard/screen-reader users, [[norma-teclado]]), but drag is now a required *additional*
interaction layered on top, not a nice-to-have.

Implementation note (still valid): `hooks/use-player-drag.ts` using RN-core `PanResponder` covers
both mouse and touch without needing `react-native-gesture-handler` (installed but unused; adding
a `GestureHandlerRootView` to `app-providers.tsx` would be a bigger change than this needs) -
`PanResponder` already unifies mouse-drag-as-pointer and touch-drag on web and native alike, so
"works on both desktop and mobile" doesn't require two separate implementations. Previous guidance
to disable under `maxSm`/`useReducedMotion()` no longer applies to the whole feature - if
`useReducedMotion()` is honored at all, it should suppress the drag *animation/visual feedback*
only, not the interaction itself, since the tap fallback is not a substitute the owner asked for
here.

**Shipped 2026-08-28**: `hooks/use-player-drag.ts` (the `PanResponder` wrapper) + `hooks/use-drop-
zones.ts` (page-coordinate zone registry, re-measures on scroll via `shared/lib/scroll-bus.ts` - the
same pub/sub [[norma-tooltips-y-ayuda-contextual]]'s "Sexto pase" built to stop tooltips sticking on
scroll) + `TeamColumn`/`PlayerRow` wiring. Self can drag to switch team
(mirrors `onJoinTeam`/`switchTeam`); host can drag *any* participant onto another column
(`movePlayer` - net-new, no tap equivalent existed for moving someone else before this). A small
`Move` icon renders on any draggable row as a pure visual affordance; the tap paths
(`onJoin`/host row actions) remain the primary, accessible way to move a player regardless.

**Deliberately avoided `Animated.ValueXY`**: this repo's eslint runs a `react-hooks/refs` rule
(flags "accessing a ref value during render") that rejects RN's `Animated` API outright - it only
works by handing out a `useRef(...).current` instance and reading/returning it outside normal
render data-flow, which the rule can't prove safe (flagged `pan.x`/`pan.y` member access AND
returning `pan` from the hook, even though `const pan = useRef(new Animated.ValueXY()).current`
itself was fine - only *further use* of that value tripped it). Switched to plain `useState` for
the drag translate offset and a `useMemo`'d (not `useRef`'d) `PanResponder.create(...)` instead -
zero refs, so the rule has nothing to flag. No existing code in this repo used RN's `Animated`
before this, which is presumably why nobody hit this until now.

**Verified live (2026-08-28, `local-up` + `claude-in-chrome`)**: `left_click_drag` (this tool's
compound drag action) silently no-ops against RN-Web's `PanResponder` - it doesn't emit the
intermediate `mousemove` events with `buttons:1` the responder system needs to claim the gesture.
Worked once `javascript_tool` dispatched a real `mousedown` → several `mousemove` steps → `mouseup`
sequence via `element.dispatchEvent(new MouseEvent(...))` directly - confirmed self-drag between
`Team A`/`Team B` moves the participant and updates both columns' counts. Touch-event
(`TouchEvent`/`Touch`) dispatch did **not** trigger it, because this was a desktop Chrome profile
with no real touch capability (`'ontouchstart' in window` false) - RN-Web feature-detects and never
attaches touch listeners there, so that's an environment limitation, not a code path difference:
`usePlayerDrag` has exactly one implementation for both platforms, already confirmed working via
the mouse path. Host-dragging *another* participant and a genuine mobile/touch-emulated browser
pass are still unverified live (no second test account, no touch-capable browser profile in this
session) - flagged here rather than silently assumed.

## 6. In-match UI (rounds, voting, loadouts) - split into two tandas

**Start + assignment + reveal - DONE (2026-08-14)**, see [[game-match-assignment-frontend]] for the
full technical note. `lobby-room-screen.tsx` now renders a real match screen (stage banner, roster,
animated sequential loadout reveal) once `snapshot.state` leaves `LOBBY`, on the same `/play/[id]`
route - no new screen/route. `EXPO_PUBLIC_SOCKET_URL` was also finally set (was blank since the WS
transport shipped), see [[docker-setup]].

**Vote buttons + live count + tiebreak revote — DONE (2026-08-26)**, see
[[game-vote-buttons-2026-08-26]] for the full writeup: both backend blockers fixed
(`VOTE_CAST.votesCast`/new `voters` field now carry a real connected-humans-only count;
`votingEndsAt` added to the snapshot mirroring `revealEndsAt`), plus an owner decision found mid-fix
(the revote now clears the ballot instead of carrying every vote over). Frontend: `vote-bar.tsx`
wired to the pre-existing `commands.vote(...)` sender, changing a cast vote confirms via
`ConfirmSheet`, full keyboard nav (new `norma-teclado.md` — roving-tabindex radio group, roster
tiles now Tab-reachable, `1`-`9`/`S` hotkeys). Verified: backend `go test` green, frontend
`typecheck`/`lint`/`test:ci` green (429/429) — **not yet done**: the live two-browser `local-up` +
`claude-in-chrome` walkthrough and a real keyboard-only manual pass, both flagged as next-session
follow-up rather than skipped silently.

**Still open, separate future tanda(s)**:
- Round-resolved feedback (who won, was it a coin flip — nothing renders `ROUND_RESOLVED` beyond
  clearing the countdown) and the final result screen (`GAME_FINISHED` still just toasts and bounces
  to `/play`).
- `LoadoutModal`'s open state isn't wired into the new hotkey `blocked` guard yet (lives inside
  `MatchRoster`, not the container) — small, documented gap, see [[norma-teclado.md]].
- No automated test for `use-roving-group.ts`'s web-only keyboard branch — `hooks/__tests__` always
  runs under jest's native project per the current `jest.config.js` split, where the web branch
  never engages. Needs either a new jsdom hooks lane or relocating this one test.
- `VOTING_OPENED`/`TIEBREAK_OPENED`'s `closesAt` is still transport-synthesized separately from the
  new authoritative `votingEndsAt` (can drift by hub-delivery latency) — clean follow-up, not folded
  in this tanda.
- If the owner asks for the rest of this, treat it as a fresh planning pass (new Explore + Plan
  agents) - it's a different scale of work than what shipped here.

**Roster redesign + sorteo reel fix - DONE (2026-08-17)**, see
[[game-match-assignment-frontend]]'s dated section for the full writeup: the ruleta's landing
frame was always blank (a `justify="center"` geometry bug, now unit-tested against regression),
redesigned with an overshoot landing + fade + stagger; the voting roster now shows only
avatar+username per participant (`ParticipantTile`) with a 1.5s-hover/long-press card and a
tap-opened full breakdown modal; a loadout's Stand/DevilFruit description+skills are now resolved
per viewer locale server-side instead of frozen to en-GB.

## How to verify when done

- Backend: `cd apps/backend && go build ./... && go vet ./... && go test ./...`
- Frontend: `cd apps/frontend && pnpm exec tsc --noEmit && pnpm run test:ci`
- Manual: `local-up` skill to bring up the full stack, then walk through create → edit config as
  host → see it reflected on a second browser session's `STATE` → set a pool restriction → start →
  confirm the match only draws from the restricted pool (existing backend behavior, just needs the
  UI path exercised end to end for the first time).

Related: [[game-lobby-management]], [[game-lobby-frontend]], [[gameplay-game-modes]],
[[frontend-stack]], [[i18n-multi-language]], [[zettelkasten-workflow]].
