---
title: "Tamagui leaks RN accessibility*/pointerEvents props raw to the DOM on web"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - gotcha
---

# Tamagui leaks RN a11y/pointerEvents props to the DOM on web

Tamagui (2.6.2 in this repo) does not translate React Native's `accessibilityLabel` /
`accessibilityRole` / `accessibilityState` props to ARIA (`aria-label` / `role` / `aria-disabled`)
on the web target — it forwards whatever prop it doesn't recognize straight through to the host DOM
element, and React logs `React does not recognize the \`X\` prop on a DOM element` for each one.
Same story for `pointerEvents="none"` passed as a top-level prop instead of inside `style` —
react-native-web's `createDOMProps` explicitly warns `props.pointerEvents is deprecated. Use
style.pointerEvents`.

**Why it matters:** not a crash, but Metro renders these `console.error`s with a code frame + call
stack, indistinguishable at a glance from a real thrown exception — cost real debugging time once
chasing a "crash" in `gloss-button.tsx` that was just this DOM-attribute warning.

## How to apply

- Never pass `accessibilityLabel` / `accessibilityRole` / `accessibilityState` directly to a
  Tamagui component that also renders on web. Use `a11yProps(label, role, state)` from
  `apps/frontend/src/shared/lib/a11y.ts` — it branches on `Platform.OS` and returns ARIA-flavored
  props on web, RN props on native.
- Never pass `pointerEvents="none"` as a top-level prop. Put it inside `style`:
  `style={{ pointerEvents: 'none' }}` (merge with any existing style object).
- When adding a new pressable/decorative component, grep first: `accessibility[A-Za-z]+=` and
  `pointerEvents=` under `apps/frontend/src` to catch a new leak before it ships.

Related: [[frontend-responsive-frutiger-aero]] (same primitive layer, same `gloss-button.tsx` this
was first noticed in), [[norma-tooltips-y-ayuda-contextual]] (the tooltip primitive spreads its
hover/focus/long-press handlers onto the same pressable for the exact same reason this note gives for
`pointerEvents`/`a11yProps` - wrapping in an extra `Pressable` would steal the tap).
