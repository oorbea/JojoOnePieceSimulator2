import { Sparkles, UserRound } from '@tamagui/lucide-icons-2'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Image } from 'react-native'
import Animated, { useAnimatedStyle, useSharedValue, withTiming } from 'react-native-reanimated'
import { XStack, YStack } from 'tamagui'

import { TraitChips } from '@/features/game/components/presentational/match/trait-chips'
import type { GameParticipant } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { InsetRing, WiiCard } from '@/shared/components/presentational/wii-card'

type Props = {
  participant: GameParticipant
  isSelf: boolean
  revealed: boolean
  reducedMotion: boolean
}

const STAND_STAT_KEYS = ['attackPower', 'speed', 'attackRange', 'endurance', 'precision', 'potential'] as const
const STAND_STAT_LABELS: Record<(typeof STAND_STAT_KEYS)[number], string> = {
  attackPower: 'PWR',
  speed: 'SPD',
  attackRange: 'RNG',
  endurance: 'END',
  precision: 'PRE',
  potential: 'DEV',
}

// Two-sided flip card: back face is a face-down "drawing" placeholder, front
// face is the actual stand/devilFruit/trait breakdown. Y-axis rotateY,
// backfaceVisibility hidden on both faces, perspective on the parent so the
// flip reads as a physical card rather than a squash. Reduced motion (or
// `!revealed`, e.g. resync) jumps straight to the end state with duration 0.
export function LoadoutCard({ participant, isSelf, revealed, reducedMotion }: Props) {
  const { t } = useTranslation()
  const progress = useSharedValue(revealed ? 1 : 0)

  useEffect(() => {
    const duration = reducedMotion ? 0 : 450
    progress.value = withTiming(revealed ? 1 : 0, { duration })
    // eslint-disable-next-line react-hooks/exhaustive-deps -- progress is a stable shared value, not a reactive dep
  }, [revealed, reducedMotion])

  const backStyle = useAnimatedStyle(() => ({
    transform: [{ perspective: 1200 }, { rotateY: `${progress.value * 180}deg` }],
    opacity: progress.value < 0.5 ? 1 : 0,
  }))
  const frontStyle = useAnimatedStyle(() => ({
    transform: [{ perspective: 1200 }, { rotateY: `${180 + progress.value * 180}deg` }],
    opacity: progress.value >= 0.5 ? 1 : 0,
  }))

  const loadout = participant.loadout

  return (
    <WiiCard
      padded
      width={220}
      gap="$2.5"
      borderColor={isSelf ? '$wiiBlue' : undefined}
      borderWidth={isSelf ? 2.5 : undefined}
    >
      <XStack items="center" gap="$1.5" flexWrap="wrap">
        <GlowText level="label" numberOfLines={1} flex={1}>
          {participant.displayName}
        </GlowText>
        {isSelf ? (
          <GlassPanel tone="plastic" px="$2" py="$0.5" rounded="$pill" elevate={0}>
            <GlowText level="label">{t('game.lobby.you')}</GlowText>
          </GlassPanel>
        ) : null}
      </XStack>

      <YStack position="relative" style={{ perspective: 1200 } as never}>
        <Animated.View style={[{ width: '100%' }, backStyle]}>
          <YStack
            width="100%"
            height={200}
            rounded="$card"
            overflow="hidden"
            items="center"
            justify="center"
            gap="$2"
            bg="$plasticEdge"
            style={{ backfaceVisibility: 'hidden' } as never}
          >
            <InsetRing rounded="$card" />
            <UserRound size={32} color="$panelTextSoft" />
            <GlowText level="label" tone="soft">
              {t('game.match.drawing')}
            </GlowText>
          </YStack>
        </Animated.View>

        <Animated.View
          style={[{ width: '100%', position: 'absolute', top: 0, left: 0 }, frontStyle]}
          pointerEvents={revealed ? 'auto' : 'none'}
        >
          <YStack width="100%" gap="$2.5" style={{ backfaceVisibility: 'hidden' } as never}>
            <YStack width="100%" height={110} rounded="$card" overflow="hidden" position="relative" bg="$plasticEdge">
              <InsetRing rounded="$card" />
              {loadout?.stand ? (
                loadout.stand.pictureThumb || null ? (
                  <Image source={{ uri: loadout.stand.pictureThumb }} style={{ width: '100%', height: '100%' }} />
                ) : (
                  <YStack flex={1} items="center" justify="center">
                    <Sparkles size={26} color="$standPurple" />
                  </YStack>
                )
              ) : (
                <YStack flex={1} items="center" justify="center">
                  <GlowText level="label" tone="soft">
                    {t('game.match.noStand')}
                  </GlowText>
                </YStack>
              )}
            </YStack>

            {loadout?.stand ? (
              (() => {
                const stand = loadout.stand
                return (
                  <YStack gap="$1.5">
                    <XStack items="center" justify="space-between">
                      <GlowText level="label" numberOfLines={1} flex={1}>
                        {stand.name}
                      </GlowText>
                      <GlassPanel tone="plastic" px="$2" py="$0.5" rounded="$pill" elevate={0}>
                        <GlowText level="label" fontSize="$1">
                          {t(`enums.rarity.${stand.rarity}`)}
                        </GlowText>
                      </GlassPanel>
                    </XStack>
                    <XStack flexWrap="wrap" gap="$1.5">
                      {STAND_STAT_KEYS.map((key) => (
                        <YStack key={key} flexBasis={56} grow={1} minW={56} items="center" gap="$0.5">
                          <GlowText level="label" tone="soft" fontSize="$1">
                            {STAND_STAT_LABELS[key]}
                          </GlowText>
                          <GlowText level="label" fontSize="$3">
                            {t(`enums.standStat.${stand[key]}`)}
                          </GlowText>
                        </YStack>
                      ))}
                    </XStack>
                  </YStack>
                )
              })()
            ) : null}

            <YStack width="100%" height={90} rounded="$card" overflow="hidden" position="relative" bg="$plasticEdge">
              <InsetRing rounded="$card" />
              {loadout?.devilFruit ? (
                loadout.devilFruit.pictureThumb || null ? (
                  <Image source={{ uri: loadout.devilFruit.pictureThumb }} style={{ width: '100%', height: '100%' }} />
                ) : (
                  <YStack flex={1} items="center" justify="center">
                    <Sparkles size={22} color="$tangerine" />
                  </YStack>
                )
              ) : (
                <YStack flex={1} items="center" justify="center">
                  <GlowText level="label" tone="soft">
                    {t('game.match.noFruit')}
                  </GlowText>
                </YStack>
              )}
            </YStack>

            {loadout?.devilFruit ? (
              <XStack items="center" justify="space-between" gap="$1.5">
                <GlowText level="label" numberOfLines={1} flex={1}>
                  {loadout.devilFruit.name}
                </GlowText>
                <GlassPanel tone="plastic" px="$2" py="$0.5" rounded="$pill" elevate={0}>
                  <GlowText level="label" fontSize="$1">
                    {t(`enums.fruitType.${loadout.devilFruit.fruitType}`)}
                  </GlowText>
                </GlassPanel>
              </XStack>
            ) : null}

            {loadout ? <TraitChips loadout={loadout} /> : null}
          </YStack>
        </Animated.View>
      </YStack>
    </WiiCard>
  )
}
