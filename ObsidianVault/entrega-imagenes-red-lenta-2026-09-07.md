---
title: "Slow-network catalogue images: five root causes and T1's fix (2026-09-07)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - performance
  - in-progress
---

# Entrega de imágenes en red lenta (2026-09-07)

Owner reported `/catalog/stands` loaded **zero** images on a poor-coverage
connection. Investigation found five independent root causes:

- **(a) Concurrency stampede.** The whole catalogue renders at once via
  `.map()`, so N `<img>` requests start simultaneously over one HTTP/2
  connection — each gets `bandwidth/N`; at 400 kbps with N=120, ~3 kbps per
  image, none finish. **Not fixed by T1** — needs T3's concurrency queue.
- **(b) Zero browser caching.** Every image URL is a presigned R2 URL with a
  15-min TTL and a fresh signature per response, so every visit re-downloads
  everything and a long-open tab gets dead links. **Not fixed by T1** — needs
  T2's immutable content-addressed media proxy.
- **(c) Oversized renditions.** Only `main` (1024px) and `thumb` (256px)
  existed for grid cells rendering at ~100-140px. **Fixed by T1.2** (see
  below).
- **(d) No compression anywhere.** **Fixed by T1.5.**
- **(e) `/stands`'s ETag structurally can't match**, because the cached body
  contains volatile presigned URLs with a shorter TTL than the HTTP cache
  header allows. **Not fixed by T1** — needs T2 (content-addressed URLs kill
  this at the root: no more volatile bytes in the body to hash).

**Bottom line: T1 alone does not fix the reported bug.** Root causes (a) and
(b) are the ones that actually caused "zero images loaded," and both need
T2/T3, still unimplemented as of this note. T1 is real, valuable, low-risk
work (smaller payloads, working compression, one less duplicate fetch) but a
retest on the original poor-coverage connection would still likely fail.

## What T1 shipped

- **T1.1** (picture_worker.go): dead-context bug on markFailed/deleteQuietly
  after JobTimeout fixed via `context.WithoutCancel` + fresh 5s timeout; two
  error paths (`PictureKeys`/`UpdatePicture` failure) that left a subject
  stuck at PENDING now call `markFailed` + publish a FAILED SSE event.
- **T1.2/T1.3** (migration `00013_picture_card_and_lqip.sql`): added a third
  `card` rendition (128px) to the pipeline, plus an embedded LQIP data: URI
  placeholder (`data:image/webp;base64,...`, dropped/stored as `""` if it
  exceeds `MEDIA_LQIP_MAX_BYTES`). `ports.IImageProcessor.Transcode` changed
  from a fixed two-return-value signature to a `VariantSpec`/
  `TranscodeOptions.Variants` ladder returning `map[string]EncodedImage`, so
  a future 4th variant doesn't need another positional-arg rewrite.
  **Deviation from the original plan**: `PicturePublisher`/repository
  `UpdatePicture` grew explicit `card, lqip *string` params (same COALESCE-
  preserve-if-nil semantics as `main`/`thumb`) instead of the plan's proposed
  `map[string]string` — judged simpler and equally correct for closing the
  card-orphan-leak the plan called out, without introducing a new map-shaped
  contract across ~10 call sites.
- **T1.4**: `GET /stands/options` (id/name only, unfiltered, locale-free —
  `powers.name` is not translatable) backs the evolvesFrom picker. Wired into
  the **public** catalogue container (`catalog-stands-container.tsx`),
  removing its second full-catalogue fetch. The **admin** container
  (`stands-container.tsx`) deliberately keeps its existing full `useStands()`
  fetch — it also needs each Stand's `evolvesFrom.id` chain (for the "can't
  evolve from your own descendant" picker exclusion) and `pictureStatus`
  (for its FAILED-picture toast watcher), neither of which the options
  endpoint carries. Extending `/stands/options` to include those was ruled
  out of scope for this tanda.
- **T1.5**: `middleware.Compress` wraps the `/api/v1` REST group (excluding
  `/events` and `/games/{id}/ws`, which must not be buffered) via new
  `HTTP_COMPRESS_LEVEL` config (default 5, 0 disables). Frontend nginx gained
  `gzip on` + friends at the `server` level for the Expo web bundle itself —
  `gzip*` directives have normal nginx inheritance, unlike `add_header`
  (which replaces per `location`, see [[admin-panel-crud-ux-fixes]] and the
  nginx template's own IMPORTANT comment), so no repetition was needed.
- **T1.6/T1.7** (earlier in this session): frontend call sites switched from
  full-size to thumb renditions where they were rendering into small boxes;
  `resolveAvatar` no longer returns an empty thumb for a user with no own
  avatar (was breaking every card bound to `avatarThumb`).

**Deliberately deferred**: `GameParticipantResponse`/`GameStageResponse`
did not gain `avatarCard`/`pictureCard`/`*Lqip` fields this tanda — lower
priority than the public catalogue responses (Stand/DevilFruit/Stage/User)
that were the actual target of the bug report.

## Env vars added (owner must add to `.env`/`.env.example`/GitHub Variables)

`PICTURE_CARD_DIMENSION=128`, `PICTURE_LQIP_DIMENSION=16`,
`PICTURE_LQIP_QUALITY=30`, `MEDIA_LQIP_MAX_BYTES=512`,
`HTTP_COMPRESS_LEVEL=5`.

## Next

T2 (media proxy) and T3 (pagination + frontend concurrency queue +
skeletons) are still fully unimplemented — see the approved plan
(`el-otro-d-a-estuve-encapsulated-feather.md` in the owner's local Claude
plans directory) for the full design. Planned vault notes
`media-proxy-content-addressed.md` and `catalogue-pagination.md` don't exist
yet — write them when T2/T3 actually start.
