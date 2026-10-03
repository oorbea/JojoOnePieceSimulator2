import { Map } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlowText } from '@/shared/components/presentational/glow-text'
import { LazyImage } from '@/shared/components/presentational/lazy-image'
import { WiiCard } from '@/shared/components/presentational/wii-card'
import { focalPosition } from '@/shared/lib/picture-source'
import type { GameStage } from '@/features/game/types/game.types'

type Props = {
  stage: GameStage
  /** Slim one-row version, pinned above the player-by-player reveal. The
   * default is the big card used when the stage is the main news. */
  compact?: boolean
}

// Tells players where a Versus round's fight will take place - the stage is
// only the arena, it grants nothing to either team - so it can be announced
// during the sorteo, before anyone votes.
export function StageAnnouncement({ stage, compact = false }: Props) {
  const { t } = useTranslation()

  const text = (
    <YStack flex={1} gap="$0.5" minW={0}>
      <XStack items="center" gap="$1.5">
        <Map size={14} color="$wiiBlue" />
        <GlowText level="label" tone="soft">
          {t('game.match.stage.venueKicker')}
        </GlowText>
      </XStack>
      <GlowText level={compact ? 'heading' : 'title'}>{stage.name}</GlowText>
      <GlowText level="label" tone="soft">
        {t('game.match.stage.venueHint')}
      </GlowText>
    </YStack>
  )

  if (compact) {
    return (
      <WiiCard padded width="100%">
        <XStack items="center" gap="$3">
          <YStack width={80}>
            <LazyImage
              uri={stage.pictureThumb || stage.picture || null}
              aspectRatio={16 / 9}
              contentPosition={focalPosition(stage)}
              pictureStatus={stage.pictureStatus}
              fallback={<Map size={24} color="$wiiBlue" />}
            />
          </YStack>
          {text}
        </XStack>
      </WiiCard>
    )
  }

  return (
    <WiiCard padded width="100%" gap="$3">
      <YStack width="100%" maxW={520} self="center">
        <LazyImage
          uri={stage.picture || null}
          aspectRatio={16 / 9}
          contentPosition={focalPosition(stage)}
          pictureStatus={stage.pictureStatus}
          priorityHint="high"
          fallback={<Map size={40} color="$wiiBlue" />}
        />
      </YStack>
      {text}
    </WiiCard>
  )
}
