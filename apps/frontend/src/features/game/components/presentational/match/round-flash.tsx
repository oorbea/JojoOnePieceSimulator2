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
import { DecayGrid } from '@/features/game/components/presentational/match/manga/decay-grid'
import { MangaVerdictText } from '@/features/game/components/presentational/match/manga/manga-verdict-text'
import { RadialGlow } from '@/features/game/components/presentational/match/manga/radial-glow'
import { SpeedLines } from '@/features/game/components/presentational/match/manga/speed-lines'
import { roundLoseSound, roundWinSound } from '@/shared/assets'
import { useSound } from '@/shared/hooks/use-sound'

// Manga JoJo × One Piece palette (owner decision, 2026-09-25 playtest
// feedback - see ObsidianVault/game-victory-defeat-cinematic-2026-09-14.md).
// Win: a golden radial burst over morado, with rotating speed-lines - the
// one beat that keeps rays-from-centre, since it's a triumphant impact.
// Lose: rojo tinta → negro with an "aura de decadencia" (a grid that only
// holds together near the centre and dissolves into black toward the
// edges) instead of rays - the owner was explicit a burst reads wrong for
// a loss.
const WIN_GLOW = [
  { offset: '0%', color: '#F2C744' },
  { offset: '45%', color: '#D99A1C' },
  { offset: '100%', color: '#7A2E86' },
]
const LOSE_GLOW = [
  { offset: '0%', color: '#6B1420' },
  { offset: '70%', color: '#2C0A0D' },
  { offset: '100%', color: '#120507' },
]

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
        position="relative"
        overflow="hidden"
        bg={outcome === 'win' ? '#7A2E86' : '#120507'}
      >
        <RadialGlow stops={outcome === 'win' ? WIN_GLOW : LOSE_GLOW} />
        <Animated.View
          style={[{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 }, style]}
          pointerEvents="none"
        >
          {outcome === 'win' ? <SpeedLines /> : <DecayGrid />}
        </Animated.View>
        <Animated.View style={style}>
          <MangaVerdictText fillColor={outcome === 'win' ? '$standGold' : '$strawHatRed'}>
            {outcome === 'win'
              ? t('game.match.cinematic.roundWin')
              : t('game.match.cinematic.roundLose')}
          </MangaVerdictText>
        </Animated.View>
      </YStack>
    </Modal>
  )
}
