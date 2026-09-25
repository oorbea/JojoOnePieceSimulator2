import { Sparkles } from '@tamagui/lucide-icons-2'
import { LinearGradient } from '@tamagui/linear-gradient'
import { useEffect } from 'react'
import Animated, {
  Easing,
  useAnimatedStyle,
  useSharedValue,
  withSequence,
  withTiming,
} from 'react-native-reanimated'
import { YStack } from 'tamagui'

import { caseStripRestX, type CaseStripCard } from '@/features/game/lib/case-strip'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { LazyImage } from '@/shared/components/presentational/lazy-image'
import { asToken } from '@/shared/lib/tamagui-token'

export const CARD_WIDTH = 88
export const CARD_GAP = 10
const CARD_HEIGHT = 108
const RARITY_BAR_HEIGHT = 5
const VISIBLE_CARDS = 5
export const WINDOW_WIDTH = CARD_WIDTH * VISIBLE_CARDS + CARD_GAP * (VISIBLE_CARDS - 1)
const EDGE_FADE_WIDTH = CARD_WIDTH * 0.6

// CS:GO-style rarity colour bar - a plain, scheme-invariant palette (owner
// request, 2026-09-25): common/rare/epic/legendary map onto an
// increasingly saturated tier, exactly the visual grammar a case-opening
// strip needs to read at a glance while it's still scrolling past. 'NONE'
// (the "landed nothing" card) gets the same muted treatment as COMMON -
// deliberately unremarkable, never drawing the eye like a real rarity would.
const RARITY_BAR_COLOR: Record<string, string> = {
  COMMON: '$plasticEdge',
  RARE: '$wiiBlue',
  EPIC: '$standPurple',
  LEGENDARY: '$standGold',
  NONE: '$plasticEdge',
}

type Props = {
  /** The full built strip (case-strip.ts's buildCaseStrip) - every device
   * computes the identical one from the same seed, so this is never built
   * inside the component itself. */
  cards: CaseStripCard[]
  landingIndex: number
  landingOffset: number
  /** true while this slot's strip should be scrolling (the 'spin' phase);
   * false during BOTH 'narrator' (before the spin starts) and 'land'
   * (after it ends) - see `landed` for telling those two apart, and
   * PowerRoulette's identical prop for the bug this distinction exists to
   * not reintroduce (the answer visible at rest before the spin even
   * started). */
  spinning: boolean
  landed: boolean
  reducedMotion: boolean
  spinMs: number
}

// The Stand/Devil Fruit half of the sorteo's "hybrid" ruleta redesign
// (owner decision, 2026-09-25 playtest feedback): a CS:GO-style horizontal
// case strip - real card art + a rarity colour bar, a fixed needle marking
// where it lands, and a "near-miss" catch (buildCaseStrip's own
// landingOffset) instead of always stopping dead-centre. The haki/scalar
// slots keep the vertical PowerRoulette (see its own doc) - they have too
// few possible values for a 50-card strip to make sense.
export function CaseStripReel({
  cards,
  landingIndex,
  landingOffset,
  spinning,
  landed,
  reducedMotion,
  spinMs,
}: Props) {
  const translateX = useSharedValue(0)
  const flash = useSharedValue(0)

  const restX = caseStripRestX(landingIndex, landingOffset, CARD_WIDTH, CARD_GAP, WINDOW_WIDTH)

  useEffect(() => {
    if (reducedMotion) {
      translateX.value = landed ? restX : 0
      flash.value = 0
      return
    }
    if (spinning) {
      translateX.value = 0
      flash.value = 0
      translateX.value = withTiming(restX, { duration: spinMs, easing: Easing.out(Easing.cubic) })
    } else if (landed) {
      translateX.value = restX
      flash.value = withSequence(withTiming(1, { duration: 140 }), withTiming(0, { duration: 460 }))
    } else {
      // Idle, before the spin has even started (the 'narrator' phase) -
      // rest at the strip's own start, NOT restX - see PowerRoulette's
      // identical `landed` doc for the bug this branch exists to not
      // reintroduce.
      translateX.value = 0
      flash.value = 0
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- translateX/flash are stable shared values, not reactive deps
  }, [spinning, landed, reducedMotion, restX, spinMs])

  const rowStyle = useAnimatedStyle(() => ({
    transform: [{ translateX: translateX.value }],
  }))
  const needleFlashStyle = useAnimatedStyle(() => ({
    opacity: 0.35 + flash.value * 0.65,
  }))

  return (
    <YStack
      width={WINDOW_WIDTH}
      height={CARD_HEIGHT}
      overflow="hidden"
      position="relative"
      rounded="$card"
    >
      <Animated.View style={[{ flexDirection: 'row' }, rowStyle]}>
        {cards.map((card, i) => (
          <YStack
            key={`${card.id}-${i}`}
            width={CARD_WIDTH}
            height={CARD_HEIGHT}
            mr={i === cards.length - 1 ? 0 : CARD_GAP}
            rounded="$card"
            overflow="hidden"
            bg="$plasticFill"
          >
            <LazyImage
              uri={card.picture ?? null}
              height={CARD_HEIGHT - RARITY_BAR_HEIGHT}
              contentFit="cover"
              fallback={<Sparkles size={22} color="$standPurple" />}
              // 'high' lane (its own longer watchdog/budget, see
              // image-queue.ts) - the strip is only on screen for a few
              // seconds, so it can't wait behind the app's normal 'grid'
              // catalogue traffic. order ranks cards by distance from the
              // landing card, so the ones the eye actually lands on/near
              // win any contention over the far-off decoys.
              order={Math.abs(i - landingIndex)}
              priorityHint="high"
              recyclingKey={card.id}
            />
            <YStack
              height={RARITY_BAR_HEIGHT}
              width="100%"
              bg={asToken(RARITY_BAR_COLOR[card.rarity] ?? '$plasticEdge')}
            />
            <YStack position="absolute" b={RARITY_BAR_HEIGHT + 2} l={2} r={2}>
              <GlowText level="label" fontSize="$1" numberOfLines={1}>
                {card.label}
              </GlowText>
            </YStack>
          </YStack>
        ))}
      </Animated.View>

      {/* Fixed needle marking the landing point - the strip moves under it,
          never the other way round. Pulses on land, same beat as
          PowerRoulette's highlight flash. */}
      <Animated.View
        style={[
          {
            position: 'absolute',
            left: WINDOW_WIDTH / 2 - 1,
            top: 0,
            bottom: 0,
            width: 2,
            backgroundColor: 'rgba(242,199,68,0.9)',
            pointerEvents: 'none',
          },
          needleFlashStyle,
        ]}
      />

      {/* Left/right fade - the strip enters/leaves instead of being hard
          clipped by the window edge. */}
      <LinearGradient
        position="absolute"
        l={0}
        t={0}
        b={0}
        width={EDGE_FADE_WIDTH}
        colors={['rgba(10,12,20,0.55)', 'rgba(10,12,20,0)']}
        start={[0, 0]}
        end={[1, 0]}
        style={{ pointerEvents: 'none' } as object}
      />
      <LinearGradient
        position="absolute"
        r={0}
        t={0}
        b={0}
        width={EDGE_FADE_WIDTH}
        colors={['rgba(10,12,20,0)', 'rgba(10,12,20,0.55)']}
        start={[0, 0]}
        end={[1, 0]}
        style={{ pointerEvents: 'none' } as object}
      />
    </YStack>
  )
}
