import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Modal, useWindowDimensions } from 'react-native'
import Animated, {
  Easing,
  useAnimatedStyle,
  useSharedValue,
  withRepeat,
  withSequence,
  withTiming,
} from 'react-native-reanimated'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { ScrollView, XStack, YStack } from 'tamagui'

import {
  STAND_STAT_KEYS,
  STAND_STAT_LABELS,
} from '@/features/game/components/presentational/match/loadout-card'
import { MangaVerdictText } from '@/features/game/components/presentational/match/manga/manga-verdict-text'
import { RadialGlow } from '@/features/game/components/presentational/match/manga/radial-glow'
import { PowerBlock } from '@/features/game/components/presentational/match/power-block'
import { revealLayout } from '@/features/game/lib/reveal-layout'
import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'
import type { StandResponse } from '@/features/stands/types/stands.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { cardSource, focalPosition, fullSource, thumbSource } from '@/shared/lib/picture-source'
import { notifyScroll } from '@/shared/lib/scroll-bus'

// The Stand slot's own evolution beats (2026-09-25 "algo esta pasando"
// feature, see reveal.go's evolveMs/loadout-reveal.ts's revealTimeline):
// 'base' plays the root form exactly like a normal (non-evolving) landing;
// 'evolving' is the suspense beat (no card, just the FX overlay below);
// 'step' is an INTERMEDIATE stage flashing in; 'final' is the last stage,
// same flash but with the "¡EVOLUCIÓN!" stamp. undefined (every other
// slot, or a non-evolving Stand) renders exactly as before this feature.
export type EvolvePhase = 'base' | 'evolving' | 'step' | 'final'

type Props = {
  visible: boolean
  kind: 'stand' | 'devilFruit'
  stand?: StandResponse
  devilFruit?: DevilFruitResponse
  participantName: string
  onSkip: () => void
  evolvePhase?: EvolvePhase
  /** The "algo esta pasando" narrator line - only shown during 'evolving'. */
  evolveMessage?: string
  reducedMotion?: boolean
}

// The sorteo's own big power reveal (owner request, 2026-08-30): when a
// participant's turn lands a Stand or a Devil Fruit, this takes over the
// whole screen for that slot's hold (game.RevealHoldStandMs/
// RevealHoldFruitMs, ~10s/5s at Swift speed) so everyone can actually read
// the art, description and skills before the sorteo moves on - the exact
// gap the pre-2026-08-30 reveal never closed (see reveal.go's superseded
// "this UI never renders a power's description" note). Reuses PowerBlock,
// the same big-card layout LoadoutModal already shows after the match, so
// the two never drift. Carries its own Skip button (bug found 2026-08-30
// manual testing: RevealStage's own Skip sits underneath this Modal, but a
// Modal captures every touch on both native and web - the button was
// visually present but completely unreachable while this card was up).
export function PowerRevealCard({
  visible,
  kind,
  stand,
  devilFruit,
  participantName,
  onSkip,
  evolvePhase,
  evolveMessage,
  reducedMotion = false,
}: Props) {
  const { t } = useTranslation()
  const insets = useSafeAreaInsets()
  const { width, height } = useWindowDimensions()
  const layout = revealLayout({ width, height }, insets)

  const isStand = kind === 'stand'
  const power = isStand ? stand : devilFruit
  const statFlexBasis = layout.statColumns === 3 ? '30%' : 72
  const isEvolving = evolvePhase === 'evolving'
  const isEvolutionLanding = evolvePhase === 'step' || evolvePhase === 'final'

  // "Menacing + flash" evolution FX (owner decision, 2026-09-25 - see
  // reveal.go's evolveMs doc): the card shakes and pulses gold with the
  // "algo esta pasando" message during 'evolving', then a full white flash
  // announces each new stage ('step'/'final') as it lands. Transform/
  // opacity only, per the project's motion norm (gameplay-power-fx.md) -
  // reducedMotion collapses straight to the end state, no shake/flash.
  const shakeX = useSharedValue(0)
  const glowOpacity = useSharedValue(0)
  const flashOpacity = useSharedValue(0)
  const cardScale = useSharedValue(1)

  useEffect(() => {
    if (reducedMotion) {
      shakeX.value = 0
      glowOpacity.value = isEvolving ? 0.6 : 0
      flashOpacity.value = 0
      cardScale.value = 1
      return
    }
    if (isEvolving) {
      shakeX.value = withRepeat(
        withSequence(
          withTiming(-4, { duration: 70, easing: Easing.inOut(Easing.quad) }),
          withTiming(4, { duration: 70, easing: Easing.inOut(Easing.quad) })
        ),
        -1,
        true
      )
      glowOpacity.value = withRepeat(
        withSequence(
          withTiming(0.85, { duration: 400, easing: Easing.out(Easing.quad) }),
          withTiming(0.35, { duration: 400, easing: Easing.in(Easing.quad) })
        ),
        -1,
        true
      )
      return
    }
    shakeX.value = withTiming(0, { duration: 150 })
    glowOpacity.value = withTiming(0, { duration: 200 })
    if (isEvolutionLanding) {
      flashOpacity.value = withSequence(
        withTiming(1, { duration: 90, easing: Easing.out(Easing.quad) }),
        withTiming(0, { duration: 260 })
      )
      cardScale.value = 0.9
      cardScale.value = withSequence(
        withTiming(1.05, { duration: 180, easing: Easing.out(Easing.back(1.6)) }),
        withTiming(1, { duration: 120 })
      )
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- shared values are stable refs; only evolvePhase/stand identity/reducedMotion should retrigger the FX
  }, [evolvePhase, stand?.id, devilFruit?.id, reducedMotion, isEvolving, isEvolutionLanding])

  const shakeStyle = useAnimatedStyle(() => ({
    transform: [{ translateX: shakeX.value }, { scale: cardScale.value }],
  }))
  const glowStyle = useAnimatedStyle(() => ({ opacity: glowOpacity.value }))
  const flashStyle = useAnimatedStyle(() => ({ opacity: flashOpacity.value }))

  return (
    <Modal visible={visible} transparent animationType="fade" statusBarTranslucent>
      <YStack
        flex={1}
        items="center"
        justify="center"
        p="$3"
        pt={insets.top + 12}
        pb={insets.bottom + 12}
        bg="rgba(10,12,20,0.72)"
      >
        <Animated.View style={shakeStyle}>
        <GlassPanel
          tone="strong"
          radiusSize="panel"
          elevate={3}
          width="100%"
          maxW={560}
          $md={{ maxW: 640 }}
          $lg={{ maxW: 720 }}
          p="$3"
          $sm={{ p: '$4' }}
          gap="$2.5"
          overflow="hidden"
          position="relative"
        >
          {isEvolving ? (
            <Animated.View
              style={[
                { position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 },
                glowStyle,
              ]}
              pointerEvents="none"
            >
              <RadialGlow
                stops={[
                  { offset: '0%', color: '#F2C744' },
                  { offset: '55%', color: '#B8841C' },
                  { offset: '100%', color: '#1A1108', opacity: 0 },
                ]}
              />
            </Animated.View>
          ) : null}
          <XStack items="center" justify="center">
            <GlowText level="label" tone="soft">
              {participantName}
            </GlowText>
          </XStack>
          {isEvolving ? (
            <YStack items="center" justify="center" gap="$3" py="$4">
              <MangaVerdictText fillColor="#F2C744">
                {evolveMessage ?? t('game.match.reveal.evolution.evolving')}
              </MangaVerdictText>
              {stand ? (
                <GlowText level="label" tone="soft" align="center">
                  {stand.name}
                </GlowText>
              ) : null}
            </YStack>
          ) : (
            <ScrollView
            maxH={layout.scrollMaxHeight}
            onScroll={notifyScroll}
            scrollEventThrottle={16}
          >
            <PowerBlock
              picture={power ? cardSource(power) : undefined}
              fullPicture={power ? fullSource(power) : undefined}
              contentPosition={power ? focalPosition(power) : undefined}
              // The strip already cached this stand/fruit's thumb during the
              // spin (case-strip-reel.tsx) - showing it as the placeholder
              // means art appears the instant the card mounts instead of a
              // blank Skeleton, then sharpens into the cardSource()
              // rendition once it loads. See power-block.tsx's doc.
              placeholderUri={power ? thumbSource(power) : undefined}
              // The single foreground image on screen - must never wait
              // behind the reel's orphaned queue slots (2026-09-25 fix, see
              // image-queue.ts's QueueLane doc).
              priorityHint="hero"
              name={power?.name}
              rarityLabel={
                isStand
                  ? stand
                    ? t(`enums.rarity.${stand.rarity}`)
                    : undefined
                  : devilFruit
                    ? t(`enums.fruitType.${devilFruit.fruitType}`)
                    : undefined
              }
              description={power?.description}
              skills={power?.skills}
              fallbackLabel={isStand ? t('game.match.noStand') : t('game.match.noFruit')}
              artHeight={layout.artHeight}
            >
              {isStand && stand ? (
                <XStack flexWrap="wrap" gap="$1.5" mt="$1" $md={{ gap: '$2' }}>
                  {STAND_STAT_KEYS.map((key) => (
                    <YStack
                      key={key}
                      flexBasis={statFlexBasis}
                      grow={1}
                      minW={64}
                      $sm={{ flexBasis: 88, minW: 88 }}
                      items="center"
                      gap="$0.5"
                    >
                      <GlowText level="label" tone="soft" fontSize="$3">
                        {STAND_STAT_LABELS[key]}
                      </GlowText>
                      <GlowText level="heading" fontSize="$5" $md={{ fontSize: '$7' }}>
                        {t(`enums.standStat.${stand[key]}`)}
                      </GlowText>
                    </YStack>
                  ))}
                </XStack>
              ) : null}
            </PowerBlock>
          </ScrollView>
          )}
          {isEvolutionLanding && evolvePhase === 'final' ? (
            <XStack justify="center">
              <GlassPanel tone="plastic" px="$3" py="$1" rounded="$pill" elevate={0}>
                <GlowText level="label" fontSize="$4" color="$standGold">
                  {t('game.match.reveal.evolution.final')}
                </GlowText>
              </GlassPanel>
            </XStack>
          ) : null}
          <XStack justify="center">
            <GlossButton
              tone="glass"
              btnSize="sm"
              onPress={onSkip}
              accessibilityLabel={t('game.match.reveal.skipA11y')}
              tooltip={t('game.match.reveal.skipA11y')}
            >
              {t('game.match.reveal.skip')}
            </GlossButton>
          </XStack>
          {isEvolutionLanding ? (
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
                flashStyle,
              ]}
              pointerEvents="none"
            />
          ) : null}
        </GlassPanel>
        </Animated.View>
      </YStack>
    </Modal>
  )
}
