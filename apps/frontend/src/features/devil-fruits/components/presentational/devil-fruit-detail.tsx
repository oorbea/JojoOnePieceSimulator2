import { Apple } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Image, Pressable } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { ImageLightbox } from '@/shared/components/presentational/image-lightbox'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'

type Props = {
  devilFruit: DevilFruitResponse
}

// Same shape as StandDetail, minus the stat grid (Devil Fruits have none),
// plus the fruitType badge.
export function DevilFruitDetail({ devilFruit }: Props) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  return (
    <YStack gap="$4">
      <Pressable
        onPress={() => setIsPreviewOpen(true)}
        disabled={!devilFruit.picture}
        {...a11yProps(t('devilFruits.previewA11y', { name: devilFruit.name }), 'imagebutton')}
      >
        <YStack width="100%" height={220} rounded="$card" overflow="hidden" position="relative" bg="$plasticEdge">
          <InsetRing rounded="$card" />
          {devilFruit.picture ? (
            <Image
              source={{ uri: devilFruit.picture }}
              style={{ width: '100%', height: '100%' }}
              resizeMode="contain"
            />
          ) : (
            <YStack flex={1} items="center" justify="center">
              <Apple size={40} color="$strawHatRed" />
            </YStack>
          )}
        </YStack>
      </Pressable>
      <ImageLightbox
        visible={isPreviewOpen}
        uri={devilFruit.picture || null}
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
