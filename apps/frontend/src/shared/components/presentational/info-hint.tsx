import { HelpCircle } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { YStack } from 'tamagui'

import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'
import { SpeechBubble } from './speech-bubble'

type Props = { text: string }

// Small '?' affordance for explaining a CONCEPT (what a field/option means)
// - not a control's current state, which is already covered by the inline
// hint text under a field (e.g. `modeGauntletHint`). Doubles as a
// `GlossButton` tooltip (hover/focus on web, long-press on native, via the
// `tooltip` prop) AND a tap-to-toggle `SpeechBubble`, so the explanation is
// reachable with a single tap on every platform, not just hover-capable
// ones. `btnSize="sm"` matches every other icon-only circle in the app
// (back/refresh/stepper buttons) rather than introducing a smaller variant.
export function InfoHint({ text }: Props) {
  const [open, setOpen] = useState(false)

  return (
    <YStack position="relative">
      <GlossButton
        tone="glass"
        btnSize="sm"
        shape="circle"
        tooltip={text}
        onPress={() => setOpen((v) => !v)}
        accessibilityLabel={text}
      >
        <HelpCircle size={14} color="$panelTextSoft" />
      </GlossButton>
      {open ? (
        <YStack position="absolute" t="100%" l={0} mt="$2" z="$overlay" width={220}>
          <SpeechBubble tailSide="top" p="$3">
            <GlowText level="label">{text}</GlowText>
          </SpeechBubble>
        </YStack>
      ) : null}
    </YStack>
  )
}
