import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Modal } from 'react-native'
import Animated, {
  Easing,
  useAnimatedStyle,
  useSharedValue,
  withDelay,
  withSequence,
  withTiming,
} from 'react-native-reanimated'
import { YStack } from 'tamagui'

import { cinematicTimeline } from '@/features/game/lib/outcome-cinematic'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { defeatCinematicSound, nameTickSound, victoryCinematicSound } from '@/shared/assets'
import { useSound } from '@/shared/hooks/use-sound'

type Props = {
  kind: 'victory' | 'defeat'
  /** Already-translated headline - "DERROTA" / the winning team's name or
   * "LA ESCUADRA SOBREVIVIÓ" (owner's medium-personalization call: names and
   * team/verdict, no power or stage art). */
  title: string
  /** VERSUS: the viewer's own outcome subtitle. GAUNTLET: null - the run is
   * collective, there is no individual line to add. */
  subtitle: string | null
  /** Winning seats' display names for the victory "memory" beat - empty for
   * defeat, which has no name list. */
  names: string[]
  reducedMotion: boolean
  /** Called once, either when the timeline finishes on its own or the
   * viewer skips - never called twice. */
  onDone: () => void
}

// The end-of-game cinematic (see ObsidianVault/game-victory-defeat-
// cinematic-2026-09-14.md): a full-screen Modal that plays ahead of
// MatchResultScreen. Each viewer skips their own copy independently - this
// component has no notion of other players, the container decides when to
// mount/unmount it per viewer.
//
// Deliberately has no `visible` prop - the CALLER only mounts this while a
// play-through should be showing (see lobby-room-screen.tsx), so every
// mount IS a fresh play-through and every bit of state below can just
// initialize fresh, with no reset-on-reopen logic needed.
//
// All motion is transform/opacity only (project norm): every beat below is
// a set of full-bleed layers whose OPACITY the phase timeline drives, never
// a background-color interpolation - see power-roulette.tsx for the same
// discipline. Reduced motion collapses straight to the final static frame
// (title + subtitle, full opacity, no audio) with skip available
// immediately, per useReducedMotion's project-wide contract.
export function OutcomeCinematic({ kind, title, subtitle, names, reducedMotion, onDone }: Props) {
  const { t } = useTranslation()
  const timeline = cinematicTimeline(kind)
  // Reduced motion always allows an immediate skip - no timer needed for
  // that case, so `canSkip` is derived rather than stored for it. For the
  // full playback, `skipTimerFired` is only ever set from inside the
  // setTimeout callback below (never synchronously in the effect body,
  // which react-hooks/set-state-in-effect rejects).
  const [skipTimerFired, setSkipTimerFired] = useState(false)
  const canSkip = reducedMotion || skipTimerFired
  // How many of `names` have been revealed so far in the victory "memory"
  // beat - driven by a cascade of setTimeouts (one per name, see the effect
  // below), each also firing `tick`. A plain counter rather than per-item
  // Reanimated values: the number of names varies per match, and Reanimated
  // shared values must be created in a fixed, hook-order-stable set.
  const [revealedCount, setRevealedCount] = useState(0)

  const impact = useSharedValue(0)
  const fracture = useSharedValue(0)
  const ink = useSharedValue(0)
  const verdict = useSharedValue(0)
  const verdictScale = useSharedValue(0.7)
  const silence = useSharedValue(0)
  const seal = useSharedValue(0)
  const epilogue = useSharedValue(0)

  const track = useSound(
    kind === 'defeat' ? defeatCinematicSound : victoryCinematicSound,
    !reducedMotion
  )
  const tick = useSound(nameTickSound, !reducedMotion && kind === 'victory')

  useEffect(() => {
    if (reducedMotion) {
      impact.value = 0
      fracture.value = kind === 'defeat' ? 1 : 0
      ink.value = kind === 'defeat' ? 1 : 0
      verdict.value = 1
      verdictScale.value = 1
      silence.value = 0
      seal.value = kind === 'victory' ? 1 : 0
      epilogue.value = 0
      return
    }

    track.play()
    const phase = (name: string) => timeline.phases.find((p) => p.name === name)!

    if (kind === 'defeat') {
      impact.value = withSequence(withTiming(1, { duration: 80 }), withTiming(0, { duration: 320 }))
      fracture.value = withDelay(phase('fracture').startMs, withTiming(1, { duration: 400 }))
      ink.value = withDelay(phase('ink').startMs, withTiming(1, { duration: 500 }))
      verdictScale.value = withDelay(
        phase('verdict').startMs,
        withSequence(
          withTiming(1.6, { duration: 0 }),
          withTiming(1, { duration: 450, easing: Easing.out(Easing.back(1.5)) })
        )
      )
      verdict.value = withDelay(phase('verdict').startMs, withTiming(1, { duration: 300 }))
      silence.value = withDelay(phase('silence').startMs, withTiming(1, { duration: 600 }))
    } else {
      impact.value = withDelay(phase('breath').startMs, withTiming(1, { duration: 500 }))
      fracture.value = withDelay(phase('dawn').startMs, withTiming(1, { duration: 900 }))
      verdictScale.value = withDelay(
        phase('coronation').startMs,
        withSequence(withTiming(0.92, { duration: 0 }), withTiming(1, { duration: 500 }))
      )
      verdict.value = withDelay(phase('coronation').startMs, withTiming(1, { duration: 400 }))
      seal.value = withDelay(phase('seal').startMs, withTiming(1, { duration: 500 }))
    }
    epilogue.value = withDelay(timeline.totalMs - 400, withTiming(1, { duration: 400 }))

    const skipTimer = setTimeout(() => setSkipTimerFired(true), timeline.skipAfterMs)
    const doneTimer = setTimeout(onDone, timeline.totalMs)

    // The "memory" beat: one name at a time, each with its own tick,
    // spread across the beat's own window - capped at 400ms apart so a big
    // squad's names don't blow past the beat into "seal".
    const nameTimers: ReturnType<typeof setTimeout>[] = []
    if (kind === 'victory' && names.length > 0) {
      const memory = phase('memory')
      const step = Math.min(400, memory.durationMs / names.length)
      names.forEach((_, index) => {
        nameTimers.push(
          setTimeout(
            () => {
              setRevealedCount(index + 1)
              tick.play()
            },
            memory.startMs + index * step
          )
        )
      })
    }

    return () => {
      clearTimeout(skipTimer)
      clearTimeout(doneTimer)
      nameTimers.forEach(clearTimeout)
      track.stop()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- shared values/sound controls are stable refs; `names` is deliberately captured at mount, not a dep - the caller may recompute a new array reference on an unrelated re-render (e.g. the follow-up STATE frame after GAME_FINISHED), which must never restart an in-progress cinematic. Only kind/reducedMotion do (component has no `visible` prop - see its own doc comment - so this effect otherwise only ever runs once per mount anyway).
  }, [kind, reducedMotion])

  const impactStyle = useAnimatedStyle(() => ({ opacity: impact.value }))
  const fractureStyle = useAnimatedStyle(() => ({ opacity: fracture.value }))
  const inkStyle = useAnimatedStyle(() => ({ opacity: ink.value }))
  const verdictStyle = useAnimatedStyle(() => ({
    opacity: verdict.value,
    transform: [{ scale: verdictScale.value }],
  }))
  const silenceStyle = useAnimatedStyle(() => ({ opacity: silence.value * 0.5 }))
  const sealStyle = useAnimatedStyle(() => ({ opacity: seal.value * 0.3 }))
  const epilogueStyle = useAnimatedStyle(() => ({ opacity: epilogue.value }))

  const handleSkip = () => {
    if (!canSkip) return
    onDone()
  }

  const skipLabel = t('game.result.cinematic.skipA11y')

  return (
    <Modal
      visible
      transparent={false}
      animationType="fade"
      onRequestClose={handleSkip}
      statusBarTranslucent
    >
      <YStack
        flex={1}
        items="center"
        justify="center"
        bg="$inkBlack"
        position="relative"
        overflow="hidden"
      >
        {/* Impact flash / breath of light - a bright pulse at t=0. */}
        <Animated.View
          style={[
            {
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              backgroundColor: '#FFFFFF',
            },
            impactStyle,
          ]}
          pointerEvents="none"
        />

        {/* Fracture (defeat: desaturated crack tone) / dawn (victory: gold
            wash rising behind the verdict). */}
        <Animated.View
          style={[
            {
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              backgroundColor: kind === 'defeat' ? '#2A2A2E' : '#7A5A0F',
            },
            fractureStyle,
          ]}
          pointerEvents="none"
        />

        {kind === 'defeat' ? (
          <Animated.View
            style={[
              {
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                backgroundColor: '#3A0A10',
              },
              inkStyle,
            ]}
            pointerEvents="none"
          />
        ) : null}

        {/* Silence pulse (defeat only) - a slow low-opacity darkening after
            the verdict lands. */}
        {kind === 'defeat' ? (
          <Animated.View
            style={[
              {
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                backgroundColor: '#000000',
              },
              silenceStyle,
            ]}
            pointerEvents="none"
          />
        ) : null}

        <Animated.View style={verdictStyle}>
          <YStack items="center" gap="$3" px="$5">
            <GlowText
              level="hero"
              align="center"
              color={kind === 'defeat' ? '$strawHatRed' : '$standGold'}
            >
              {title}
            </GlowText>
            {subtitle ? (
              <GlowText level="heading" align="center" tone="onColor">
                {subtitle}
              </GlowText>
            ) : null}
          </YStack>
        </Animated.View>

        {kind === 'victory' && names.length > 0 ? (
          <YStack items="center" gap="$1.5" mt="$4">
            {names.map((name, index) => (
              <GlowText
                key={name}
                level="label"
                tone="onColor"
                align="center"
                opacity={reducedMotion || index < revealedCount ? 1 : 0}
              >
                {name}
              </GlowText>
            ))}
          </YStack>
        ) : null}

        {kind === 'victory' ? (
          <Animated.View
            style={[
              {
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                backgroundColor: '#F2C744',
              },
              sealStyle,
            ]}
            pointerEvents="none"
          />
        ) : null}

        {/* Cross-fade to the result screen underneath. */}
        <Animated.View
          style={[
            {
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              backgroundColor: '#000000',
            },
            epilogueStyle,
          ]}
          pointerEvents="none"
        />

        <YStack position="absolute" b="$5" r="$5">
          <GlossButton
            tone="glass"
            btnSize="sm"
            disabled={!canSkip}
            onPress={handleSkip}
            accessibilityLabel={skipLabel}
            tooltip={skipLabel}
          >
            {t('game.result.cinematic.skip')}
          </GlossButton>
        </YStack>
      </YStack>
    </Modal>
  )
}
