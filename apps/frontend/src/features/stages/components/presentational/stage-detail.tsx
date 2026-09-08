import { Map } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Pressable } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { ImageLightbox } from '@/shared/components/presentational/image-lightbox'
import { LazyImage, type LazyImageState } from '@/shared/components/presentational/lazy-image'
import { a11yProps } from '@/shared/lib/a11y'
import { fullSource, lqipSource } from '@/shared/lib/picture-source'
import type { StageResponse } from '@/features/stages/types/stages.types'

type Props = {
  stage: StageResponse
}

export function StageDetail({ stage }: Props) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [imageState, setImageState] = useState<LazyImageState>('queued')
  const [retryToken, setRetryToken] = useState(0)
  const uri = fullSource(stage)
  const isImageError = imageState === 'error'
  return (
    <YStack gap="$4">
      <Pressable
        onPress={() => (isImageError ? setRetryToken((n) => n + 1) : setIsPreviewOpen(true))}
        disabled={!uri}
        {...a11yProps(
          t(isImageError ? 'common.imageRetry' : 'stages.previewA11y', { name: stage.name }),
          'imagebutton'
        )}
      >
        <LazyImage
          uri={uri}
          lqip={lqipSource(stage)}
          height={220}
          contentFit="contain"
          pictureStatus={stage.pictureStatus}
          retryToken={retryToken}
          onStateChange={setImageState}
          fallback={<Map size={40} color="$wiiBlue" />}
        />
      </Pressable>
      <ImageLightbox visible={isPreviewOpen} uri={uri} lqip={lqipSource(stage)} onClose={() => setIsPreviewOpen(false)} />

      <XStack gap="$2" flexWrap="wrap">
        <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
          <GlowText level="label">{t(`enums.manga.${stage.manga}`)}</GlowText>
        </GlassPanel>
        <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
          <GlowText level="label">{t('stages.orderBadge', { order: stage.order })}</GlowText>
        </GlassPanel>
      </XStack>

      {stage.description ? (
        <YStack gap="$1">
          <GlowText level="label" tone="soft">
            {t('stages.description')}
          </GlowText>
          <GlowText level="label">{stage.description}</GlowText>
        </YStack>
      ) : null}
    </YStack>
  )
}
