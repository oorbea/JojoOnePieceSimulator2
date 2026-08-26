---
name: a11y-web-leak
description: "RN accessibility* / pointerEvents props leak raw to DOM on web via Tamagui, causing React console errors — use a11yProps() helper instead"
metadata: 
  node_type: memory
  type: project
  originSessionId: 5ba28674-b863-4d2d-9f04-de1e220738d0
  modified: 2026-08-04T12:10:31.123Z
---

Tamagui (2.6.2 in this repo) does not translate React Native's
`accessibilityLabel` / `accessibilityRole` / `accessibilityState` props to
ARIA (`aria-label` / `role` / `aria-disabled`) on the web target — it
forwards whatever prop it doesn't recognize straight through to the host
DOM element. React then logs `React does not recognize the \`X\` prop on a
DOM element` for each one. Same story for `pointerEvents="none"` passed as
a top-level prop instead of inside `style` — react-native-web's
`createDOMProps` explicitly warns `props.pointerEvents is deprecated. Use
style.pointerEvents` (source:
`react-native-web/src/modules/createDOMProps/index.js`).

**Why this matters:** these aren't crashes (page still renders), but they
look exactly like a real thrown error in Expo's terminal — Metro renders
`console.error` with a Code frame + Call Stack, indistinguishable at a
glance from an actual exception. Cost real debugging time chasing a
"crash" in `apps/frontend/src/shared/components/presentational/gloss-button.tsx`
that was just this DOM-attribute warning.

**How to apply / fix pattern:**
- Never pass `accessibilityLabel` / `accessibilityRole` / `accessibilityState`
  directly to a Tamagui component that also renders on web. Use
  `a11yProps(label, role, state)` from
  `apps/frontend/src/shared/lib/a11y.ts` instead — it branches on
  `Platform.OS` and returns the ARIA-flavored props on web, the RN props on
  native.
- Never pass `pointerEvents="none"` as a top-level JSX/styled-config prop.
  Put it inside `style`: `style={{ pointerEvents: 'none' }}` (merge with any
  existing style object rather than replacing it).
- When adding a new pressable/decorative component, grep first:
  `accessibility[A-Za-z]+=` and `pointerEvents=` under `apps/frontend/src`
  to catch any new leak before it ships.

Related: [[project_context]] (vips pipeline + auth flow are the other known
weak spots; this is the third recurring gotcha, specific to the frontend's
Tamagui/RN-web layer).
