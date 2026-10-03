import { useEffect } from 'react'
import Animated, {
  Easing,
  useAnimatedStyle,
  useSharedValue,
  withSequence,
  withTiming,
} from 'react-native-reanimated'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  /** What is rising, e.g. "Spin" or "Haki de Armadura". */
  statLabel: string
  /** The value as it was drawn. */
  fromLabel: string
  /** The value the power effect raised it to. */
  toLabel: string
  /** False while the drawn value is still showing (the effect's intro), true
   * once it has risen (the effect's closing hold). */
  raised: boolean
  /** The "it went up" stamp's text, shown once raised. */
  stamp: string
  reducedMotion?: boolean
}

// The quiet half of a power effect's beat in the sorteo (the loud half is the
// stand/fruit evolution, which reuses PowerRevealCard): the card the effect
// changes comes back with its drawn value, then the value steps up in gold
// with a short pulse. Transform/opacity only, per the project's motion norm
// (gameplay-power-fx.md) - reducedMotion collapses straight to the end state.
// Not a button, so nothing here needs a tooltip.
export function EffectLevelUp({
  statLabel,
  fromLabel,
  toLabel,
  raised,
  stamp,
  reducedMotion = false,
}: Props) {
  const scale = useSharedValue(1)

  useEffect(() => {
    if (!raised || reducedMotion) {
      scale.value = 1
      return
    }
    scale.value = 0.9
    scale.value = withSequence(
      withTiming(1.08, { duration: 180, easing: Easing.out(Easing.back(1.6)) }),
      withTiming(1, { duration: 140 })
    )
    // eslint-disable-next-line react-hooks/exhaustive-deps -- scale is a stable shared value; only raised/reducedMotion should retrigger the pulse
  }, [raised, reducedMotion])

  const pulseStyle = useAnimatedStyle(() => ({ transform: [{ scale: scale.value }] }))

  return (
    <YStack items="center" gap="$2" width="100%">
      <GlowText level="label" tone="soft">
        {statLabel}
      </GlowText>
      <Animated.View style={pulseStyle}>
        <GlassPanel tone="plastic" px="$4" py="$2.5" rounded="$card" elevate={raised ? 2 : 0}>
          <XStack items="center" justify="center" gap="$2.5" flexWrap="wrap">
            <GlowText
              level="heading"
              tone={raised ? 'soft' : undefined}
              opacity={raised ? 0.55 : 1}
            >
              {fromLabel}
            </GlowText>
            {raised ? (
              <>
                <GlowText level="heading" tone="soft">
                  {'›'}
                </GlowText>
                <GlowText level="heading" color="$standGold">
                  {toLabel}
                </GlowText>
              </>
            ) : null}
          </XStack>
        </GlassPanel>
      </Animated.View>
      {raised ? (
        <GlassPanel tone="plastic" px="$3" py="$1" rounded="$pill" elevate={0}>
          <GlowText level="label" fontSize="$4" color="$standGold">
            {stamp}
          </GlowText>
        </GlassPanel>
      ) : null}
    </YStack>
  )
}
