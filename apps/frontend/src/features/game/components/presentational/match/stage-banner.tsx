import { Map } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { LazyImage } from '@/shared/components/presentational/lazy-image'
import { WiiCard } from '@/shared/components/presentational/wii-card'
import type { GameStage } from '@/features/game/types/game.types'

type Props = {
  stage: GameStage
  roundIndex: number
}

// Hero card for the round's Stage - same art-well/badge recipe as
// StageCard/StandCard (the `|| null` guard, not `??`, matters here too:
// picture/pictureThumb can be an empty string mid-generation).
export function StageBanner({ stage, roundIndex }: Props) {
  const { t } = useTranslation()

  return (
    <WiiCard padded width="100%" gap="$3">
      <LazyImage
        uri={stage.picture || null}
        aspectRatio={16 / 9}
        pictureStatus={stage.pictureStatus}
        priorityHint="high"
        fallback={<Map size={40} color="$wiiBlue" />}
      />

      <YStack gap="$1.5">
        <XStack gap="$2" flexWrap="wrap">
          <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
            <GlowText level="label">{t(`enums.manga.${stage.manga}`)}</GlowText>
          </GlassPanel>
          <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
            <GlowText level="label">{t('game.match.roundBadge', { index: roundIndex + 1 })}</GlowText>
          </GlassPanel>
        </XStack>
        <GlowText level="title">{stage.name}</GlowText>
        <GlowText level="label" tone="soft">
          {stage.description}
        </GlowText>
      </YStack>
    </WiiCard>
  )
}
