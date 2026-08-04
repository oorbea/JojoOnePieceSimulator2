---
title: User profile / admin panel feature
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - decision
---

# User profile / admin panel (2026-08-04)

Full implementation plan lived at the session's plan file; this note captures
the decisions and structural changes that matter for future work, since the
plan file itself isn't part of the repo history.

## Decisions

- Self-editable: **only `username`** and a self-uploaded avatar. `email`,
  `role`, `googleSub`, and `completeName` are never self-editable —
  `completeName` stays Google-owned on purpose (simpler than tracking a
  second "is this custom" flag).
- Authorization shape: `/users/me` routes with **no id in the path** — every
  handler reads the caller's id exclusively from JWT claims
  (`claimsFromContext`/`ClaimsFromRequest`), so there is no code path where a
  path or body id could address another user's profile.
- Avatar vs Google picture: **separate columns.** `users.picture` was renamed
  to `google_picture` (sync'd on every login, read-only) and
  `avatar_key`/`avatar_thumb_key`/`avatar_status` were added (R2, user-owned,
  never touched by login sync). API resolves: own avatar if present, else
  `google_picture`.
- `GET /users/{id}` is authenticated but public-projection: a dedicated
  `dto.PublicUserResponse` (separate Go struct, not a field-omitting variant
  of `UserResponse`) makes it structurally impossible to leak email/role by
  future-editor oversight.
- Admin can: list users, edit another user's username (moderation), change
  role, delete. Self-demotion/self-delete via the admin routes returns 403;
  demoting/deleting the last admin returns 409 (`services.ErrLastAdmin`).
- Role changes only take effect on the affected user's next login (role
  lives in the JWT). `ADMIN_EMAILS` re-promotes on every login, so demoting
  one of those emails is temporary — documented in the swagger godoc, not
  enforced in code.

## Structural refactor: picture worker generalized beyond Powers

The single background image-compression worker
(`internal/application/services/picture_worker.go`) previously only knew
about `powers.PowerID` and `enums.PowerKind` (the Stand/DevilFruit
class-table discriminator, persisted via the `power_kind` column). Avatars
needed the same pipeline (transcode → WebP main+thumb → R2 → publish), so:

- `ports.PictureJob.PowerID powers.PowerID` → `SubjectID string`.
- New **`enums.PictureSubjectKind`** (`STAND|DEVIL_FRUIT|USER`) — deliberately
  a *new*, in-memory-only enum, not an extension of `enums.PowerKind`, since
  `PowerKind` is a DB-persisted discriminator and mixing routing concerns
  into it would leak worker plumbing into the Power domain model.
- `PicturePublisher.PictureKeys`/`UpdatePicture` now take a string id; each
  adapter (`standPicturePublisher`, `devilFruitPicturePublisher`,
  new `userPicturePublisher`) parses it back into its own concrete id type
  internally.

This is the pattern to follow if a fourth picture-bearing subject ever shows
up: add a `PictureSubjectKind` value + a `PicturePublisher` adapter + a
`PictureTarget` entry in `cmd/app/main.go` — the worker itself never changes.

## Frontend/backend contract coupling caught mid-build

Changing the login response's `picture` field to `avatar`/`avatarThumb`/
`avatarStatus` broke `SessionUser`/`AuthGoogleResponse` (which predate this
feature and directly reused the backend's user shape 1:1). Fixed by keeping
`SessionUser.picture` as the *display* field (unchanged everywhere it's read,
e.g. `home-screen.tsx`) and mapping `avatar → picture` at the one boundary
that matters: `use-google-auth.ts`'s `completeSignIn`. Lesson: a DTO field
rename on the backend needs an explicit audit of every frontend type that
mirrors that DTO by name, not just the feature that motivated the rename.

## Where things live

- Backend: `internal/domain/entities/user/{user.go,username.go}`,
  `internal/application/services/user_service.go`,
  `internal/infrastructure/api/endpoints/user_endpoints.go`,
  `internal/infrastructure/api/dto/{user_request,user_response,public_user_response}.go`,
  migration `db/migrations/00005_user_avatars.sql`.
- Frontend: `src/features/profile/` (full feature: api/hooks/containers/
  presentational), `src/shared/components/presentational/glass-field.tsx`
  (first Input/TextField wrapper in the shared component set), route
  `app/(app)/profile.tsx`.
- Not built yet (deferred per plan): the admin users panel screen
  (`app/(app)/admin/users.tsx`) — backend admin routes exist and are tested,
  frontend screen is future work.

See also [[norma-diseno-ui-ux]] (design-skill norm applied while building the
profile screen) and [[backend-contract]] (API shape summary).
