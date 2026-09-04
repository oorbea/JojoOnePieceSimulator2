import { Map } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Image, Pressable } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { ImageLightbox } from '@/shared/components/presentational/image-lightbox'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import type { StageResponse } from '@/features/stages/types/stages.types'

type Props = {
  stage: StageResponse
}

export function StageDetail({ stage }: Props) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  return (
    <YStack gap="$4">
      <Pressable
        onPress={() => setIsPreviewOpen(true)}
        disabled={!stage.picture}
        {...a11yProps(t('stages.previewA11y', { name: stage.name }), 'imagebutton')}
      >
        <YStack width="100%" height={220} rounded="$card" overflow="hidden" position="relative" bg="$plasticEdge">
          <InsetRing rounded="$card" />
          {/* `|| null`, not `??` - see admin-panel-crud-ux-fixes.md, empty
              string is falsy but not nullish and shouldn't sneak through. */}
          {stage.picture || null ? (
            <Image source={{ uri: stage.picture }} style={{ width: '100%', height: '100%' }} resizeMode="contain" />
          ) : (
            <YStack flex={1} items="center" justify="center">
              <Map size={40} color="$wiiBlue" />
            </YStack>
          )}
        </YStack>
      </Pressable>
      <ImageLightbox visible={isPreviewOpen} uri={stage.picture || null} onClose={() => setIsPreviewOpen(false)} />

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
