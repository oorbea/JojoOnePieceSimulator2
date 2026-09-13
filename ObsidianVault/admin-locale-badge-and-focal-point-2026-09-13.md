---
title: "Admin panel: locale badge on translated fields + picture focal point (2026-09-13)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - backend
  - decision
---

# Admin panel: locale badge + picture focal point (2026-09-13)

Two owner-requested admin-panel improvements, shipped together, unrelated to each other.

## 1. Which language am I editing?

`LocaleTabs` ([[i18n-multi-language]]) already existed above the translated fields in the
Stand/Devil Fruit/Stage/Character forms, but the modal scrolls and the tabs leave view once you
reach Description/Skills. Fix: `TranslatedContentGroup`
(`shared/components/presentational/translated-content-group.tsx`) wraps `LocaleTabs` + the
translated Controllers in a `GlassPanel` with a header pill showing the active locale's endonym
(`LOCALE_ENDONYMS`). Purely presentational, no new state - `activeLocale` was already a prop on
every one of the four form modals.

## 2. Card image crop — focal point

Nothing cropped server-side before this: `imaging/processor.go`'s `ThumbnailWithSize(...,
InterestingNone, SizeDown)` is a fit-inside scale of the whole image ([[media-proxy-content-addressed]]),
and the visible crop happens purely client-side via `contentFit:'cover'` in `lazy-image.tsx`. That
always crops to the image's own center, which cuts off heads on vertical portraits.

**Decision: store-only focal point, no server crop.** `focalX`/`focalY` (0..1, default 0.5/0.5) on
`powers`, `stages`, `characters`, and `avatar_focal_x/y` on `users` — migration `00016`. Nothing in
the worker/vips pipeline changed; a client just reads the point and passes it to `expo-image`'s
`contentPosition`. Rejected alternative: a real crop rect in `PictureJob`/`VariantSpec` — would need
a re-transcode round trip on every focal change and the media group id (hashes the *main*
rendition's bytes) would need to fold the crop params in too, since main is the uncropped one.

Transport: **plain fields on the existing create/update request**, not a new endpoint —
`FocalX/FocalY *float64` on every `*Input`/`*Request`, nil preserves the existing value on update
(same pattern the picture-rendition fields already use). The one exception: users have no
catalogue-style `UpdateProfileRequest`-adjacent "Input" struct for a general profile PUT, so the
avatar's focal point rides on the existing `PATCH /users/me` (`dto.UpdateProfileRequest` gained
`FocalX/FocalY`) via a new `UserService.ChangeAvatarFocalPoint` + `UpdateUserAvatarFocalPoint` sqlc
query, rather than inventing a fifth picture-pipeline endpoint.

**Traps hit along the way:**
- `cmd/typegen` panics on any field kind it doesn't recognize — no DTO had ever had a `float64`
  field before. Added `reflect.Float32/Float64` to both `zodExprForType` (model.go) and
  `tsTypeForType` (emit.go) before regenerating contracts, or `make types`/CI's `contracts` job
  would panic.
- Redis cache namespaces (`internal/infrastructure/cache/keys.go`) bumped `v2→v3`
  (stands/devil_fruits/stages) and `v1→v2` (jojo/one_piece characters) — same reasoning as the
  card/lqip rollout: a pre-migration cache entry would otherwise deserialize with focal 0.0 (top-left
  crop) instead of the correct 0.5 center.
- `0` is a legitimate focal value, so every request DTO uses `*float64` (nil = "leave unchanged"),
  never a plain `float64` that couldn't distinguish "omitted" from "explicitly zero".

**Frontend:** `picture-source.ts` gained `focalPosition(entity)` (the one place that turns a stored
point into `LazyImage`'s `contentPosition`, mirroring how `cardSource`/`thumbSource` already own the
rendition ladder) — returns `null` for the default center so every call site can pass it
unconditionally. Threaded through every `'cover'` consumer: catalogue cards, the in-game stage
banner, loadout cards/modal, `power-reveal-card` (via a new `PowerBlock.contentPosition` prop), and
the profile screen's own avatar. New `FocalPointPicker` component
(`shared/components/presentational/focal-point-picker.tsx`): a `PanResponder`-driven crosshair over
the full uncropped image, with a live mini-preview using the same `LazyImage` the real card uses.
Wired into all four admin form modals (nested nulls, but wired) via the container's `setValue`
resetting both fields to 0.5 whenever a new image is picked (owner decision: a new image has no
relationship to the previous crop).

**Update (same day): game participant avatars closed too.** `game.Participant` gained
`avatarFocalX/avatarFocalY float64` (default 0.5/0.5) + a `SetAvatar` signature that now also takes
the focal pair, set at all four call sites in `game_service.go` (`CreateGame`, join, host reassign,
reseat) straight from `user.User.AvatarFocalX()/AvatarFocalY()`. `ParticipantSnapshot` carries the
same two fields so a live game's Redis snapshot survives a restart with the right crop.
`GameParticipantResponse` passes them through as plain fields (same pattern as `GameStageResponse`).
Frontend: `ParticipantAvatar` (used by both the roster tile and `VoterAvatar`) now passes
`contentPosition={focalPosition({ focalX: participant.avatarFocalX, focalY: participant.avatarFocalY })}`
into its `LazyImage`. No new endpoint, no cache-namespace bump needed — gamestore's Redis snapshot
isn't behind the versioned `cache/keys.go` scheme the catalogue reads use, it's ephemeral per-game
state with its own TTL, so a stale in-flight game restored mid-deploy just shows a centered avatar
until someone rejoins/reconnects, not worth guarding against.

**Gotcha for anyone touching `focal-point-picker.tsx`:** a new ESLint rule (`react-hooks/refs`,
React Compiler-oriented) flags *any* `.current` read or write outside an effect/handler, including
the common "`useRef` mirror of the latest state for a stable callback" pattern. The fix here was to
stop caching the `PanResponder` in a ref at all — `PanResponder.create(...)` is rebuilt fresh every
render (cheap, no native binding until a touch starts) so its closures always see the current
render's `wellSize`/`onChange` directly, no ref needed.

Verified: backend `go build`/`go vet`/`go test ./...` (unit + `-tags integration` against real
Postgres/Redis via `make db-up`) all green, contracts regenerated with zero drift; frontend
`tsc --noEmit` clean, `eslint` 0 errors, `pnpm jest` 69/69 suites — 1340/1340 tests.

**Update (later same day): the shipped gesture didn't actually work** — see
[[focal-point-gesture-fix-and-modal-2026-09-13]] for the root cause (wrong host element on web,
wrong coordinate space) and the redesign into a mandatory post-upload framing modal.

Related: [[i18n-multi-language]], [[media-proxy-content-addressed]], [[admin-panel-crud-ux-fixes]],
[[entrega-imagenes-red-lenta-2026-09-07]], [[norma-verificacion-docker]],
[[focal-point-gesture-fix-and-modal-2026-09-13]]
