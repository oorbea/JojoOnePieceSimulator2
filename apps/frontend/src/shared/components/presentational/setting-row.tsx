import type { ReactNode } from 'react'
import { XStack, YStack } from 'tamagui'

import { GlowText } from './glow-text'

type Props = {
  label: string
  /** Optional slot rendered right after the label - e.g. an `InfoHint`. */
  help?: ReactNode
  children: ReactNode
  /** Keeps label-above-control at every breakpoint instead of switching to
   * a row at `$md` - for a control that reads better close to its label
   * (e.g. a stepper's -/value/+ group) than pushed to the far end of a wide
   * row. Default `false` keeps every existing call site's layout (privacy,
   * allow bots, ...) unchanged. */
  stacked?: boolean
}

// A settings-form row: label (+ optional help icon) on one side, control on
// the other. Mobile stacks label above control, both left-aligned; `$md`
// switches to a real row with the control pushed to the end - unless
// `stacked` is set, which keeps the stacked layout at every breakpoint.
//
// Root cause this fixes: `GlossButton`'s own wrapper (`YStack items="center"
// justify="center"`, gloss-button.tsx) has no width of its own, so it
// stretches to fill whatever column it's given and centers the pill under a
// left-aligned label. Wrapping the control in `self="flex-start"` here
// keeps it sized to its content instead, without touching `GlossButton`
// itself (20+ call sites depend on its current sizing behavior).
export function SettingRow({ label, help, children, stacked = false }: Props) {
  return (
    <YStack
      width="100%"
      gap="$1.5"
      items="flex-start"
      {...(stacked ? {} : { $md: { flexDirection: 'row', justify: 'space-between', items: 'center', gap: '$3' } })}
    >
      <XStack items="center" gap="$1.5">
        <GlowText level="label">{label}</GlowText>
        {help}
      </XStack>
      <YStack self="flex-start" {...(stacked ? {} : { $md: { self: 'auto' } })}>
        {children}
      </YStack>
    </YStack>
  )
}
