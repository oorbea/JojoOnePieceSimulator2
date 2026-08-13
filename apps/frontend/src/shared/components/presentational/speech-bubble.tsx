import { YStack } from 'tamagui'

import { GlassPanel } from './glass-panel'

type TailSide = 'top' | 'bottom' | 'left' | 'right'

const TAIL_POSITION: Record<TailSide, object> = {
  top: { t: -9, l: '18%', borderTopWidth: 1.5, borderLeftWidth: 1.5 },
  bottom: { b: -9, l: '18%', borderBottomWidth: 1.5, borderRightWidth: 1.5 },
  left: { l: -9, t: '40%', borderTopWidth: 1.5, borderLeftWidth: 1.5 },
  right: { r: -9, t: '40%', borderBottomWidth: 1.5, borderRightWidth: 1.5 },
}

type SpeechBubbleProps = React.ComponentProps<typeof GlassPanel> & {
  tailSide?: TailSide
}

// A Wii-style dialogue bubble. RN has no CSS triangle borders, so the tail
// is a small rotated square with two borders drawn — it must live OUTSIDE
// the panel's `overflow: hidden`, so this wraps GlassPanel and the tail as
// relative siblings rather than nesting the tail inside the panel.
export function SpeechBubble({ tailSide = 'bottom', children, ...rest }: SpeechBubbleProps) {
  return (
    <YStack position="relative">
      <GlassPanel radiusSize="bubble" p="$4" gap="$3" items="flex-start" {...rest}>
        {children}
      </GlassPanel>
      <YStack
        position="absolute"
        width={18}
        height={18}
        rotate="45deg"
        bg="$glassFill"
        borderColor="$glassEdge"
        {...TAIL_POSITION[tailSide]}
      />
    </YStack>
  )
}
