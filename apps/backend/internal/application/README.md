# Application

Use cases, one Go package (`services/`) that turns domain entities +
`internal/domain/ports` interfaces into application behaviour. No transport
(HTTP/WS) and no concrete infrastructure — services depend only on port
interfaces injected at construction.

## `services/`

- **`auth_service.go`** — `AuthService`: Google login/registration, username
  derivation, admin-role assignment (from `ADMIN_EMAILS`).
- **`user_service.go`** — `UserService`: profile/avatar/language self-service.
- **`stand_service.go`**, **`devil_fruit_service.go`**, **`stage_service.go`**
  — CRUD-style catalogue services, same shape: an `...Input` struct, a
  `New...Service(repo, idGen, ...)` constructor, Create/Update/Get/List/
  Filter/Delete methods.
- **`game_service.go`** (largest file, ~1.5k lines) — `GameService`,
  `CreateGameInput`, `VotingPolicy`, `LobbyListing`, `phaseTimer`,
  `gameLocks`: orchestrates lobby → game lifecycle, voting, and phase
  timers on top of the `game.Game` state machine.
- **`game_event_hub.go`**, **`picture_event_hub.go`** — pub/sub hubs
  (`GameEventHub` + event structs) that fan domain events out to the
  infrastructure SSE/WS handlers.
- **`picture_worker.go`**, **`picture_publishers.go`** — the background
  picture-transcode worker (`PictureWorker`, `WorkerConfig`) shared by
  Stands and Devil Fruits, and its event publishers.
- **`storage_reconciler.go`** — `StorageReconciler`: periodic job
  reconciling R2 object-storage usage against a ledger.
- **`clock.go`** — `Clock`/`Timer` interfaces abstracting `time.AfterFunc`,
  so `GameService`'s voting/revote timers are deterministic in tests.

## Conventions

- A service is a struct holding injected `ports.I...` interfaces (plus a
  generic `IIdGenerator[T]` where it mints ids), constructed via
  `New<Name>Service(...) *Service`.
- Errors are package-level sentinels: `var Err... = errors.New(...)`.
- Every exported type/func/var carries a doc comment explaining *why*,
  often cross-referencing a sibling implementation for consistency (e.g.
  "mirrors UserService.DeleteAvatar's delete-after-write pattern").
- Side-effecting picture operations follow persist-then-best-effort-cleanup
  (log on error), and async work is dispatched to `PictureWorker` and
  reverted on failure.
- Test doubles live behind small local interfaces (e.g. `usageRefresher` in
  `storage_reconciler.go`) so services stay testable without real infra.
- Each service has one or more colocated `_test.go` files, sometimes split
  by scenario (`game_service_rematch_test.go`,
  `game_service_voting_ends_test.go`, ...).
