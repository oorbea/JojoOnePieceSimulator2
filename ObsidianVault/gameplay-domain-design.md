---
title: "Feature: game domain layer design (2026-08-10)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - feature
  - gameplay
---

# Gameplay domain layer design (2026-08-10)

For the actual gameplay rules (what this code implements), see [[gameplay-game-modes]]. This note
is the technical map: patterns, files, and what's deliberately deferred.

## Why

`internal/domain/entities/game/` and `internal/domain/entities/user/player.go` were dead-code
skeletons — a `Game{gameMode *IGameMode, ...}` with a pointer-to-interface, an unimplemented
`IGameMode`, a `Team.Players()` returning `*map[...]` (mutable state leak), and a `Player` with 9
raw `byte` stats, no constructor, no validation, referenced nowhere. Rebuilt as the real domain
layer for Gauntlet/Versus, per the owner's explicit request to use known patterns (State,
Strategy, Template Method) and to keep this pass domain-only — no Redis, no websockets, no
endpoints, no migrations.

## Patterns

- **State** — `game.Game` over `enums.GameState`
  (`LOBBY → ASSIGNING → VOTING → [TIEBREAK] → RESOLVING → ASSIGNING → ... → FINISHED`, with
  `ABORTED` reachable from anywhere). Every transition method validates the current state and
  returns `ErrInvalidStateTransition` otherwise.
- **Strategy** — `game.IGameMode`, implemented by `GauntletMode` and `VersusMode`. `Game` never
  branches on `enums.GameModeKind` itself; every mode-specific decision (ballot options, whether
  Loadouts reassign each round, which Stage backs a round, how a resolved round affects the game,
  the final outcome) is delegated. Both mode structs are stateless — they read/derive everything
  from the `*Game` passed in, so there's no per-mode state to keep in sync with `Game` itself.
- **Template Method** — `game.LoadoutBuilder.Build`: the step *order* is fixed (**changed
  2026-08-14, owner request**: Physical Form → Stand → Devil Fruit → Fruit Mastery → Hamon →
  Armament Haki → Observation Haki → Conqueror Haki → Spin → apply invariant rules — was Stand →
  Spin → Hamon → Devil Fruit → Fruit Mastery → the three Hakis → Physical Form; see
  `TestLoadoutBuilder_DrawOrder` in `loadout_builder_test.go` for the pinned sequence, a
  `recordingRandom` fingerprinting the exact `IntN(n)` calls made so an accidental reordering fails
  loudly), and which steps run at all is fixed by the selected manga(s) — Physical Form stays
  One Piece-only, just moved first. Each concrete draw is a
  weighted random pick (`RandomSource` + `AssignmentWeights`, both swappable). `NewLoadout`
  re-validates the hard invariants independently at the end, so the template method's correctness
  doesn't rely on getting the draw order right — a second implementation of ability sourcing
  (inventory-based Versus, eventually) gets the same guarantees for free.
- **Value objects** — `Loadout`, `Stage`, `Config`, `AssignmentWeights`, `GameResult` are all
  immutable-after-construction with validating constructors, following the existing
  `powers.NewPower`/`user.NewUser` convention (`(*T, error)`, sentinel errors for real business
  rules, ad-hoc `errors.New` for simple "required" guards).
- **Domain events** — `Game.PullEvents()` drains a `[]DomainEvent` (`PlayerJoined`, `HostReassigned`,
  `VotingOpened`, `RoundResolved`, `GameFinished`, ...) accumulated since the last call. Mirrors
  `services.PictureEventHub`'s pub/sub role ([[picture-events-sse]]) but sourced from the aggregate
  itself rather than a background worker — the application layer will drain these to publish over
  websockets and to feed the (not yet built) game history.

## Avoiding an import cycle

`ports` already depends on `entities/*` (e.g. `IStandRepository` references `powers.Stand`). If
`entities/game` needed to import `ports` too (for `RandomGenerator`, a tiebreaker, a stage
catalog), that would be a cycle the moment any port referenced a `game` type. Resolved by keeping
`entities/game` fully port-free:

- `game.RandomSource` is a **local, minimal interface** (`IntN(n int) int`) — not
  `ports.RandomGenerator[T]`. Any type satisfying the ports interface for any `T` (e.g.
  `infrastructure/random.StdRandomGenerator[T]`) already satisfies `game.RandomSource`
  structurally, no import needed.
- Coin-flip/LLM tiebreaking is **not called from inside the domain**. `Game.CloseVoting` reports
  `tied=true` and stops; the application layer calls `ports.ITiebreaker.Break(ctx, []string)`
  (plain strings, no `game` import needed on that port either) and feeds the winner back in via
  `Game.ResolveTiebreak(winner)`. The domain does the pure logic (state transition, recording the
  round), the port does the IO/policy.
- `AssignmentWeights` is a **plain value struct in `game`**, not a port. `ports.IAssignmentWeights`
  (which *does* import `game`, one-directionally — fine) resolves and returns one; the application
  layer passes the resolved value into `LoadoutBuilder`.

Net effect: `entities/game` imports only `enums`, `valueobjects`, `entities/powers`,
`entities/user` — verified with `go list -deps ./internal/domain/entities/game/... | grep -E
"infrastructure|application|ports"` (empty).

## File map

All under `apps/backend/internal/domain/entities/game/` unless noted.

- `game.go` — the `Game` aggregate/state machine.
- `game_mode.go`, `gauntlet_mode.go`, `versus_mode.go` — the Strategy.
- `loadout.go`, `loadout_builder.go`, `loadout_evaluator.go`, `weights.go` — ability assignment +
  bot voting heuristic (`DefaultLoadoutEvaluator`/`BotVoter`).
- `power_pool.go` — `AvailablePowers`, the per-team draw-without-replacement view over the catalog.
- `power_traits.go` — `TraitsOf`/`HasTrait`, the `REQUIRES_SPIN_4` stopgap (see below).
- `ballot.go` — the shared vote-tally algorithm both modes map their options onto.
- `participant.go`, `team.go`, `config.go`, `stage.go`, `round.go`, `game_result.go`, `events.go`,
  `errors.go` — supporting entities/value objects.
- `*_id.go` — `GameID`/`ParticipantID`/`TeamID`/`StageID`, each following `powers.PowerID`'s
  `[16]byte` + `valueobjects.Format/Parse/IsNil` convention exactly.
- New enums in `internal/domain/enums/`: `Manga`, `GameModeKind`, `AbilitySource`, `GameState`,
  `ParticipantKind`, `SpinLevel`, `HamonLevel`, `FruitMastery`, `HakiLevel`, `PhysicalForm`,
  `SquadVerdict`, `PowerTrait` — every one follows `StandStat`'s shape (`byte` + `iota` +
  `String()`/`IsValid()`/`ParseX()` + a sentinel `ErrInvalidX`).
- New ports in `internal/domain/ports/`: `IStageCatalog`, `IGamePowerPool`, `IAssignmentWeights`,
  `ITiebreaker`, `IGameHistory`, `IInventory` (this pass) plus `IStageRepository` (2026-08-11, admin
  CRUD counterpart to `IStageCatalog`) — all interfaces only when first added; every one but
  `IInventory` now has a real adapter, see [[game-lobby-persistence]].
- **2026-08-11 (later same day)**: `Stage` gained `description`/`picture`/`pictureThumb`/
  `pictureStatus` fields, same shape as `powers.Power` - see [[game-stage-content]]. `NewStage`
  grew two trailing params; `SetPictureRenditions` is the new pointer-receiver mutator.
- **2026-08-11**: `snapshot.go` added `Snapshot`/`Restore` — the seam a Redis-backed `IGameStore`
  needed to round-trip a `*Game` out of process (see [[game-lobby-persistence]] for the full
  rationale). `ballot.go` gained one new getter, `Votes() map[ParticipantID]OptionID`
  (copy-returning) — the only other domain-level change that tanda required, since `HasVoted`/
  `Count` alone can't reconstruct *which* option someone voted for. `GameResult` (`game_result.go`)
  gained `Participants []ParticipantOutcome`, filled by both `IGameMode.Outcome` implementations
  via a shared private `participantOutcomes(g *Game)` helper, so `IGameHistory` can record who
  played, not just what happened.
- Deleted: `internal/domain/entities/user/player.go` (the dead 9-`byte`-field `Player` struct).
  `Participant` + `Loadout` replace it; verified nothing else referenced it before deleting
  (`go build ./...` after deletion is the real proof).

## Deliberate deviations from the plain existing convention

- Domain methods here take no `context.Context` — matches every other entity in the codebase
  (`Power`, `Stand`, `User` are all ctx-free); only `ports` interfaces and application services use
  `context`.
- `enums.SpinLevel`/`HamonLevel`/etc. use SCREAMING_SNAKE wire strings (`"REGULAR"`, `"YONKO_PLUS"`)
  rather than the display labels the old `player.go` `*ToString()` methods returned (`"Regular"`,
  `"Yonko+"`) — matches `StandStat`'s wire convention; pretty labels are the frontend's job via
  i18n, same split as everywhere else in this backend.

## Known debt (flagged on purpose, not derivable from code)

- `power_traits.go` matches `REQUIRES_SPIN_4` (Tusk ACT4, Ball Breaker, Soft & Wet: Go Beyond) by
  **Stand name**. Fragile against renames — the single place to fix once `powers` gains a real
  persisted traits column (e.g. `power_traits` table/array), at which point `TraitsOf` becomes a
  lookup instead of a name match and nothing else in the game package changes.
- **Resolved 2026-08-11**: stage catalog content now has a schema, seed data, and admin CRUD (see
  [[game-lobby-persistence]]) — `ports.IStageCatalog`/`IStageRepository` both have real adapters.
- No player inventory (no schema, no unlock flow) — `ports.IInventory` is the seam;
  `enums.Inventory` as an `AbilitySource` is rejected by `game.NewConfig` until it exists. Still the
  one game-feature port with no adapter after [[game-lobby-persistence]]/[[game-realtime-transport]].
  Full planned design once this gets built: [[gameplay-versus-inventory-characters]].
- **Resolved 2026-08-11**: `ports.IGameHistory` now has a Postgres-backed adapter (see
  [[game-lobby-persistence]]) — finished/aborted games are recorded before being deleted from the
  (Redis or in-memory) lobby store.
- Bot vote heuristic (`DefaultLoadoutEvaluator`) is a first-pass linear scoring function (stand
  stats E..A→1..5, INFINITE→6, plus ability levels plus a rarity bonus) — reasonable enough to
  exercise `IGameMode`/`Game` in tests, not tuned for actual game balance.

The application layer that wires this domain up (game store, voting timer, event hub, the cheap
adapters, and what's still stubbed) landed the same day — see [[gameplay-application-layer]].

Related: [[gameplay-game-modes]], [[gameplay-application-layer]], [[ADR]], [[backend-contract]],
[[picture-events-sse]] (the existing pub/sub precedent this feature's domain events are designed to
feed).
