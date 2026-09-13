---
title: "Focal point: gesture fix + moved to a mandatory post-upload modal (2026-09-13)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - decision
  - gotcha
---

# Focal point: gesture fix + mandatory post-upload modal (2026-09-13)

Follow-up to [[admin-locale-badge-and-focal-point-2026-09-13]]: the picker shipped that day did
not respond to click or drag at all, on any platform. Two real bugs, both in
`shared/components/presentational/focal-point-picker.tsx`.

## Bug 1: dead on web — wrong host element

`panResponder.panHandlers` were spread onto a Tamagui `YStack`. Tamagui's web renderer emits a
plain `div`, which never implements React Native's responder system — `onStartShouldSetResponder`/
`onResponderGrant`/`onResponderMove` were present as *props* but nothing in react-native-web ever
calls them, since only RN-core host components (`View`, `Image`, …) carry that wiring on web.

Proof by contrast: the one drag gesture in this codebase that *does* work on web
(`features/game/hooks/use-player-drag.ts`, used by the lobby roster) spreads the identical
`panHandlers` onto a real react-native `View` (`player-row.tsx`), not a Tamagui primitive.

> **Rule going forward:** anything using `PanResponder.create(...).panHandlers` must host it on
> `View` from `'react-native'`, never a Tamagui/styled component, or the gesture silently does
> nothing on web while looking completely normal on native (Tamagui's native output *does* forward
> those props to a real `View` underneath, so this bug is invisible unless someone actually tests
> the web build).

## Bug 2: wrong coordinate mapping (native too, just less visible)

`locationX/locationY` (relative to the *well*) were divided by the well's own size, but the image
inside it rendered `resizeMode="contain"` — letterboxed whenever the well and the image didn't
share an aspect ratio. A click landed at a different normalized point than the one the user
actually touched, and the crosshair was positioned against the well's coordinate space too.

Fix: factored the "fit inside, preserve aspect ratio" math into a pure, tested helper —
`imageRectForWell(natural, well)` in `shared/lib/picture-source.ts` — which the picker uses to size
an inner `View` to exactly the image's *displayed* rect (read via the `<Image>`'s `onLoad`
`source.width/height`), and hosts the gesture + crosshair on that same rect. Gesture and rendering
now share one coordinate space unconditionally. A second pure helper, `focalFromLocation(locationX,
locationY, width, height)`, does the location→0..1 clamp+normalize math the picker and the modal
both need — same "push the interesting math into a pure lib function, keep the component thin"
shape as `focalPosition` itself.

## Redesign: mandatory post-upload framing modal

Owner ask alongside the fix: framing should be asked for in its own modal immediately after
picking a picture (not an inline widget buried in the long entity form), mandatory to confirm, on
every picture-bearing resource — Stands, Devil Fruits, Stages, both Character kinds, and the user
avatar.

New `shared/components/presentational/focal-point-modal.tsx`: same `Modal` + dimmed-backdrop +
centered `GlassPanel` recipe as `ConfirmSheet` (a real RN `Modal`, not an absolute overlay — RN
only compares `zIndex` between direct siblings, the same reason `ConfirmSheet` itself is a real
`Modal`). Holds its own local draft so dragging never writes through to the form/network on every
move — only `onConfirm` pushes the value out, once. Two modes:
- **Mandatory** (right after picking a picture): no `onCancel` passed, backdrop press is a no-op,
  and `onRequestClose` (Esc on web / Android back) **commits the current draft** instead of
  trapping the admin with no way out — framing is never destructive, worst case is the default
  center.
- **Reopened** ("Ajustar encuadre" button in the edit form, once a picture already exists):
  `onCancel` restores the value the form had before reopening.

Each of the five admin flows follows the same shape: the container's `onPickPicture` now opens the
modal (mandatory) right after `pickPicture()` resolves — since admin uploads are already deferred
(the picked asset stays local in `pendingPicture` until the entity save succeeds), "right after
uploading" is really "right after picking", against the local file URI, no network round-trip
needed first. A new `onAdjustFocal` container function reopens it non-mandatory. The four
catalogue form modals no longer render `FocalPointPicker` inline — just a small `LazyImage`
preview plus the "Ajustar encuadre" button (tooltipped, per [[feedback_tooltips_norma]]).

The user avatar flow was the one real behavior change beyond "add a modal": its upload is
*immediate* (`PATCH /users/me/picture`, not deferred), and the previous implementation persisted
the focal point on a 400ms debounce per drag tick (`profile-container.tsx`'s `saveFocalTimeout`/
`avatarFocalDirty` ref). That debounce is gone — the modal's confirm-once model replaces
save-on-every-drag, so the avatar's `PATCH /users/me` now fires exactly once per framing edit
instead of once per drag pause.

`FocalPointPicker` itself also grew three live previews (card crop, circular avatar, wide banner —
the three real shapes the app crops this picture into) instead of one, and desktop keyboard
support (arrow keys move the point 1%/press, 10% with Shift, web-only via `isWeb` from
`shared/lib/web-blur.ts`).

## Gotcha confirmed a second time: `react-hooks/set-state-in-effect`

Both the picker (resetting `naturalSize` when `uri` changes) and the modal (reseeding its draft
when `visible` flips to true) originally used a `useEffect` calling `setState` in the body. This
project's `react-hooks/set-state-in-effect` rule flags that. Same fix pattern already established
in `lobby-room-container.tsx` (see its own comments): compare the relevant key **during render**
and call `setState` directly in the render body, guarded by an inequality check against a
previously-seen value kept in its own bit of state — not inside `useEffect` at all. React explicitly
supports this ("adjusting state during rendering"); it bails out after one extra render instead of
round-tripping through the effect phase. Worth grepping for this pattern
(`lobby-room-container.tsx`'s "reseed block"/"handled key" comments) before reaching for a plain
`useEffect(() => setX(...), [dep])` anywhere in this codebase.

Verified: frontend `tsc --noEmit` clean, `eslint` 0 errors (411 pre-existing warnings, none new),
`pnpm jest` 71/71 suites — 1358/1358 tests, both on host and via the Docker CI-equivalent check
([[norma-verificacion-docker]]). No backend/DTO/contract change was needed — the wire shape didn't
move, only the frontend gesture/coordinate math and the UI flow.

Related: [[admin-locale-badge-and-focal-point-2026-09-13]], [[frontend-responsive-frutiger-aero]],
[[game-lobby-todo]], [[feedback_tooltips_norma]], [[norma-teclado]]
