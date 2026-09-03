---
title: "Feature: sliding top-nav indicator (2026-09-03)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - feature
  - animation
---

# Sliding top-nav indicator (2026-09-03)

Before this, the active tab in `AppShell`'s top bar/bottom dock was a flat `bg: '$channelActive'`
Tamagui variant on `ChannelBarItemFrame` - it popped on/off with no transition. Owner asked for a
slide, like the light-blue Wii-channel selection cue moving between Home/Play/Profile/Admin.

## Status

Done, both the top bar and the bottom dock. See [[catalogo-publico-stands-devil-fruits-stages]]
for the unrelated catalogue feature shipped in the same pass, and
[[frontend-responsive-frutiger-aero]] for the general animation constraints this had to fit.

## What animates, and the deliberate rule exception

New `shared/components/presentational/channel-bar-indicator.tsx`: one absolutely-positioned
`Animated.View` (Reanimated), driven by `useSharedValue`/`useAnimatedStyle` on the UI thread -
`transform: [{translateX}, {translateY}]` **and `width`**.

Animating `width` is a deliberate exception to this repo's "only transform/opacity animate" rule
(`frontend-responsive-frutiger-aero.md`, restated in `gameplay-power-fx.md`). Reasoning it was
accepted here: it's one 48px-tall pill, not a per-frame FX category; the alternative
(`scaleX` instead of animating `width` directly) visibly squashes the pill's rounded ends while
sliding between items of different widths (top bar's label+icon items aren't all the same size,
unlike the dock's uniform icon-only cells). If this pattern gets copied elsewhere, re-litigate
whether it should stay an exception or become a documented second case.

`ChannelBarItemFrame`'s `active` variant is now a no-op (`{ true: {} }`) - kept only so callers
don't need to drop the `active` prop; the colour now comes entirely from the indicator layered
underneath. `item.active` still governs the icon/label colour swap (white/`$panelText`,
`onColor`/`ink` `GlowText` tone) independently of the indicator.

## Measurement: the coordinate-space trap

Each item's box is measured via `onLayout` on a wrapper (top bar) or the existing per-item cell
(dock), keyed by `href`, stored in two separate `Record<href, layout>` maps in `AppShell` (top bar
and dock never coexist - `showTopLinks` guarantees exactly one is mounted - but they'd clash on
the same hrefs if merged into one map).

**The bug this avoids**: `onLayout`'s `x`/`y` are relative to the item's *immediate* parent, not
to whatever ancestor the absolutely-positioned indicator is placed against. The top bar's items
sit inside an inner `<XStack gap="$2">`, itself offset within `ChannelBarFrame` by the
logo/title/spacer before it - measuring against `ChannelBarFrame` while positioning the indicator
there too would put the pill in the wrong place. Fix: give that inner `XStack` its own
`position="relative"` and mount `ChannelBarIndicator` **inside** it, as a sibling of the measured
item wrappers - now both share the same coordinate space by construction. The bottom dock doesn't
have this problem: its items are already direct children of `ChannelBarFrame` with nothing before
them, so the indicator can sit there directly.

Re-measured on every `onLayout` (bar wraps to two rows, viewport resizes) - no coordinates are
ever cached across renders, per the recurring "coordinates captured once, never refreshed" trap
that motivated `shared/lib/scroll-bus.ts`.

## Z-order

`ChannelBarFrame`'s own `GlossOverlay` sheen sits at token `z: '$gloss'` (5); `ChannelBarItemFrame`
sits at `z: '$content'` (10). The indicator uses an explicit numeric `zIndex: 6` - above the gloss
(so its fill colour isn't dulled by the sheen, matching how the old `bg`-on-item approach looked,
since items were already above the gloss) but below the item content (so label/icon stay legible
on top of the pill). RN only compares `zIndex` between direct siblings, which is why the indicator
must live in the same flex container as the items it's stacking against, not a level up or down.

## Reduced motion and first paint

`useReducedMotion()` collapses every update to an instant snap (no `withTiming`), same contract as
`gameplay-power-fx.md`. The very first non-null layout for a session also snaps instead of
sliding in from `(0,0)` - a `hasMeasuredOnce` ref inside the indicator gates animation vs. instant
set, independent of the reduced-motion check.

## Things to watch

- If a 6th nav item is ever added and item widths shrink further, re-check that `zIndex: 6` still
  reads correctly against `$gloss`/`$content` if those token values ever change - it's a plain
  number, not a token reference, since Reanimated's raw style object can't consume Tamagui tokens
  directly (this file resolves `$channelActive`'s actual colour via `useTheme().channelActive?.val`
  instead, same pattern as `glass-field.tsx`'s placeholder colour).
