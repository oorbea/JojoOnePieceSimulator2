import { Apple } from '@tamagui/lucide-icons-2'
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
import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'

type Props = {
  devilFruit: DevilFruitResponse
}

// Same shape as StandDetail, minus the stat grid (Devil Fruits have none),
// plus the fruitType badge.
export function DevilFruitDetail({ devilFruit }: Props) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [imageState, setImageState] = useState<LazyImageState>('queued')
  const [retryToken, setRetryToken] = useState(0)
  const uri = fullSource(devilFruit)
  const isImageError = imageState === 'error'
  return (
    <YStack gap="$4">
      <Pressable
        onPress={() => (isImageError ? setRetryToken((n) => n + 1) : setIsPreviewOpen(true))}
        disabled={!uri}
        {...a11yProps(
          t(isImageError ? 'common.imageRetry' : 'devilFruits.previewA11y', { name: devilFruit.name }),
          'imagebutton'
        )}
      >
        <LazyImage
          uri={uri}
          lqip={lqipSource(devilFruit)}
          height={220}
          contentFit="contain"
          pictureStatus={devilFruit.pictureStatus}
          retryToken={retryToken}
          onStateChange={setImageState}
          fallback={<Apple size={40} color="$strawHatRed" />}
        />
      </Pressable>
      <ImageLightbox
        visible={isPreviewOpen}
        uri={uri}
        lqip={lqipSource(devilFruit)}
        onClose={() => setIsPreviewOpen(false)}
      />

      <XStack gap="$2" flexWrap="wrap">
        <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
          <GlowText level="label">{t(`enums.rarity.${devilFruit.rarity}`)}</GlowText>
        </GlassPanel>
        <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
          <GlowText level="label">{t(`enums.fruitType.${devilFruit.fruitType}`)}</GlowText>
        </GlassPanel>
      </XStack>

      {devilFruit.description ? (
        <YStack gap="$1">
          <GlowText level="label" tone="soft">
            {t('devilFruits.description')}
          </GlowText>
          <GlowText level="label">{devilFruit.description}</GlowText>
        </YStack>
      ) : null}

      {devilFruit.skills.length > 0 ? (
        <YStack gap="$2">
          <GlowText level="label" tone="soft">
            {t('devilFruits.skills')}
          </GlowText>
          <YStack gap="$2">
            {devilFruit.skills.map((skill, index) => (
              <GlassPanel key={`${skill}-${index}`} tone="plastic" px="$3" py="$2" rounded="$card" elevate={0}>
                <GlowText level="label" flex={1}>
                  {skill}
                </GlowText>
              </GlassPanel>
            ))}
          </YStack>
        </YStack>
      ) : null}
    </YStack>
  )
}
