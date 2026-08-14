import { useEffect, useMemo } from 'react'
import Animated, {
  Easing,
  useAnimatedStyle,
  useSharedValue,
  withSequence,
  withSpring,
  withTiming,
} from 'react-native-reanimated'
import { YStack } from 'tamagui'

import { GlowText } from '@/shared/components/presentational/glow-text'

const ITEM_HEIGHT = 26
// How many extra candidate ticks scroll past before landing - varies by the
// candidate pool size so a short list (e.g. 4 haki levels) doesn't look like
// it's repeating the same handful of names twice as fast as a long one
// (dozens of Stands). The final tick is always the real answer.
function tickCount(poolSize: number): number {
  return Math.max(10, Math.min(24, poolSize * 3))
}

type Props = {
  /** Decorative names/labels to flash past while spinning - never the
   * actual answer, which is always `finalLabel`. */
  candidates: string[]
  finalLabel: string
  /** true while this slot's ruleta should be spinning (the 'spin' phase);
   * false once it should show finalLabel at rest (the 'land' phase). */
  spinning: boolean
  reducedMotion: boolean
  spinMs: number
}

// A Wii Party-style vertical slot-reel: a strip of candidate labels scrolls
// past behind a fixed window, decelerating into the real answer with a
// small overshoot bounce on landing. Reduced motion skips straight to the
// resting frame - no scroll, no bounce, but still exactly finalLabel.
export function PowerRoulette({ candidates, finalLabel, spinning, reducedMotion, spinMs }: Props) {
  const translateY = useSharedValue(0)
  const scale = useSharedValue(1)

  const reel = useMemo(() => {
    if (candidates.length === 0) return [finalLabel]
    const ticks = tickCount(candidates.length)
    const items: string[] = []
    for (let i = 0; i < ticks; i++) items.push(candidates[i % candidates.length])
    items.push(finalLabel)
    return items
  }, [candidates, finalLabel])

  const restY = -(reel.length - 1) * ITEM_HEIGHT

  useEffect(() => {
    if (reducedMotion) {
      translateY.value = restY
      scale.value = 1
      return
    }
    if (spinning) {
      translateY.value = 0
      translateY.value = withTiming(restY, { duration: spinMs, easing: Easing.out(Easing.cubic) })
      scale.value = 1
    } else {
      translateY.value = restY
      scale.value = withSequence(withTiming(1.16, { duration: 130 }), withSpring(1, { damping: 7, stiffness: 220 }))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- translateY/scale are stable shared values, not reactive deps
  }, [spinning, reducedMotion, restY, spinMs])

  const style = useAnimatedStyle(() => ({
    transform: [{ translateY: translateY.value }, { scale: scale.value }],
  }))

  return (
    <YStack height={ITEM_HEIGHT} width="100%" overflow="hidden" items="center" justify="center">
      <Animated.View style={style}>
        {reel.map((label, i) => (
          <YStack key={i} height={ITEM_HEIGHT} items="center" justify="center">
            <GlowText level="label" fontSize="$2" numberOfLines={1}>
              {label}
            </GlowText>
          </YStack>
        ))}
      </Animated.View>
    </YStack>
  )
}
