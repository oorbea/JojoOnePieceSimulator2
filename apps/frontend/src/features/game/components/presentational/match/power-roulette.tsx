import { useEffect, useMemo } from 'react'
import { LinearGradient } from '@tamagui/linear-gradient'
import Animated, {
  Easing,
  useAnimatedStyle,
  useSharedValue,
  withDelay,
  withSequence,
  withSpring,
  withTiming,
} from 'react-native-reanimated'
import { YStack } from 'tamagui'

import { buildReel, WINDOW_ROWS, restRows } from '@/features/game/lib/reel-geometry'
import { GlowText } from '@/shared/components/presentational/glow-text'

const ITEM_HEIGHT = 34
// The window always shows WINDOW_ROWS rows (above / landed / below) so the
// reel reads as a real slot machine instead of a single value popping in.
// The landed row is the MIDDLE one - see restY below.
const WINDOW_HEIGHT = ITEM_HEIGHT * WINDOW_ROWS
// How far past the resting frame the reel overshoots before springing back,
// the "catch" that makes it read as physically decelerating instead of
// gliding to a stop.
const OVERSHOOT_PX = ITEM_HEIGHT * 0.35
// Per-lane stagger so every carril doesn't stop on the exact same frame -
// capped at 30% of the spin budget so even the last lane's timing leg still
// starts (and, overshoot aside, finishes) inside spinMs.
const MAX_STAGGER_MS = 70
const MAX_STAGGER_SHARE = 0.3

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
  /** This lane's position among its siblings, purely for the landing
   * stagger below - has no bearing on which value is drawn. */
  laneIndex?: number
}

// A Wii Party-style vertical slot-reel: a 3-row window (landed value
// centered, its neighbours bleeding into a soft top/bottom fade) scrolls a
// strip of candidate labels past, overshoots slightly past the real answer,
// then springs back to catch on it - with a brief highlight flash on the
// centre band when it lands. Reduced motion skips straight to the resting
// frame - no scroll, no overshoot, no stagger, but still exactly finalLabel
// centered in the window.
export function PowerRoulette({
  candidates,
  finalLabel,
  spinning,
  reducedMotion,
  spinMs,
  laneIndex = 0,
}: Props) {
  const translateY = useSharedValue(0)
  const scale = useSharedValue(1)
  const flash = useSharedValue(0)

  const reel = useMemo(() => buildReel(candidates, finalLabel), [candidates, finalLabel])

  // Resting position: the window shows items [reel.length-WINDOW_ROWS, ...,
  // reel.length-1], i.e. the FINAL label (at reel.length-2) sits in the
  // centre row. See lib/reel-geometry.ts for the invariant this maintains.
  const restY = restRows(reel.length) * ITEM_HEIGHT

  useEffect(() => {
    if (reducedMotion) {
      translateY.value = restY
      scale.value = 1
      flash.value = 0
      return
    }
    if (spinning) {
      translateY.value = 0
      scale.value = 1
      flash.value = 0
      const delay = Math.min(laneIndex * MAX_STAGGER_MS, spinMs * MAX_STAGGER_SHARE)
      const duration = Math.max(spinMs - delay, spinMs * 0.5)
      translateY.value = withDelay(
        delay,
        withSequence(
          withTiming(restY - OVERSHOOT_PX, {
            duration: duration * 0.85,
            easing: Easing.out(Easing.cubic),
          }),
          withSpring(restY, { damping: 9, stiffness: 260 })
        )
      )
    } else {
      translateY.value = restY
      scale.value = withSequence(
        withTiming(1.16, { duration: 130 }),
        withSpring(1, { damping: 7, stiffness: 220 })
      )
      flash.value = withSequence(withTiming(1, { duration: 140 }), withTiming(0, { duration: 460 }))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- translateY/scale/flash are stable shared values, not reactive deps
  }, [spinning, reducedMotion, restY, spinMs, laneIndex])

  const reelStyle = useAnimatedStyle(() => ({
    transform: [{ translateY: translateY.value }, { scale: scale.value }],
  }))
  const flashStyle = useAnimatedStyle(() => ({
    opacity: 0.3 + flash.value * 0.7,
  }))

  return (
    <YStack
      height={WINDOW_HEIGHT}
      width="100%"
      overflow="hidden"
      position="relative"
      rounded="$card"
    >
      <Animated.View style={reelStyle}>
        {reel.map((label, i) => (
          <YStack key={i} height={ITEM_HEIGHT} items="center" justify="center">
            <GlowText level="label" fontSize="$2" numberOfLines={1}>
              {label}
            </GlowText>
          </YStack>
        ))}
      </Animated.View>

      {/* Centre-row highlight band - marks where the answer lands, and
          pulses brighter on the land beat above. Purely decorative, never
          intercepts touch/hover. */}
      <Animated.View
        style={[
          {
            position: 'absolute',
            left: 0,
            right: 0,
            top: ITEM_HEIGHT,
            height: ITEM_HEIGHT,
            borderTopWidth: 1,
            borderBottomWidth: 1,
            borderColor: 'rgba(120,180,255,0.55)',
            backgroundColor: 'rgba(120,180,255,0.12)',
            pointerEvents: 'none',
          },
          flashStyle,
        ]}
      />

      {/* Top/bottom fade - the strip enters/leaves instead of being hard
          clipped by the window edge. */}
      <LinearGradient
        position="absolute"
        t={0}
        l={0}
        r={0}
        height={ITEM_HEIGHT * 0.8}
        colors={['rgba(10,12,20,0.4)', 'rgba(10,12,20,0)']}
        start={[0, 0]}
        end={[0, 1]}
        style={{ pointerEvents: 'none' } as object}
      />
      <LinearGradient
        position="absolute"
        b={0}
        l={0}
        r={0}
        height={ITEM_HEIGHT * 0.8}
        colors={['rgba(10,12,20,0)', 'rgba(10,12,20,0.4)']}
        start={[0, 0]}
        end={[0, 1]}
        style={{ pointerEvents: 'none' } as object}
      />
    </YStack>
  )
}
