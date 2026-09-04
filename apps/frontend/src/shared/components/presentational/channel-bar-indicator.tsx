import { useEffect, useRef } from 'react'
import Animated, { Easing, useAnimatedStyle, useSharedValue, withTiming } from 'react-native-reanimated'
import { useTheme } from 'tamagui'

import { useReducedMotion } from '@/shared/hooks/use-reduced-motion'

export type ChannelBarIndicatorLayout = { x: number; y: number; width: number; height: number }

type Props = {
  layout: ChannelBarIndicatorLayout | null
}

const SLIDE_DURATION_MS = 220

// The sliding "you are here" pill behind the active top-bar/dock item.
// Reanimated drives translateX/translateY + width on the UI thread, no JS
// re-render per frame. Animating `width` is a deliberate exception to this
// repo's usual "only transform/opacity animate" rule (see
// nav-indicador-deslizante.md in the vault): this is one 48px pill, not a
// per-frame FX, and a `scaleX` stand-in would visibly squash its rounded
// ends as it stretches between differently-sized items.
//
// `layout` is measured by the caller (ChannelBar) via onLayout on each item
// - this component only owns interpolating between whatever layout it's
// handed. The first non-null layout snaps instantly (opacity fades in from
// 0 without a slide from the origin); every layout after that animates,
// unless reduced motion is on, in which case every update snaps.
export function ChannelBarIndicator({ layout }: Props) {
  const theme = useTheme()
  const reducedMotion = useReducedMotion()
  const x = useSharedValue(0)
  const y = useSharedValue(0)
  const width = useSharedValue(0)
  const opacity = useSharedValue(0)
  const hasMeasuredOnce = useRef(false)

  useEffect(() => {
    if (!layout) return
    const shouldAnimate = hasMeasuredOnce.current && !reducedMotion
    const timing = { duration: SLIDE_DURATION_MS, easing: Easing.out(Easing.cubic) }
    x.value = shouldAnimate ? withTiming(layout.x, timing) : layout.x
    y.value = shouldAnimate ? withTiming(layout.y, timing) : layout.y
    width.value = shouldAnimate ? withTiming(layout.width, timing) : layout.width
    opacity.value = shouldAnimate ? withTiming(1, timing) : 1
    hasMeasuredOnce.current = true
  }, [layout, reducedMotion, x, y, width, opacity])

  const style = useAnimatedStyle(() => ({
    transform: [{ translateX: x.value }, { translateY: y.value }],
    width: width.value,
    opacity: opacity.value,
  }))

  if (!layout) return null

  return (
    <Animated.View
      pointerEvents="none"
      style={[
        {
          position: 'absolute',
          top: 0,
          left: 0,
          height: layout.height,
          borderRadius: 9999,
          backgroundColor: theme.channelActive?.val,
          // Above the bar's own GlossOverlay ($gloss = 5, so the pill's
          // color isn't dulled by the sheen) but below ChannelBarItem's
          // `z: '$content'` (10) so the item's own label/icon stay on top.
          zIndex: 6,
        },
        style,
      ]}
    />
  )
}
