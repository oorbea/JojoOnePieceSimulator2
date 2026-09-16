---
title: "Bugfix (2026-09-16): lobby config panel columns overflowed/overlapped on mobile"
tags:
  - project
  - jojo-onepiece-simulator
  - bugfix
  - frontend
  - gotcha
---

# Lobby config panel: `flexBasis` on a column stack becomes a height on mobile

Reported (Android, ca-ES, screenshot): the "Finestra de votació" stepper painted
over "Jugadors màxims", and "Permet bots" got cut off at the panel's bottom edge.

## Root cause

`create-lobby-screen.tsx` and `lobby-config-panel.tsx` gave the panel's two
columns an unconditional `flexBasis={320}`:

```tsx
<GlassPanel glossy p="$5" gap="$4" width="100%" $md={{ flexDirection: 'row' }}>
  <YStack flexBasis={320} grow={1} gap="$4">...</YStack>
  <YStack flexBasis={320} grow={1} gap="$4">...</YStack>
</GlassPanel>
```

`flexDirection: 'row'` only applies at `$md` (≥900px, see
[[frontend-responsive-frutiger-aero]]). Below that the panel is a **column**, so
`flexBasis` resolves against the main axis of a column container — i.e. it
becomes a *height* hint of 320px, not a width. Each column's real content
(several stepper rows) is taller than that, overflows its 320px box, and
`GlassPanel` is `overflow:'hidden'` — the second column's content painted over
by the first's overflow, and the last row of the second column (bots) got
clipped by the panel's own bottom edge.

Secondary: `SettingRow` (shared label+control row, used by every stepper,
toggle, privacy switch, bots switch) had no wrap/shrink control — the label
`XStack` and `GlowText` had neither `shrink` nor `minW={0}`, so a long label
(ca-ES/es-ES are the longest of the three locales) could push the control past
the column's edge instead of wrapping.

## Fix

Condition `flexBasis` on the same breakpoint that turns the panel into a row,
with `minW={0}` so the two 320px columns can still shrink in the 900–1100px
range instead of overflowing:

```tsx
<YStack grow={1} gap="$4" $md={{ flexBasis: 320, minW: 0 }}>...</YStack>
```

And in `setting-row.tsx`: `shrink={1} minW={0}` on the label wrapper, `shrink={1}`
on the label `GlowText`, `shrink={0}` on the `help` slot and the control
wrapper (so the fixed-size `InfoHint`/stepper never compresses), plus
`flexWrap: 'wrap'` on the `$md` row variant so the control drops to its own
line before it would overflow.

## Shorthand gotcha (typecheck)

`onlyAllowShorthands: true` on this project's Tamagui config (`@tamagui/config/v4`)
means `flexShrink`/`minWidth` as full prop names don't typecheck at all — the
shorthands are `shrink` and `minW` (see prior use in `power-block.tsx`,
`stand-card.tsx`, etc.). `position` is the one longhand exception already
documented in [[frontend-responsive-frutiger-aero]]; add `shrink`/`minW` to
that same "shorthand-only" list for future reference.

## Related feature: pressable number → digits-only edit

Same tanda added inline editing to `NumberStepper` (max players, voting
window, powers summary): tapping the value swaps it for a `GlossButton`-hosted
tap target with a tooltip (`game.create.editValueHint`), which opens a
`GlassField`-style Tamagui `Input` (`keyboardType="number-pad"`,
`inputMode="numeric"`, digits stripped via the existing
`text.replace(/[^0-9]/g, '')` idiom from `stage-form-modal.tsx`/
`character-form-modal.tsx`). Commit clamps to `[min, max]` (decided with the
owner: silent clamp, no error state); an empty draft reverts to the previous
value. Escape cancels without committing. A `committedRef` guard prevents
`onSubmitEditing` (Enter) and the trailing `onBlur` it triggers from both
firing a commit.

`VOTING_WINDOW_LIMITS`/`SUMMARY_DURATION_LIMITS` were added to `config-form.ts`
next to the existing `TEAM_SIZE_LIMITS`, mirroring
`apps/backend/internal/domain/entities/game/config.go`'s
`Min/MaxVotingWindowSeconds` (5–180) and `Min/MaxSummaryDurationSeconds`
(10–300) — not part of any generated contract (see
[[contratos-tipos-generados]]), so keep them in sync by hand if the backend
bounds change.

## Related

- [[bugfix-page-shell-scroll-clearance-2026-09-15]] — same family of bug
  (a layout prop behaving differently than assumed depending on flex
  direction/scroll context), different mechanism.
- [[frontend-responsive-frutiger-aero]] — breakpoint semantics, shorthand-only
  prop list.
- [[game-lobby-frontend]] — config-edit panel history this touches.
