# Domain

Framework-free core: entities, value objects, enums, and outbound port
interfaces. No dependency on `internal/application` or
`internal/infrastructure` — everything here is pure Go plus `errors`/`time`.

## `entities/`

One package per bounded concept. Entities are built via a `NewX(...)
(*X, error)` factory that validates invariants; fields are private, mutated
only through named methods (`SetPictureRenditions`, `ChangeUsername`,
`ChangeRole`, ...).

- **`powers/`** — `Power` (id, name, description, rarity, skills, picture +
  thumbnail + processing status), embedded by `Stand` (stat block of
  `enums.StandStat` E→A/Infinite/Null, optional `evolvesFrom *Stand`) and
  `DevilFruit`.
- **`user/`** — `User` (Google-synced identity + self-service username,
  avatar, language, role) and username validation.
- **`game/`** — the largest package, the lobby/game aggregate. `Game` is an
  explicit state machine over `enums.GameState`
  (`LOBBY → ASSIGNING → VOTING → [TIEBREAK] → RESOLVING → ... → FINISHED/ABORTED`),
  delegating all mode-specific behaviour to an `IGameMode` strategy
  (`gauntlet_mode.go`, `versus_mode.go`). Supporting entities: `Team`,
  `Participant`, `Stage`, `Round`, `Ballot`, `Loadout` (+ builder/evaluator),
  the power pool/filter, `reveal.go`, `snapshot.go` (persistence snapshot),
  `events.go` (domain events), `config.go`, `weights.go`, `game_result.go`.
  Near-1:1 file-to-test-file coverage.

## `valueobjects/`

`id.go` — generic `Format`/`Parse`/`IsNil` helpers over any `~[16]byte`
type. Each entity package defines its own concrete id type on top of this
(`powers.PowerID`, `user.UserID`, `game.GameID`, `game.ParticipantID`,
`game.TeamID`, `game.StageID`), so id handling is never duplicated per
entity.

## `enums/`

~20 small typed-constant packages (byte- or string-backed), each exposing
`String()`, `IsValid()`, a `Parse*` function, and an `ErrInvalid*` sentinel:
`StandStat`, `PowerRarity`, `PowerKind`, `PowerTrait`, `FruitType`,
`FruitMastery`, `HamonLevel`, `HakiLevel`, `SpinLevel`, `PhysicalForm`,
`AbilitySource`, `Manga`, `Locale` (with `FallbackChain`), `UserRole`,
`PictureStatus`, `PictureSubjectKind`, `GameModeKind`, `GameState`,
`ParticipantKind`, `LobbyVisibility`, `RevealSpeed`, `SquadVerdict`.

## `ports/`

Outbound interfaces the domain/application layers depend on, implemented in
`internal/infrastructure`: repositories (`IStandRepository`,
`IDevilFruitRepository`, `IUserRepository`, `IStageRepository`,
`stage_catalog.go`), `auth.go`, `cache.go`, `random.go`, the picture
pipeline (`image_processor.go`, `picture_storage.go`, `picture_enqueuer.go`,
`storage_backend.go`, `storage_ledger.go`), game-specific ports
(`game_power_pool.go`, `game_history.go`, `tiebreaker.go`, `inventory.go`,
`assignment_weights.go`), translation ports
(`power_translations.go`, `stage_translations.go`), and `errors.go` — the
sentinel domain errors (not-found / already-exists / auth / rate-limit
variants) every layer above matches against. Repository ports accept an
`enums.Locale` and resolve translations via `enums.FallbackChain`.
