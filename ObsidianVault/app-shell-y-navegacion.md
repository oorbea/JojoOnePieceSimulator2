---
title: "App shell & navigation"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - reference
---

# App shell & navigation

Atomic reference note for `shared/components/presentational/app-shell.tsx` /
`channel-bar.tsx` - this knowledge used to be scattered across
[[frontend-responsive-frutiger-aero]] and [[game-lobby-frontend]], which made
it slow to find. See [[nav-indicador-deslizante]] for the sliding-indicator
feature built on top of this shape.

## Shape

Two floating glass pills ("channels", Wii-menu styled) over `AquaBackground`:
a **top bar** (logo, title, nav links, theme toggle, logout) on wide screens,
or a **bottom dock** (icon-only nav items) on narrow ones - never both, never
neither. One boolean, `showTopLinks = media.md`, decides which; this used to
be two independent `$md` checks (a real bug source, see
`app-shell.test.tsx`'s regression-target comment) and is now structurally one.

Mounted once, inside `app/(app)/_layout.tsx`'s auth guard - every
authenticated route inherits it via `<AppShell>{children}</AppShell>`.

`AppShell` measures both bars' real rendered height via `onLayout` and
publishes the reservation through `NavInsetsProvider` (`shared/lib/
nav-insets.ts`), so `PageShell`'s clearance always matches what's actually on
screen - a bar that wraps onto two rows never ends up covering page content.

`ChannelBar` (`dock="top"|"bottom"|"static"`) is the reusable glass-pill host:
docking wraps the (self-centering, in-normal-flow) pill in a thin absolute
centering layer that never captures touches outside the pill's own bounds
(see the doc comment on `ChannelBar` for the `left:0;right:0` vs
`maxW:1080;self:center` bug this fixed).

## Home's channel grid is the other nav surface

`features/home/components/presentational/home-screen.tsx`'s `CHANNELS` array
+ `ChannelTile` grid is a second, separate navigation surface - the "pick a
channel" menu on the home screen, distinct from the persistent top
bar/dock. As of 2026-09-03 every tile is unlocked and routes somewhere real
(`/play`, `/profile`, `/catalog/stands`, `/catalog/devil-fruits`,
`/catalog/stages`) - see [[catalogo-publico-stands-devil-fruits-stages]].

## Things to know before touching either

- `ChannelBarFrame` is `flexWrap:'wrap'` + `overflow:'hidden'` - it grows
  instead of clipping when content doesn't fit one row, so anything
  measuring child layout inside it must re-measure on every `onLayout`, never
  cache across renders.
- `ChannelBarItem` attaches its tooltip trigger ref directly to its own host
  node (`.measure()` for the bubble's anchor) - wrapping it in an extra
  `Pressable` would steal the touch; a non-interactive wrapper `View`/`YStack`
  is fine.
- `GlossButton` (used elsewhere, not `ChannelBarItem` itself) isn't
  `forwardRef` - expect the same ref friction there if you ever need to
  measure one (`InfoHint` wraps it in a `View ref={...} collapsable={false}`).
