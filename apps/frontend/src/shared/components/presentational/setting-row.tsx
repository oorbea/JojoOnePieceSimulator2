import type { ReactNode } from 'react'
import { XStack, YStack } from 'tamagui'

import { GlowText } from './glow-text'

type Props = {
  label: string
  /** Optional slot rendered right after the label - e.g. an `InfoHint`. */
  help?: ReactNode
  children: ReactNode
}

// A settings-form row: label (+ optional help icon) on one side, control on
// the other. Mobile stacks label above control, both left-aligned; `$md`
// switches to a real row with the control pushed to the end.
//
// Root cause this fixes: `GlossButton`'s own wrapper (`YStack items="center"
// justify="center"`, gloss-button.tsx) has no width of its own, so it
// stretches to fill whatever column it's given and centers the pill under a
// left-aligned label. Wrapping the control in `self="flex-start"` here
// keeps it sized to its content instead, without touching `GlossButton`
// itself (20+ call sites depend on its current sizing behavior).
export function SettingRow({ label, help, children }: Props) {
  return (
    <YStack
      width="100%"
      gap="$1.5"
      items="flex-start"
      $md={{ flexDirection: 'row', justify: 'space-between', items: 'center', gap: '$3' }}
    >
      <XStack items="center" gap="$1.5">
        <GlowText level="label">{label}</GlowText>
        {help}
      </XStack>
      <YStack self="flex-start" $md={{ self: 'auto' }}>
        {children}
      </YStack>
    </YStack>
  )
}
