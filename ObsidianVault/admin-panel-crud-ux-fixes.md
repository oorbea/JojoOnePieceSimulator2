---
title: "Bug: admin Stand/Devil Fruit CRUD looked broken without a cache clear, plus 3 modal bugs"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - bug
  - fixed
---

# Admin panel CRUD/UX fixes (2026-08-10)

Four separate bugs in the admin Stands/Devil Fruits screens, reported together
by the owner after using the panel. All frontend-only.

## 1. UI stayed stale after create/update/delete until a hard cache clear

Distinct from — and downstream of — the 304-body-loss bug in
[[etag-304-body-loss]]. That fix (caching `{etag, data}` together) stopped a
304 from wiping the response entirely, but it never addressed the other half:
after a mutation, the **old** `{etag, data}` pair for the list URL was still
sitting in `etag.ts`'s `cacheByKey` Map. `invalidateQueries()` triggers a
refetch, that refetch's request interceptor attaches the stale `If-None-Match`,
and — because the *list* URL's ETag hadn't been touched by the write itself —
the backend legitimately still had reason to answer `304` on the very first
refetch in some timing windows, re-serving the pre-write body. `clearEtags()`
already existed in `etag.ts` (added by the original fix) but nothing ever
called it.

Fix: every mutation hook (`useCreateStand`/`useUpdateStand`/`useDeleteStand`/
`useUploadStandPicture` and the Devil Fruit equivalents in
`use-stand-mutations.ts`/`use-devil-fruit-mutations.ts`) now calls
`clearEtags()` as the first line of `onSuccess`, before `invalidateQueries()`.
Blunt (clears the whole in-memory map, not just the affected URL) but cheap —
it's a plain JS `Map`, and the alternative (targeted per-URL eviction) isn't
worth the bookkeeping for an admin-only panel with a handful of list endpoints.

## 2. Picture upload never appeared without a manual page refresh

**Superseded 2026-08-10** — see [[picture-events-sse]]: this polling had a
counter bug (fixed same day) and was then replaced outright with SSE push on
web; native keeps a corrected version of the polling below as its fallback.

The picture pipeline (see [[user-profile-feature]]'s "picture worker
generalized" section) is async: `PATCH .../picture` returns `202` with
`pictureStatus: PENDING`, and a background worker flips it to `READY`/`FAILED`
once the WebP transcode + upload lands. Nothing was polling for that
transition — the mutation's `onSuccess` just fired an "uploading" toast and
invalidated once, so the image only ever showed up if the admin happened to
trigger another fetch (e.g. a manual reload).

Fix: `useStands`/`useDevilFruits` (`use-stands.ts`/`use-devil-fruits.ts`) now
pass a `refetchInterval` function to `useQuery`, polling only while any item
in the list has `pictureStatus === 'PENDING'`. Exponential backoff 2s → 4s →
8s ... capped at 30s, gives up after 8 attempts — copied verbatim from
`use-profile.ts`'s `pollInterval` (same problem: no websocket exists in this
project, per [[backend-contract]], so polling the list is the only way to
observe a background job finish). The containers
(`stands-container.tsx`/`devil-fruits-container.tsx`) additionally track which
ids were seen as `PENDING` in a `useRef<Set<string>>`; when polling reports
one of those ids as `FAILED`, an error toast fires
(`toasts.standPictureFailed`/`devilFruitPictureFailed`, added to all three
locale catalogs) — this is the only place that can see the transition, so it
owns telling the user the upload didn't make it.

## 3. Description field rendered ~3x shorter than the height passed to it

`glass-field.tsx`'s `<Input>` spread `{...rest}` (which includes whatever
`height`/`multiline`/`numberOfLines` the caller passes) and *then* set
`height={52}` afterward in JSX — later props win, so the hardcoded `52`
always overrode the modal's `height={90}`. Every `GlassField` usage with a
custom height was silently getting 52px instead, not just the Stand/Devil
Fruit description field.

Fix: destructure `height`/`multiline` out of `rest` in `glass-field.tsx` and
use `height ?? 52` as the actual prop, so an explicit height passed by the
caller wins and only the unset case falls back to the original default. Also
added `textAlignVertical: 'top'` + top padding when `multiline` is set, so
text starts at the top of a tall field instead of vertically centered.

## 4. Skill chips truncated long skill text with no way to read the rest

`skills-field.tsx` laid skills out as `rounded="$pill"` chips inside a
horizontally-wrapping `XStack` — fine for short labels, but the admin's
skills are full sentences ("Fencing Mastery: Demonstrates extraordinary
precision and speed..."), and a pill enforces single-line width, clipping
everything past the modal's ~520px panel width. There was no expand/detail
view to recover the clipped text.

Fix: changed the layout from horizontal-wrap pills to a vertical stack of
rectangular (`rounded="$card"`) chips, one per row, with the skill text given
`flex={1}` so it wraps across multiple lines instead of clipping.

## 5. (Same session, lower confidence) Modal sometimes showed no image although the card list did

Both card and modal read `pictureThumb`. Root cause wasn't confirmed (likely
presigned-URL expiry on a long-open admin tab — R2 links aren't permanent, see
[[storage-fallback-chain]]), but the modal was changed to read `picture`
(full-resolution) instead of `pictureThumb` regardless, since a 96×96 preview
shouldn't be relying on the smallest rendition anyway, and the container's
fallback changed from `??` to `|| null` so an empty-string URL (falsy, but not
nullish) can't sneak through as a "valid" image source.

## Where things live

- `apps/frontend/src/shared/api/etag.ts` (existing `clearEtags()`, now called)
- `apps/frontend/src/features/{stands,devil-fruits}/hooks/use-{stand,devil-fruit}-mutations.ts`
- `apps/frontend/src/features/{stands,devil-fruits}/hooks/use-{stands,devil-fruits}.ts`
- `apps/frontend/src/features/{stands,devil-fruits}/components/containers/*-container.tsx`
- `apps/frontend/src/shared/components/presentational/glass-field.tsx`
- `apps/frontend/src/shared/components/presentational/skills-field.tsx`

See also [[etag-304-body-loss]] (the earlier, related but distinct ETag bug),
[[user-profile-feature]] (source of the picture-worker/polling pattern),
[[i18n-multi-language]] (locale catalogs touched for the new failure toast).
