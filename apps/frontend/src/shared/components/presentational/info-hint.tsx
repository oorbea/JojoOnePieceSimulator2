import { HelpCircle } from '@tamagui/lucide-icons-2'
import { useRef, useState } from 'react'
import { Modal, View } from 'react-native'
import { YStack } from 'tamagui'

import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'
import { SpeechBubble } from './speech-bubble'

type Props = { text: string }

type Anchor = { x: number; y: number; width: number; height: number }

// Small '?' affordance for explaining a CONCEPT (what a field/option means)
// - not a control's current state, which is already covered by the inline
// hint text under a field (e.g. `modeGauntletHint`). Doubles as a
// `GlossButton` tooltip (hover/focus on web, long-press on native, via the
// `tooltip` prop) AND a tap-to-toggle popover, so the explanation is
// reachable with a single tap on every platform, not just hover-capable
// ones.
//
// The popover renders through a root-level `Modal`, anchored to the '?'
// icon's measured screen position - same reasoning as `TooltipBubble`
// (tooltip.tsx): nested deep inside a form (often itself inside a
// `FilterDisclosure`/`GlassPanel` with its own `overflow:hidden`), a plain
// `position:absolute` popover would get clipped or painted over by later,
// unrelated siblings. A plain `View` wraps the button only so `.measure()`
// has a host node to call — `GlossButton` itself isn't `forwardRef`.
// `collapsable={false}` keeps Android from optimizing this wrapper `View`
// out of the native tree, which would make it unmeasurable.
export function InfoHint({ text }: Props) {
  const [open, setOpen] = useState(false)
  const [anchor, setAnchor] = useState<Anchor | null>(null)
  const anchorRef = useRef<View>(null)

  const toggle = () => {
    if (open) {
      setOpen(false)
      return
    }
    anchorRef.current?.measure((_x, _y, width, height, pageX, pageY) => {
      setAnchor({ x: pageX, y: pageY, width, height })
      setOpen(true)
    })
  }

  return (
    <View ref={anchorRef} collapsable={false}>
      <GlossButton
        tone="glass"
        btnSize="sm"
        shape="circle"
        tooltip={text}
        onPress={toggle}
        accessibilityLabel={text}
      >
        <HelpCircle size={14} color="$panelTextSoft" />
      </GlossButton>
      {open && anchor ? (
        <Modal visible transparent animationType="fade" onRequestClose={() => setOpen(false)} statusBarTranslucent>
          <YStack flex={1} onPress={() => setOpen(false)}>
            <YStack
              position="absolute"
              t={anchor.y + anchor.height + 8}
              l={Math.max(8, anchor.x + anchor.width / 2 - 110)}
              width={220}
            >
              <SpeechBubble tailSide="top" p="$3" onPress={(e: { stopPropagation?: () => void }) => e.stopPropagation?.()}>
                <GlowText level="label">{text}</GlowText>
              </SpeechBubble>
            </YStack>
          </YStack>
        </Modal>
      ) : null}
    </View>
  )
}
