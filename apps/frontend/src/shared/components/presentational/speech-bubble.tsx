import { YStack } from 'tamagui'

import { isWeb } from '@/shared/lib/web-blur'

import { GlassPanel } from './glass-panel'

type TailSide = 'top' | 'bottom' | 'left' | 'right'

// Every real call site uses this as a plain centered dialog (subtitle bubble,
// error message, help popover) - a tail fixed at a literal 18%/40% offset
// only lined up by coincidence for one narrow box width and looked off-center
// for any other (full-width bubbles especially). Centered on both axes
// instead, same web-only `transform` trick `TooltipBubble` uses (RN's native
// transform doesn't take percentage values, so native anchors from the
// midpoint without the final nudge - fine for an 18px tail).
// On web, the tail's `rotate="45deg"` (Tamagui prop, applied to the element
// below) and this `style.transform` would otherwise fight over the same
// underlying `transform` style key - RN's style flattening takes the LAST
// object wholesale per key, it doesn't merge transform arrays. Baking
// `rotate` into the same array here, web-only, sidesteps that: this object
// fully replaces the rotate-only one Tamagui generates, but it already
// includes the same rotation, so the visual result is identical plus the
// centering nudge. Native never sets `style` here, so its `rotate` prop
// applies completely undisturbed.
const TAIL_POSITION: Record<TailSide, object> = {
  top: {
    t: -9,
    l: '50%',
    borderTopWidth: 1.5,
    borderLeftWidth: 1.5,
    style: isWeb ? { transform: [{ rotate: '45deg' }, { translateX: '-50%' }] } : undefined,
  },
  bottom: {
    b: -9,
    l: '50%',
    borderBottomWidth: 1.5,
    borderRightWidth: 1.5,
    style: isWeb ? { transform: [{ rotate: '45deg' }, { translateX: '-50%' }] } : undefined,
  },
  left: {
    l: -9,
    t: '50%',
    borderTopWidth: 1.5,
    borderLeftWidth: 1.5,
    style: isWeb ? { transform: [{ rotate: '45deg' }, { translateY: '-50%' }] } : undefined,
  },
  right: {
    r: -9,
    t: '50%',
    borderBottomWidth: 1.5,
    borderRightWidth: 1.5,
    style: isWeb ? { transform: [{ rotate: '45deg' }, { translateY: '-50%' }] } : undefined,
  },
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
