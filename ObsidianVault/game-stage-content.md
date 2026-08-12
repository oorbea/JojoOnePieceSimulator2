---
title: "Feature: stage description + picture (2026-08-11)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - feature
  - gameplay
---

# Stage description (3 locales) + picture (2026-08-11)

`game.Stage` (added in [[game-lobby-persistence]]'s stage catalog) only had `id/manga/order/name`.
This tanda adds `description` (translated in all three locales, mandatory - stricter than Power,
where only en-GB is) and `picture` (same background WebP pipeline as Stand/DevilFruit/user
avatars), on the owner's explicit instruction to treat both exactly like `powers.Power`.

## Status

Done. Migration `00010_stage_content.sql` applied cleanly against the live dev stack; the 19
seeded stages were backfilled with real one-sentence descriptions in all three locales
(57 strings, written by hand in the migration itself - content, not logic).

## Domain: `entities/game/stage.go`

`Stage` gained `description`, `picture`, `pictureThumb`, `pictureStatus` - same shape as
`powers.Power`. `NewStage` grew two new trailing params (`description string, picture string`,
mirroring `NewPower`'s single `picture` param); `SetPictureRenditions(main, thumb string, status)`
is the pointer-receiver mutator, same pattern as `Power.SetPictureRenditions`. `description`
non-empty is a domain invariant now, same as `name`.

**The one deliberate asymmetry**: `Stage.description` (the domain field) is always resolved at a
fixed `enums.EnGB` when read through `ports.IStageCatalog` (the gameplay-facing port `GameService`
uses to assign Stages to rounds) - documented directly on that port. Nothing on the gameplay path
ever reads it. A live match's actual per-player description comes entirely from a separate
resolution at the transport layer - see below - because a `Game` is one instance shared by every
participant, so a `Stage` frozen into a `Round` cannot carry an already-resolved description for
more than one locale at once.

## Per-player language in a live match

The owner's requirement: **each player sees text in their own configured language, even mid-match**,
not a locale fixed at round-assignment time. `user.User` already had `Language() enums.Locale`
(column added 2026-08-06, just never read by anything until now).

- `endpoints.GameEndpoints` gained two new dependencies: `ports.IStageRepository` (for
  `Translations(ctx, id)`) and `ports.IUserRepository` (for `Language()`).
- `GameEndpoints.viewerLocale(ctx, g, self)` resolves the connected participant's `UserID`, looks
  up their `User`, and returns `.Language()` - `enums.EnGB` for a bot or a lookup failure, which
  must never block rendering a game state.
- `GameEndpoints.stageTextResolver(locale)` returns a `dto.StageTextResolver` closure bound to that
  locale, calling `stages.Translations(ctx, id)` and picking the first present locale in
  `enums.FallbackChain(locale)`.
- `dto.NewGameStateResponse` gained `resolveStagePicture PictureURLResolver` and
  `resolveStageDescription StageTextResolver` params; every `GameRoundResponse.Stage` is built
  through `newGameStageResponse(ctx, stage, resolveStagePicture, resolveStageDescription)`, which
  calls the resolver instead of ever reading `Stage.Description()` directly. `Picture()`/
  `PictureThumb()`/`PictureStatus()` stay locale-independent and come straight off the domain value.
- Cost: one extra `stage_translations` lookup per round per snapshot build (create/join/get/
  by-code, and every WebSocket resend after a state-changing event) - a handful of indexed rows at
  most, same order of magnitude as everything else `GameStateResponse` already does.

## Admin CRUD (`/stages`)

Unlike the live-match path, admin browsing/editing resolves via the **request's** locale
(`LocaleFromRequest`, `Accept-Language`/`?lang=`), same convention as stands/devil-fruits - this is
"I'm managing the catalog", not "I'm playing", so it doesn't touch `user.Language()` at all.

- `ports.IStageRepository` gained `ListByManga(ctx, manga, locale)` (the admin, locale-aware
  counterpart to `IStageCatalog.Stages`, which stays fixed-EnGB and manga-only), `Translations`,
  `UpdatePicture`, and both `List`/`FindByID` grew a `locale` parameter. `StageService` dropped its
  `IStageCatalog` dependency entirely - every read now goes through `IStageRepository`.
- New `ports.StageTranslations map[enums.Locale]string` - `PowerTranslations`' shape without
  `Skills` (a Stage has none). `dto.StageTranslationRequest`/`validateStageTranslations` mirror
  `TranslationRequest`/`validateTranslations`, except **every** locale is mandatory on write (the
  owner's explicit call - stricter than Power's "only en-GB required").
- `StageService` grew the same four picture-pipeline fields as `StandService`
  (`pictures/processor/enqueuer/picPolicy`); `SetStagePicture`/`PictureURL`/`MaxPictureBytes` are
  line-for-line copies of `StandService`'s. `enums.PictureSubjectKind` gained `StageSubject`
  (`"STAGE"`); `services.NewStagePicturePublisher` adapts `IStageRepository` into the shared
  `PicturePublisher` interface, same as the Stand/DevilFruit/User adapters in
  `picture_publishers.go`; registered in `main.go`'s `pictureTargets` map under key prefix
  `"stages"`. `stageRepository`'s construction moved earlier in `main.go` purely so it could join
  that map before anything else needed it.
- New routes: `PATCH /stages/{id}/picture` (202, `PENDING`, same worker/SSE progress as every other
  picture) and `GET /stages/{id}/translations` (admin edit-all-locales-at-once form, mirrors
  `GET /stands/{id}/translations`).

## Schema (`db/migrations/00010_stage_content.sql`)

`stages` gains `picture/picture_thumb/picture_status` (reusing the existing `picture_status` enum
from `00003_picture_renditions.sql`). New `stage_translations(stage_id, locale, description)` -
`power_translations`' shape without `skills`. Backfilled inline via a `VALUES` list matched by
`(manga, name)` against the existing seed from `00008_stages.sql` (ids are `gen_random_uuid()`, so
name-matching was the only option) - dry-run verified against the real dev Postgres before
shipping (`BEGIN; ...; ROLLBACK;`), confirming exactly 57 rows (19 stages × 3 locales).

`db/query/stages.sql`: `ListStages`/`ListStagesByManga`/`GetStageByID` all gained the same
`LEFT JOIN LATERAL` fallback-chain pattern `GetStandRowsByID` uses over `power_translations`;
`UpsertStage` gained the three picture columns; new `UpdateStagePicture` (`UpdatePowerPicture`'s
shape), `UpsertStageTranslation`/`DeleteStageTranslations`/`GetStageTranslations`
(`*PowerTranslation*`'s shape, minus skills).

## Bug caught before shipping

The dry-run migration test first failed with `column "locale" is of type locale but expression is
of type text` - the `VALUES (...)` literal columns needed an explicit `::locale` cast in the
`SELECT` feeding the backfill `INSERT`. Caught immediately by running the migration inside a
rolled-back transaction against the real dev database before ever touching the live schema - worth
doing for any future migration with a data backfill, not just a `CREATE TABLE`.

## Tests

Extended `stage_repository_test.go` (`-tags integration`): the seed's descriptions are non-empty,
a new `TestStageRepository_SeedTranslations_ResolvePerLocale` asserts three *different* strings
come back for en-GB/es-ES/ca-ES on the same seeded stage (proving the LATERAL join actually
discriminates locale, not just falls back to en-GB every time), and the create/update/duplicate
tests now thread `ports.StageTranslations` through `Save` and assert `Translations()` reads back
what was written. Verified end to end against the live dev stack: migration applied cleanly on a
fresh container rebuild, and the new `/stages/{id}/translations` route responds `401` unauthenticated
(routing/auth wiring confirmed) exactly like every other admin route.

## Frontend admin CRUD + remaining backend tests (2026-08-11, follow-up tanda)

`game-realtime-transport.md` had explicitly flagged the admin CRUD screens for Stages as **not
built**. This tanda closes that gap end to end, following [[admin-panel-crud-ux-fixes]] and
[[picture-events-sse]]'s existing patterns rather than inventing new ones.

**Backend delta** - production code was already ~95% done (see above); the actual gap was test
coverage. Added `stage_service_test.go`, `stage_endpoints_test.go` (with a handwritten
`fakeStageRepository`, since none existed), `stage_mapper_test.go`, a `StageSubject` case in
`picture_worker_test.go`, and wired `router_test.go`'s `NewStageEndpoints` fixture to a real fake
instead of `nil`. All existing `go test ./...` still green. `-tags vips` not exercised locally (no
libvips on this dev machine) - left to CI per [[cicd-picture-pipeline|the CI/CD gate]].

**Frontend**: full `src/features/stages/` feature, copied from `stands`' shape (api/keys/hooks/
types/components, container+presentational split), `/admin/stages` route, and a third
`ChannelTile` (icon `Map`, tone `blue`) on the admin hub.

Three deliberate deviations from the Stand template, each because Stage's contract genuinely
differs, not because the pattern was wrong:

- **No `power-translations.ts` reuse.** Stage's translation rule (every locale mandatory, no
  `skills`) is the opposite shape from Power's ("only en-GB mandatory, others all-or-nothing"), so
  parametrizing the existing module would have meant threading a "which locales are required" flag
  through code that assumes Power's rule everywhere. A sibling `shared/lib/stage-translations.ts`
  was added instead - same factory-not-singleton/i18n-key-message conventions, no `superRefine`
  since there's no partial-fill case to reject.
- **`LocaleTabs` extended, not forked.** It only ever supported one starred ("mandatory") locale.
  Since every locale is mandatory for Stage, `requiredLocale` now accepts `Locale | Locale[]` -
  existing Stand/DevilFruit call sites are unaffected (a single `Locale` still works), Stage passes
  `SUPPORTED_LOCALES` to star all three tabs.
- **No server-side pagination.** The catalogue is ~19 rows total. "Search + filter" from the
  original ask became: the manga filter goes to the backend's existing `?manga=` (a distinct
  `stageKeys.list(filters)` cache branch, same mechanism Stand's rarity/stat filters already use),
  and free-text search over name+description runs client-side over whatever that returned. Adding
  `LIMIT/OFFSET` would have been complexity with no real user at this catalogue size.

**SSE bug caught and fixed along the way**: `picture-events-bridge.tsx`'s `PictureEventDTO.kind`
handler was an `if/else-if/else`, and the final `else` treated *any* kind other than
`STAND`/`DEVIL_FRUIT` as a profile (`USER`) event. The backend has emitted `kind:"STAGE"` since the
picture-pipeline section above landed, so every Stage picture-ready/failed event was silently
invalidating `profileKeys.me` instead of the stage catalogue - nothing crashed, it just never
refreshed the Stages screen over SSE (native's polling fallback masked it). Replaced with an
explicit `switch` with one case per kind and a no-op `default`, added `STAGE` to the DTO union, and
added `stageKeys.allLocales` to the bridge's reconnect resync.

**Known follow-up, not done this tanda**: no `cache/stage_repository.go` Redis decorator exists yet
(Stand/DevilFruit have one). `stageRepository` in `main.go` stays uncached - functionally fine
(no staleness risk, since there's no cache to go stale), but every `IStageCatalog.Stages` call in
`GameService.CreateGame` is a live Postgres round trip, same debt already flagged in
[[game-lobby-persistence]].

Related: [[game-lobby-persistence]], [[game-realtime-transport]], [[gameplay-domain-design]],
[[gameplay-application-layer]], [[i18n-multi-language]], [[admin-panel-crud-ux-fixes]],
[[picture-events-sse]].
