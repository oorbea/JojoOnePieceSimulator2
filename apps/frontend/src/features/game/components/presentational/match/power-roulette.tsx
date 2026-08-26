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

import { buildReel, landingTiming, restRows, WINDOW_ROWS } from '@/features/game/lib/reel-geometry'
import { GlowText } from '@/shared/components/presentational/glow-text'

const ITEM_HEIGHT = 34
// The window always shows WINDOW_ROWS rows (above / landed / below) so the
// reel reads as a real slot machine instead of a single value popping in.
// The landed row is the MIDDLE one - see restY below.
const WINDOW_HEIGHT = ITEM_HEIGHT * WINDOW_ROWS
// How far past the resting frame the reel overshoots before catching, the
// "catch" that makes it read as physically decelerating instead of gliding
// to a stop. Small and row-relative on purpose - big enough overshoot here
// used to bleed a neighbouring row's label into the highlight band.
const OVERSHOOT_PX = ITEM_HEIGHT * 0.2
// The catch leg's fixed duration - a bounded withTiming, not a physics
// withSpring, specifically so delay + decel + catch always completes
// within spinMs. See lib/reel-geometry.ts's `landingTiming` for the timing
// split and why an unbounded spring couldn't guarantee that.
const CATCH_MS = 180

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
// then catches on it with two bounded withTiming legs (never a physics
// spring - see CATCH_MS's doc) - with a brief highlight flash + pop on the
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
      // Two bounded withTiming legs, not a withTiming+withSpring pair: a
      // physics spring's settle time isn't easily bounded (at these
      // damping/stiffness values it can run ~900ms), but delay+duration
      // always equals spinMs exactly, leaving as little as ~15% of that for
      // the catch - nowhere near enough for a spring to settle before the
      // phase machine (use-loadout-reveal.ts's own fixed timers) cuts to
      // 'land' and hard-resets translateY regardless. landingTiming's
      // deterministic decelMs+CATCH_MS always finishes in time, landing
      // exactly on restY with no residual velocity for that reset to
      // visibly interrupt - see lib/__tests__/reel-geometry.test.ts's
      // invariant test for the exact bound this guarantees.
      const { delayMs, decelMs, catchMs } = landingTiming(spinMs, laneIndex, CATCH_MS)
      translateY.value = withDelay(
        delayMs,
        withSequence(
          withTiming(restY - OVERSHOOT_PX, { duration: decelMs, easing: Easing.out(Easing.cubic) }),
          withTiming(restY, { duration: catchMs, easing: Easing.out(Easing.quad) })
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

  // translateY-only: applied to the inner scrolling strip.
  const reelStyle = useAnimatedStyle(() => ({
    transform: [{ translateY: translateY.value }],
  }))
  // scale-only: applied to the OUTER window wrapper, not the strip - the
  // strip's own centre can be many rows away from the visible window (a
  // 26-item Stand/DevilFruit reel is ~884px tall), so scaling it around its
  // own centre visibly displaced the landed answer sideways-in-time by up
  // to ~1.8 rows before settling back, reading as "still spinning toward
  // another power." The window itself is small and fixed, and its centre
  // is exactly the highlight band's centre, so scaling it pops the whole
  // visible strip symmetrically around the answer instead of past it.
  const popStyle = useAnimatedStyle(() => ({
    transform: [{ scale: scale.value }],
  }))
  const flashStyle = useAnimatedStyle(() => ({
    opacity: 0.3 + flash.value * 0.7,
  }))

  return (
    <Animated.View style={popStyle}>
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
    </Animated.View>
  )
}
