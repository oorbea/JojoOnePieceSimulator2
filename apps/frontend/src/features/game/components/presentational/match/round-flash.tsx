import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Modal } from 'react-native'
import Animated, {
  Easing,
  useAnimatedStyle,
  useSharedValue,
  withSequence,
  withTiming,
} from 'react-native-reanimated'
import { YStack } from 'tamagui'

import { cinematicTimeline } from '@/features/game/lib/outcome-cinematic'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { roundLoseSound, roundWinSound } from '@/shared/assets'
import { useSound } from '@/shared/hooks/use-sound'

type Props = {
  visible: boolean
  outcome: 'win' | 'lose'
  reducedMotion: boolean
  /** Called once, exactly timeline.totalMs after mounting (1.5s) - never
   * skippable, it's too short to need a skip control. */
  onDone: () => void
}

// The 1.5s full-screen flash at the start of RESOLVING (see
// ObsidianVault/game-victory-defeat-cinematic-2026-09-14.md): a small
// anticipatory beat ahead of RoundResultPanel's own 6s window, driven by
// game-result.ts's roundOutcome. Reduced motion skips it outright - it is
// pure reinforcement of information RoundResultPanel already states, unlike
// the end-of-game cinematic there is no unique content to preserve as a
// static frame.
export function RoundFlash({ visible, outcome, reducedMotion, onDone }: Props) {
  const { t } = useTranslation()
  const timeline = cinematicTimeline(outcome === 'win' ? 'roundWin' : 'roundLose')
  const opacity = useSharedValue(0)
  const scale = useSharedValue(0.85)
  const sound = useSound(outcome === 'win' ? roundWinSound : roundLoseSound, !reducedMotion)

  useEffect(() => {
    if (!visible || reducedMotion) {
      if (visible) onDone()
      return
    }

    sound.play()
    opacity.value = withSequence(
      withTiming(1, { duration: 220, easing: Easing.out(Easing.quad) }),
      withTiming(1, { duration: timeline.totalMs - 220 - 350 }),
      withTiming(0, { duration: 350 })
    )
    scale.value = withTiming(1, { duration: 260, easing: Easing.out(Easing.back(1.4)) })

    const doneTimer = setTimeout(onDone, timeline.totalMs)
    return () => {
      clearTimeout(doneTimer)
      sound.stop()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- shared values are stable refs; only visible/reducedMotion/outcome should restart the flash
  }, [visible, reducedMotion, outcome])

  const style = useAnimatedStyle(() => ({
    opacity: opacity.value,
    transform: [{ scale: scale.value }],
  }))

  if (reducedMotion) return null

  return (
    <Modal visible={visible} transparent animationType="none" statusBarTranslucent>
      <YStack
        flex={1}
        items="center"
        justify="center"
        bg={outcome === 'win' ? 'rgba(242,199,68,0.9)' : 'rgba(58,10,16,0.9)'}
      >
        <Animated.View style={style}>
          <GlowText level="hero" align="center" tone="onColor">
            {outcome === 'win'
              ? t('game.match.cinematic.roundWin')
              : t('game.match.cinematic.roundLose')}
          </GlowText>
        </Animated.View>
      </YStack>
    </Modal>
  )
}
