---
title: "Sorteo case-strip images blank: image-queue dedupe bug + prefetch fix (2026-09-25)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - performance
  - bugfix
---

# Sorteo case-strip images blank (2026-09-25)

Since [[game-match-assignment-frontend]]'s CS:GO-style case strip (commit
`97ed74e`), most tiles in the Stand/Devil Fruit strip rendered as blank
Sparkles fallbacks instead of art. Root cause was in the app-wide image
concurrency queue, not in the strip or the backend.

## Root cause: `image-queue.ts` never granted duplicate keys

`shared/lib/image-queue.ts` (see [[entrega-imagenes-red-lenta-2026-09-07]]
for why it exists) dedupes entries by URI so N tiles sharing one picture
share one concurrency slot. The strip is the first screen to show the same
uri on many tiles at once (50 cards, drawn with replacement from the pool),
and that exposed three bugs stacked on top of each other:

1. **A second subscriber of an already-enqueued entry never got its own
   `onGrant`** — only the first caller's callback was stored. Every
   duplicate tile stayed on the Skeleton forever.
2. **The slot was only freed once `refCount` hit 0**, which only happened
   on unmount for the never-granted duplicates — but `cancel()` is
   deliberately a no-op once the shared entry is granted (the queue's own
   in-flight cancellation policy), so those duplicates could never release
   their ref. The slot stayed occupied for the full 20s watchdog.
3. **The watchdog then errored the FIRST consumer too**, even if its image
   had already loaded — `onTimeout` fired unconditionally on slot
   reclaim, flipping a loaded tile back to the fallback icon.

Net effect: ~4 distinct images granted per 20s window, on a strip that's
only on screen for ~5s (spin duration × `REVEAL_SPEED_MULTIPLIER`).

## Fix

`image-queue.ts` now tracks a `Set<Subscriber>` per entry instead of a
single `onGrant`/`onTimeout` pair:

- Every subscriber of a granted entry gets its own `onGrant` (immediately,
  if joining after the grant already happened).
- The slot frees as soon as the **first** subscriber settles (loaded or
  errored) — the image is now in the platform's own cache (browser HTTP
  cache / expo-image disk cache), so every other subscriber's `<Image>`
  resolves from cache near-instantly without holding a slot.
- The watchdog only times out subscribers that haven't settled yet — a
  subscriber whose image already loaded is never flipped to `error`.
- `cancel()` on a granted entry stays a no-op per the existing policy, but
  now correctly waits for every subscriber to leave (not just the one that
  happened to hold the original grant) before releasing the slot.

Tests added: `shared/lib/__tests__/image-queue.test.ts` (the file had none
before — duplicate grant fan-out, first-settle slot release, watchdog not
firing after load, unmount-after-grant slot release).

## Also shipped: pool prefetch

New `features/game/hooks/use-power-pool-prefetch.ts`, mounted in
`lobby-room-container.tsx`: once the lobby has both `snapshot.config.
poolFilter` and the full Stand/Devil Fruit catalogue, it warms
`expo-image`'s own cache for every filtered candidate's thumb (`thumbSource`
from `picture-source.ts`) in small batches (6 at a time, 150ms apart),
deliberately OUTSIDE `image-queue.ts` — this is background warmup with no
user-visible well and no deadline, not something that should compete for
the queue's concurrency budget.

- Re-fires when the candidate set's ids or urls change (covers Versus's
  `ReassignsEachRound`, and a config edit banning/allow-listing power
  pool entries).
- Also refetches the Stand/Devil Fruit catalogue once per ASSIGNING window
  if the cached data is older than 5 minutes, since rows without a
  backfilled `picture_media_id` fall back to a 15-minute presigned URL (see
  [[media-proxy-content-addressed]]) — a lobby open longer than that can be
  sitting on expired thumb URLs otherwise.

## Also fixed in the strip itself

- `reveal-stage.tsx` now builds `CaseStripCard.picture` via `thumbSource()`
  instead of reading `pictureThumb` directly — an empty (not-yet-
  transcoded) thumb no longer skips straight to the fallback icon (see
  [[admin-panel-crud-ux-fixes]] for why `thumbSource` exists).
- `case-strip-reel.tsx` tiles now key by `${card.id}-${i}` (not the bare
  array index), pass `priorityHint="high"` (the queue's longer-watchdog
  lane — the strip can't wait behind normal `grid`-lane catalogue traffic)
  and `order` ranked by distance from the landing card, and set
  `recyclingKey={card.id}`.
- Deliberately did NOT add a manual `useMemo` around the strip build in
  `reveal-stage.tsx` — the existing code comment there notes it's left
  unmemoized on purpose because the React Compiler (already adopted app-
  wide) handles this automatically; adding one by hand would fight the
  compiler's own lint rule.

## Still open

- `mediabackfill` still needs a prod run so every catalogue row gets an
  immutable content-addressed URL instead of a 15-min presigned fallback
  (tracked already in [[media-proxy-content-addressed]] — not new to this
  note, just still blocking a real fix for stale-URL cases the 5-minute
  refetch above only mitigates).

Related: [[entrega-imagenes-red-lenta-2026-09-07]],
[[media-proxy-content-addressed]], [[game-match-assignment-frontend]],
[[admin-panel-crud-ux-fixes]]
