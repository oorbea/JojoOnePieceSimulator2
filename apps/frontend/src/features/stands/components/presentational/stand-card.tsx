import { Pencil, Sparkles, Trash2 } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Image, Pressable } from 'react-native'
import { Spinner, XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { ImageLightbox } from '@/shared/components/presentational/image-lightbox'
import { InsetRing, WiiCard } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import { STAND_STAT_LABELS } from '@/features/stands/lib/stand-stats'
import type { StandResponse } from '@/features/stands/types/stands.types'

type Props = {
  stand: StandResponse
  onOpenDetail: () => void
  readOnly?: boolean
  onEdit?: () => void
  onDelete?: () => void
  isEditBusy?: boolean
}

// The card body (name/badges/stats) opens the read-only detail view on
// press; the thumbnail keeps its own Pressable for the full-size lightbox
// so the two never fight over the same tap - see stands.previewA11y vs
// stands.detailA11y for the two distinct affordances.
export function StandCard({ stand, onOpenDetail, readOnly, onEdit, onDelete, isEditBusy }: Props) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  return (
    <WiiCard padded width={280} gap="$3">
      <Pressable
        onPress={() => setIsPreviewOpen(true)}
        disabled={!stand.pictureThumb}
        {...a11yProps(t('stands.previewA11y', { name: stand.name }), 'imagebutton')}
      >
        <YStack width="100%" height={140} rounded="$card" overflow="hidden" position="relative" bg="$plasticEdge">
          <InsetRing rounded="$card" />
          {stand.pictureThumb ? (
            <Image source={{ uri: stand.pictureThumb }} style={{ width: '100%', height: '100%' }} />
          ) : (
            <YStack flex={1} items="center" justify="center">
              <Sparkles size={32} color="$standPurple" />
            </YStack>
          )}
        </YStack>
      </Pressable>
      <ImageLightbox visible={isPreviewOpen} uri={stand.picture} onClose={() => setIsPreviewOpen(false)} />

      <Pressable onPress={onOpenDetail} {...a11yProps(t('stands.detailA11y', { name: stand.name }), 'button')}>
        <YStack gap="$3">
          <YStack gap="$1">
            <GlowText level="heading" numberOfLines={1}>
              {stand.name}
            </GlowText>
            <XStack gap="$2" flexWrap="wrap">
              <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
                <GlowText level="label">{t(`enums.rarity.${stand.rarity}`)}</GlowText>
              </GlassPanel>
              {stand.evolvesFrom ? (
                <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
                  <GlowText level="label" numberOfLines={1}>
                    {t('stands.evolvesFromBadge', { name: stand.evolvesFrom.name })}
                  </GlowText>
                </GlassPanel>
              ) : null}
            </XStack>
          </YStack>

          <XStack flexWrap="wrap" gap="$2">
            {STAND_STAT_LABELS.map(({ key, label }) => (
              <YStack key={key} flexBasis={72} grow={1} minW={72} items="center" gap="$0.5">
                <GlowText level="label" tone="soft" fontSize="$1">
                  {label}
                </GlowText>
                <GlowText level="label" fontSize="$4">
                  {t(`enums.standStat.${stand[key]}`)}
                </GlowText>
              </YStack>
            ))}
          </XStack>
        </YStack>
      </Pressable>

      {readOnly ? null : (
        <XStack gap="$2" justify="flex-end">
          <GlossButton
            tone="blue"
            btnSize="sm"
            shape="circle"
            onPress={onEdit}
            disabled={isEditBusy}
            accessibilityLabel={t('stands.editA11y', { name: stand.name })}
          >
            {isEditBusy ? <Spinner size="small" color="white" /> : <Pencil size={16} color="white" />}
          </GlossButton>
          <GlossButton
            tone="red"
            btnSize="sm"
            shape="circle"
            onPress={onDelete}
            accessibilityLabel={t('stands.deleteA11y', { name: stand.name })}
          >
            <Trash2 size={16} color="white" />
          </GlossButton>
        </XStack>
      )}
    </WiiCard>
  )
}
