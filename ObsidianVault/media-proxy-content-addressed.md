---
title: "T2: immutable content-addressed media proxy"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - performance
  - in-progress
---

# Media proxy content-addressed (2026-09-08)

Root cause (e) from [[entrega-imagenes-red-lenta-2026-09-07]]: `/stands`'
ETag structurally couldn't match because its cached body contained volatile
presigned URLs. T2 kills that at the root by making every catalogue picture
URL a pure function of content — no signature, no expiry, no I/O to build.

## URL scheme

```
GET /api/v1/media/{group}/{variant}.webp                    # scope=public
GET /api/v1/media/p/{exp}/{sig}/{group}/{variant}.webp      # scope=private
```

`{group}` — 32 lowercase hex chars, `sha256(MEDIA_ID_SALT || mainRenditionBytes)[:16]`
hex-encoded. Computed once by `picture_worker.go` right after transcode,
persisted as `media_objects` rows (one per variant, keyed by `(group_id,
variant)`), and set on the entity via `SetMediaID`/`SetAvatarMediaID` -
separate calls from `UpdatePicture`/`UpdateAvatar` since the group id is
only known after both the transcode *and* the media_objects write succeed.

Same source bytes + same transcode settings → same group, always. A
re-upload of an identical file reuses the same URL (free dedup); any pixel
change gets a brand-new group — **zero invalidation, anywhere**. That
property is also why `dto.MediaURLBuilder` is deliberately pure: no `ctx`,
no `error`, no I/O. `NewStandResponse` (and its DevilFruit/Stage/User
siblings) build a media URL as string concatenation once `PictureMediaID`
is set, and only fall back to the old presign path (`PictureURLResolver`)
while it's still empty (not backfilled yet).

`{variant}` is a path segment (`card`/`thumb`/`main`), not a query param -
three independent immutable resources per group.

## Public vs private scope

**Public** (Stand/DevilFruit/Stage): no signature, no expiry,
`Cache-Control: public, max-age=31536000, immutable`. Safe because these
are identical for every viewer and already readable by any logged-in user
via the existing catalogue endpoints - the marginal exposure is "an image
to someone who already has an unguessable 128-bit URL for it," and nothing
about the group id is enumerable (it only resolves if a `media_objects` row
exists, which only an admin upload creates).

**Private** (User avatars): a durable handle to a real person's face, so it
requires a signed URL - `sig = base64url(HMAC-SHA256(MEDIA_URL_SECRET,
exp|group|variant))[:22]`, verified with `subtle.ConstantTimeCompare`.
`exp` is **quantized** to `MEDIA_PRIVATE_URL_TTL/2` windows
(`dto.MediaURLBuilder.privateWindow`/`quantizeExp`): every call within the
same window returns the byte-identical URL, which is the property browser
caching, a future service worker, and the response ETag over a user list
(many avatar URLs in one body) all depend on. Authority lives in the URL,
never in a header - deliberately **no** `Vary: Authorization`.

**Security tradeoff, stated plainly**: a private URL is a valid bearer
capability for anyone who obtains it (a shared HAR, a screenshot, an access
log entry) until `exp` — same class of exposure the old 15-minute presigned
URL had, just a longer window (default 24h). Mitigations: `MEDIA_URL_SECRET`
rotation revokes every live URL at once; the TTL is configurable to shorten
it; and avatars are already visible to every logged-in user via
`PublicUserResponse` today, so the actual delta is "logged-in users" →
"URL holders," not "private" → "public."

## Rejected: AVIF renditions

AVIF encode is 5-20x WebP's cost for a diminishing return on an already-tiny
128px card, and would force `Vary: Accept` on the response — which destroys
the one-URL-per-content-hash property this entire design depends on. If
AVIF is ever wanted, the correct shape is a separate `{variant}.avif` URL
under `<picture><source type="image/avif">`, never content negotiation.

## Proxy vs 302 redirect

Chosen: **proxy** (`MEDIA_MODE=proxy`, the default) — the backend downloads
from R2/B2/Supabase (through an on-disk LRU cache, `MEDIA_CACHE_DIR`,
default 256 MiB) and serves the bytes itself. `MEDIA_MODE=redirect` is kept
as an escape hatch (302 to a fresh presigned URL) but is strictly worse
here: a 302 gets cached by the client pointing at a `Location` that a SigV4
presign (7-day hard cap) will eventually expire, so it could never actually
be `immutable`. Proxy's R2 egress cost is paid once per rendition per
backend restart (or cache eviction) and amortizes to ~zero after that -
`immutable, max-age=1y` means a repeat client visit downloads **zero
bytes**, versus the pre-T2 baseline of re-downloading the whole catalogue
on every visit.

## Rate limiting

`globalRateLimit` moved off the router root (it used to apply everywhere,
including to what became `/media`) onto every group that isn't `/media`
explicitly. `/media` gets its own `mediaRateLimit` tier
(`RATE_LIMIT_MEDIA_PER_IP`, default 1200) - a cold catalogue load alone is
~100 image requests, which would blow through the old 240/min global tier
instantly. Not optional: this was the #1 way T2 could turn into a *new*
outage instead of a fix.

## Storage read path

`ports.IStorageBackend.Get`/`ports.IPictureStorage.Download` are new -
before T2 there was no way to read bytes back at all (only
Upload/PresignGetURL/Delete). `s3store.Backend.Get` maps a missing key to
`ports.ErrObjectNotFound` (→ 404 via `endpoints/errors.go`) instead of a
generic 500.

## Deviations from the original plan

- `PicturePublisher`/repository interfaces grew explicit `SetMediaID`/
  `SetAvatarMediaID` methods instead of folding the media id into
  `UpdatePicture`'s already-6-parameter signature - keeps that signature
  stable and matches the "media id is set after a separate DB write"
  timing.
- Endpoints structs (`StandEndpoints`, `UserEndpoints`, etc.) get the
  `dto.MediaURLBuilder` via a post-construction `SetMediaURLBuilder(...)`
  setter rather than a constructor parameter, so every existing test/call
  site kept compiling unchanged - the zero value is safe (never invoked
  while `PictureMediaID` is empty).
- `mediabackfill` (T2.7) doesn't re-upload existing main/thumb bytes - it
  downloads main once, transcodes only `card`+`lqip` from it, and indexes
  all three variants (two pointing at their existing storage keys, one
  newly uploaded) under the freshly computed group id.

## Still open

- Backfill hasn't run in prod - every existing row still resolves through
  the presign fallback until `mediabackfill` is run (see cmd/mediabackfill's
  doc comment for the CLI, and its two-layer safety net: DTO falls back to
  presign while `picture_media_id == ""`; the media handler itself falls
  down the card→thumb→main ladder for a partially-backfilled group).
- No frontend changes needed or made - `picture`/`pictureThumb`/
  `pictureCard`'s JSON field names are unchanged, only their *values*
  sometimes become media proxy URLs instead of presigned ones. Confirmed via
  `make types-check-docker`: the only contract diff from this whole tanda is
  two new `apierr` codes (`OBJECT_NOT_FOUND`, `MEDIA_SIGNATURE_INVALID`).
- T3 (pagination + frontend concurrency queue + skeletons) is still
  unimplemented - see [[entrega-imagenes-red-lenta-2026-09-07]].
