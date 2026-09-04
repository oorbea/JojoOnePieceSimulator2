import { Sparkles } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Image, Pressable } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { ImageLightbox } from '@/shared/components/presentational/image-lightbox'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import { STAND_STAT_LABELS } from '@/features/stands/lib/stand-stats'
import type { StandResponse } from '@/features/stands/types/stands.types'

type Props = {
  stand: StandResponse
}

// Body for DetailModal - everything the public DTO carries: photo, badges,
// the full stat grid, description and skills. No id/pictureStatus/etc,
// those are admin-only concerns the modal never needs.
export function StandDetail({ stand }: Props) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  return (
    <YStack gap="$4">
      <Pressable
        onPress={() => setIsPreviewOpen(true)}
        disabled={!stand.picture}
        {...a11yProps(t('stands.previewA11y', { name: stand.name }), 'imagebutton')}
      >
        <YStack width="100%" height={220} rounded="$card" overflow="hidden" position="relative" bg="$plasticEdge">
          <InsetRing rounded="$card" />
          {stand.picture ? (
            <Image source={{ uri: stand.picture }} style={{ width: '100%', height: '100%' }} resizeMode="contain" />
          ) : (
            <YStack flex={1} items="center" justify="center">
              <Sparkles size={40} color="$standPurple" />
            </YStack>
          )}
        </YStack>
      </Pressable>
      <ImageLightbox visible={isPreviewOpen} uri={stand.picture || null} onClose={() => setIsPreviewOpen(false)} />

      <XStack gap="$2" flexWrap="wrap">
        <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
          <GlowText level="label">{t(`enums.rarity.${stand.rarity}`)}</GlowText>
        </GlassPanel>
        {stand.evolvesFrom ? (
          <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
            <GlowText level="label">{t('stands.evolvesFromBadge', { name: stand.evolvesFrom.name })}</GlowText>
          </GlassPanel>
        ) : null}
      </XStack>

      <XStack flexWrap="wrap" gap="$3">
        {STAND_STAT_LABELS.map(({ key, label }) => (
          <YStack key={key} flexBasis={80} grow={1} minW={80} items="center" gap="$0.5">
            <GlowText level="label" tone="soft" fontSize="$1">
              {label}
            </GlowText>
            <GlowText level="label" fontSize="$5">
              {t(`enums.standStat.${stand[key]}`)}
            </GlowText>
          </YStack>
        ))}
      </XStack>

      {stand.description ? (
        <YStack gap="$1">
          <GlowText level="label" tone="soft">
            {t('stands.description')}
          </GlowText>
          <GlowText level="label">{stand.description}</GlowText>
        </YStack>
      ) : null}

      {stand.skills.length > 0 ? (
        <YStack gap="$2">
          <GlowText level="label" tone="soft">
            {t('stands.skills')}
          </GlowText>
          <YStack gap="$2">
            {stand.skills.map((skill, index) => (
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
