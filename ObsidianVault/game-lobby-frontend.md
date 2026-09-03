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
  lobby list), room (`[id].tsx`). Home's channel grid gained an unlocked "Play" tile (~~was
  all-locked placeholders before~~ **Corrected 2026-09-03**: by 2026-09-03 every channel is
  unlocked - see [[catalogo-publico-stands-devil-fruits-stages]]); nav bar gained a `/play` item.
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

## UX pass (2026-08-13) - pool restriction is banlist-only, not whitelist

Owner reviewed the shipped screens and confirmed the power-pool restriction UI (rarity/fruit-type
whitelist chips + banlist, both built in the earlier §4 pass) was confusing - unclear whether a chip
meant "only these" or "not these". Decision: **drop the rarity/fruit-type whitelist chips from the
UI entirely, keep only the banlist.** `PowerPoolFields` (`fields/power-pool-fields.tsx`) is now just a
`FilterDisclosure` wrapping `BanlistField`, with `activeCount` = `poolFilter.banned.length`. The
backend `PoolFilter` type is untouched (`standRarities`/`fruitRarities`/`fruitTypes` still exist and
still get sent as empty arrays) - this was a UI-only simplification, not a backend or type change, so
whitelisting can be reintroduced later without a migration if ever asked for again.

Same pass also fixed: the join-code screen's alphabet copy (removed, `CODE_ALPHABET` in
`lib/game-code.ts` still enforces it silently), manga names no longer translated
(`enums.manga.JOJO` reads "JoJo's Bizarre Adventure" in all three locales - see
[[i18n-multi-language]]'s "names are never translated" rule, which this had violated), the
"Guardar"/"Ninguno" copy bug on the allow-bots toggle (was reusing `common.save`/`common.none` by
mistake - now `game.create.allowBotsOn`/`allowBotsOff`), `SettingRow`/`InfoHint`/tooltip primitives
(see [[norma-tooltips-y-ayuda-contextual]]), and `SpeechBubble`'s missing default padding (browse's
empty-state bubble had no `p`/`gap`, so text touched the edge and got clipped by the panel's
`overflow:hidden`).

## Segundo UX pass, mismo día - "ban by filter" recuperado (no era whitelist)

El owner reportó que quitar el todo el filtro de rareza/tipo-de-fruta/stat de Stand (pase anterior)
perdió una funcionalidad real: poder banear en bloque ("banea todo lo LEGENDARY", "banea todos los
Stand con SPD:A") en vez de buscar poder a poder en `BanlistField`. Distinto de la whitelist que sí
se retiró a propósito - `BanByFilterFields` (`fields/ban-by-filter-fields.tsx`) sigue siendo
**puro baneo**: calcula qué `BannableItem[]` cumplen los criterios elegidos (rarezas, tipos de fruta,
6 stats de Stand vía `GlassSelect`, mismo patrón que el filtro de administración de Stands) y llama
`onBanMatching(ids)`, que añade esos ids al mismo `poolFilter.banned` plano - no toca
`standRarities`/`fruitRarities`/`fruitTypes` del backend en ningún momento. Un stat filter excluye
DevilFruits del match (no tienen stats) y un fruit-type filter excluye Stands (no tienen tipo de
fruta) - misma regla "la dimensión implica el tipo" que ya usaba el resto de esta feature.
`BannableItem` (`fields/banlist-field.tsx`) ganó `rarity`/`fruitType`/`stats` para poder filtrar sin
una segunda query - los containers ya tenían `useStands()`/`useDevilFruits()` completos.

Otros arreglos del mismo pase: tooltips reescritos a `Modal` (ver
[[norma-tooltips-y-ayuda-contextual]]), cola de `SpeechBubble` centrada de verdad, tooltip de
Gauntlet/Versus separado en dos claves (`game.create.help.modeGauntlet`/`modeVersus`, antes
compartían una), `/play/join` ya no muestra el texto "6 caracteres." - el botón simplemente se
deshabilita hasta `isCompleteCode(code)` (ya existía en `lib/game-code.ts`, sin usar hasta ahora), y
el vacío/error de `/play/browse` pasó de `SpeechBubble` a la misma `GlassPanel` plana que usa el
vacío de Stands en admin (`stands-screen.tsx`) - una tarjeta plana con icono se lee mejor ahí que un
bocadillo con cola apuntando a la nada.

## Manga selector moved out of "Lobby settings", stepper label stacked (2026-08-14)

Superseded by a later playtest pass - the manga field described above (inside the config panel,
column-only, no `InfoHint`) moved to the main lobby screen as its own always-visible `MangaRow`
component, and `NumberStepper` gained a `stacked` `SettingRow` layout. Full rationale in
[[game-match-assignment-frontend]]'s "Lobby manga selector moved out of Lobby settings" section -
not re-derived here.

## Config-edit panel hardened (2026-08-31)

Audit-and-close pass on [[game-lobby-todo]]'s §3, which had gone stale describing the config-edit
panel as unbuilt when it had actually shipped. Two things worth keeping as reusable patterns:

- **`requestId`-correlation to attribute a WS error to a specific in-flight command.**
  `useGameSocketStore`'s `send()` already returns the `requestId` it stamps on the outgoing
  `ClientCommand` (see [[game-realtime-transport]]'s protocol section - the backend only ever used
  `requestId` to correlate its own `ERROR` frame back to the command that caused it, but nothing on
  the frontend was reading it back for that purpose before this). `LobbyRoomContainer` now captures
  the `requestId` from its `UPDATE_CONFIG` submit and compares it against `socket.lastError.requestId`
  when an `ERROR` frame lands, so the config panel can show "your edit was rejected" instead of a
  generic toast (kept as well, unconditionally, for every WS error). Any future command with its own
  submit-state UI (saving/saved/error) can reuse the same trick instead of inventing a new one -
  there's nothing config-panel-specific about it.
- **`lib/config-form.ts`** is now the single source of truth for `TEAM_SIZE_LIMITS` (GAUNTLET 1-10,
  VERSUS 1-5), previously duplicated between `lobby-room-container.tsx` and
  `create-lobby-container.tsx`. Also holds `ConfigFormState`, `clampTeamSize`,
  `configFormFromSnapshot`, `applyModeChange`, `buildUpdateConfigPayload` (always builds a full
  replacement payload, never a patch, mirroring `CreateGameRequest`). Both create and edit flows
  should keep pulling from here rather than re-deriving limits locally.

Also fixed while auditing: three host-only field values in `lobby-config-panel.tsx` (reveal speed,
privacy, allow-bots labels) were raw strings outside a `GlowText`, unlike their non-host siblings -
a real RN `Invariant Violation` crash on native, not a style nit. And added `InfoHint` parity with
`create-lobby-screen.tsx` on every config field, reusing the existing `game.create.help.*` keys
(zero new i18n keys). Component tests for this feature existed nowhere before this pass
(`features/game/components/**/__tests__/` didn't exist) - added coverage for `config-form.ts` and
`lobby-config-panel.tsx`. Full writeup and commit hashes in [[game-lobby-todo]]'s §3.

## Power-pool hardening (§4 closed, 2026-09-01)

Closed [[game-lobby-todo]]'s §4 for real - the banlist-only UI (see this note's "UX pass" and
"Segundo UX pass" sections above) had no remaining-count feedback, no shortfall warning, and a
plain `includes()` substring search with no diacritic folding, no result-count copy, and no way to
clear a search or the whole banlist from the field itself.

- **New leaf module** `features/game/lib/pool-stats.ts` - `computePoolCounts` (per-kind
  total/remaining over the catalog minus a Set of banned ids, never negative, ignores stale banned
  ids no longer in the catalog), `poolShortfalls` (mirrors the backend's `checkPoolSufficiency` in
  `game_service.go`: JOJO's selected power manga needs enough Stands, ONE_PIECE's needs enough
  Devil Fruits, only for mangas actually selected - **deliberately stricter than the backend**,
  which checks actual seated occupancy; the UI has no live roster here, so it uses the configured
  `teamSize` instead), and `searchPoolItems` (NFD-normalize + lowercase diacritic folding, every
  whitespace-split query token must match AND order-independent, excludes already-banned ids,
  respects a result limit while reporting the true total). Declares its own structural `PoolItem`
  type instead of importing `BannableItem` from `banlist-field.tsx`, on purpose - keeps this a leaf
  module with no lib→component edge, importable from jest's `logic` project.
- `power-pool-fields.tsx` now renders a warning banner (`GlassPanel` + `AlertTriangle`, `role=alert`
  via `a11yProps`) ABOVE the `FilterDisclosure` so it stays visible while collapsed, plus a
  remaining-count line per selected power manga inside the disclosure body. `FilterDisclosure`
  itself was left untouched (it's shared with Stands/Stages filters).
- `banlist-field.tsx`'s search now goes through `searchPoolItems`, gained a clear-search
  `GlossButton` (circle, `X` icon, shown only while the query is non-empty), a clear-banlist
  `GlossButton` behind a new optional `onClearBanlist` prop (shown only when `editable && banned.
  length > 0`), a per-result kind badge (Stand vs Devil Fruit, since same-named powers exist across
  both catalogs), and result-count/no-matches copy.
- Both `lobby-config-panel.tsx` and `create-lobby-screen.tsx` compute `counts`/`shortfalls` locally
  via `useMemo` straight from `banlistItems`/`poolFilter.banned`/`powerMangas`/`teamSize`, all of
  which they already had - **no container or endpoint changes needed**, confirmed by reading the
  container tree first rather than assuming a new query was required.
- `banlistItems.length > 0` stands in for "catalog known" in both screens (`poolShortfalls`'s
  `catalogKnown` param) - `banlistItems` comes from `useStands()`/`useDevilFruits()` further up the
  tree, so an empty array can mean either "still loading" or "genuinely empty catalog"; treating
  both as unknown avoids a false shortfall alarm while loading, at the cost of skipping the check on
  a truly empty catalog (an edge case this project doesn't otherwise support).
- i18n: 13 new keys under `game.pool.*` in all three locales (`remainingTitle`, `remainingStands`,
  `remainingFruits`, `insufficientStands`, `insufficientFruits`, `banlistClear(Hint)`,
  `banlistClearSearch(Hint)`, `banlistNoMatches`, `banlistResultCount`, `kindStand`,
  `kindDevilFruit`) - no em/en dashes, passes `copy.test.ts`.
- **Gotcha hit while writing `banlist-field.test.tsx`**: `fireEvent.press`/`fireEvent.changeText`
  against this component's `GlossButton`/`GlassField` need `await act(async () => { ... })`, not a
  bare synchronous `act(() => { ... })` - the latter left the query text/state update unflushed by
  the time the next assertion ran (`skills-field.test.tsx` already used the `await act(async...)`
  form for its one `changeText` call; this pass needed it on every interaction, including
  `fireEvent.press`, not just `changeText`).
- Backend: `checkPoolSufficiency`/`beginRound`'s per-round `PoolFilter.Apply` were audited and
  confirmed already mode-agnostic and covering both power kinds - purely a test-coverage gap, no
  production code changed. New
  `apps/backend/internal/application/services/game_service_pool_filter_test.go` covers Gauntlet +
  Versus (per-round reassignment proven across all `VersusRounds`), both exhaustion paths, the
  JoJo-only irrelevant-fruit-exhaustion case, actual-occupancy vs configured-capacity, and the
  `EditLobbyConfig` path the frontend's config-edit panel actually uses. `pool_filter_test.go` also
  gained a banlist-only (no allowlist) `Apply` case banning both a Stand and a DevilFruit in one
  call.

Verified: backend `go build`/`go vet`/`go test ./...` clean (via Docker, host `go test` still
blocked - see [[feedback_backend_tests_via_docker]]); frontend `tsc --noEmit` clean, `pnpm lint` 0
errors (pre-existing CRLF prettier warnings only, environment artifact unrelated to this pass),
`pnpm test:ci` green 48 suites / 511 tests (`--maxWorkers=2` to avoid the documented
`use-loadout-reveal`/`tooltip` full-parallel flake - see `norma-verificacion-docker.md`). Merged to
`develop` as a fast-forward (3 atomic commits: `pool-stats.ts` + tests, UI wiring + i18n, backend
tests).

Related: [[game-lobby-management]], [[gameplay-game-modes]], [[frontend-stack]],
[[frontend-responsive-frutiger-aero]], [[i18n-multi-language]], [[zettelkasten-workflow]],
[[norma-tooltips-y-ayuda-contextual]], [[game-match-assignment-frontend]].
