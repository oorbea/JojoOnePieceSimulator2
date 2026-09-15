---
title: "Bugfix (2026-09-15): PageShell's nav clearance never applied on tall scrolled screens"
tags:
  - project
  - jojo-onepiece-simulator
  - bugfix
  - frontend
  - gotcha
---

# PageShell: `flex:1` inside a scroll container pins it to the viewport height

Reported: on mobile, scrolling to the bottom of a tall screen (lobby waiting room,
`StartBar`'s "abandonar sala"/"comenzar" buttons) left the last element covered by
the floating bottom dock — the reserved nav clearance (`useNavInsets()`, see
[[app-shell-y-navegacion]]) simply never showed up.

## Root cause

`page-shell.tsx`'s scroll branch had the padding on the **content YStack**, not the
`ScrollView`:

```tsx
<ScrollView flex={1} contentContainerStyle={{ flexGrow: 1 }}>
  <YStack flex={1} pt={navInsets.top} pb={navInsets.bottom}>...</YStack>
</ScrollView>
```

`flex={1}` means `flexBasis: 0` with no min-height floor — Yoga/RNW don't give it
`min-height: auto` the way plain block content gets on the web. That pins the YStack
to *exactly* the viewport height. Content taller than the viewport overflows straight
out of the box, past its own `paddingBottom` — the scroll extent is measured from the
overflowing content, not from the padded box, so the reserved clearance ends up
sitting at the viewport's bottom edge instead of after the last child. Nothing
separates the last button from the dock (`z:$nav`=500).

## Fix

Move the reservation to the thing the scroll extent is actually measured against —
the `ScrollView`'s own `contentContainerStyle` — and let the inner YStack just grow
to fit its content (`grow={1}` instead of `flex={1}`, i.e. `flexGrow:1` with the
default `flexBasis:'auto'`):

```tsx
<ScrollView
  flex={1}
  contentContainerStyle={{ flexGrow: 1, paddingTop: padTop, paddingBottom: padBottom }}
  onScroll={notifyScroll}
  scrollEventThrottle={16}
>
  <YStack grow={1} ...>...</YStack>
</ScrollView>
```

The non-scroll branch (`align="center"` short pages, no overflow risk) keeps `flex={1}`
+ `pt`/`pb` on the YStack as before — this only ever bit the `scroll` branch.

## Test gotcha

RN's `ScrollView` (`RCTScrollView` in the jest test renderer) carries
`contentContainerStyle` as its **own separate prop**, not merged into `style`. A test
helper (`page-shell.test.tsx`'s `styleOf`) that only reads `node.props.style` silently
sees no padding at all on the ScrollView node and has to also spread in
`node.props.contentContainerStyle`.

**Norm going forward:** nav clearance (or any reservation meant to survive scrolling)
belongs on the scroll container's `contentContainerStyle`, never on a `flex:1` child
living inside it. See also the `VoteBar` `position:sticky` variant of the same family
of bug in [[bugfixes-partida-2026-09-14]] §1 — different mechanism (sticky resolves
against the scrollport, not `pb`), same lesson: anything anchored relative to the
viewport inside a scrollable must be checked against the *actual* scroll/content
geometry, not assumed to inherit a parent's padding.

## Related

- [[app-shell-y-navegacion]] — the measured-inset machinery this padding consumes
  (unchanged by this fix).
- [[bugfixes-partida-2026-09-14]] — prior occurrence of a bottom-anchored element
  escaping `PageShell`'s clearance, different root cause.
- [[frontend-responsive-frutiger-aero]] — z-index token order, breakpoints.
