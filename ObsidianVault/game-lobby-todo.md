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
- **Remaining, not part of "finish the lobbies"**: §5 (optional drag-to-move polish, skip unless
  asked), §6 (in-match UI - separate, larger future tanda).

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

## 5. Frontend: drag-to-move-player (optional polish, not required)

Team switching already works (tap own row to request a move; host taps another player's row
action to move them) - this is pure enhancement, skip unless explicitly requested. If picked up:
`hooks/use-player-drag.ts` using RN-core `PanResponder` (no `react-native-gesture-handler` root
view needed - it's installed but unused, and adding a `GestureHandlerRootView` to
`app-providers.tsx` is a bigger change than this feature needs). Disable when `maxSm` breakpoint or
`useReducedMotion()`. The tap-based path must stay as the accessible primary path regardless.

## 6. In-match UI (rounds, voting, loadouts) - separate, larger tanda

Explicitly out of scope for "finish the lobbies" - this is the *next* feature, not a missing piece
of this one. The socket store (`features/game/stores/game-socket.store.ts`) already has stub
handling for `VOTING_OPENED`/`VOTE_CAST`/`ROUND_RESOLVED`/etc. (see the frame-type switch in
`onmessage`), and `types/game-ws.types.ts` has their payload shapes - both ready for that tanda to
build on, don't redo them. If the owner asks for this, treat it as a fresh planning pass (new
Explore + Plan agents), not a continuation of this checklist - it's a different scale of work
(round timer UI, live vote counts, loadout cards, tiebreak flow, result screen).

## How to verify when done

- Backend: `cd apps/backend && go build ./... && go vet ./... && go test ./...`
- Frontend: `cd apps/frontend && pnpm exec tsc --noEmit && pnpm run test:ci`
- Manual: `local-up` skill to bring up the full stack, then walk through create → edit config as
  host → see it reflected on a second browser session's `STATE` → set a pool restriction → start →
  confirm the match only draws from the restricted pool (existing backend behavior, just needs the
  UI path exercised end to end for the first time).

Related: [[game-lobby-management]], [[game-lobby-frontend]], [[gameplay-game-modes]],
[[frontend-stack]], [[i18n-multi-language]], [[zettelkasten-workflow]].
