import { Map, Pencil, Trash2 } from '@tamagui/lucide-icons-2'
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
import type { StageResponse } from '@/features/stages/types/stages.types'

type Props = {
  stage: StageResponse
  onEdit: () => void
  onDelete: () => void
  isEditBusy?: boolean
}

// Same grid-card recipe as StandCard - thumb well, name + badges, actions -
// swapping the stat grid (Stands have none here) for the stage's
// description, since that's the field an admin actually wants to preview.
export function StageCard({ stage, onEdit, onDelete, isEditBusy }: Props) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  return (
    <WiiCard padded width={280} gap="$3">
      <Pressable
        onPress={() => setIsPreviewOpen(true)}
        disabled={!stage.picture}
        {...a11yProps(t('stages.previewA11y', { name: stage.name }), 'imagebutton')}
      >
        <YStack
          width="100%"
          height={140}
          rounded="$card"
          overflow="hidden"
          position="relative"
          bg="$plasticEdge"
        >
          <InsetRing rounded="$card" />
          {/* Read `picture` (not `pictureThumb`) with a `|| null` fallback -
              see admin-panel-crud-ux-fixes.md: pictureThumb is empty until the
              worker finishes, and `??` wouldn't catch an empty string. */}
          {stage.picture || null ? (
            <Image source={{ uri: stage.picture }} style={{ width: '100%', height: '100%' }} />
          ) : (
            <YStack flex={1} items="center" justify="center">
              <Map size={32} color="$wiiBlue" />
            </YStack>
          )}
        </YStack>
      </Pressable>
      <ImageLightbox visible={isPreviewOpen} uri={stage.picture || null} onClose={() => setIsPreviewOpen(false)} />

      <YStack gap="$1">
        <GlowText level="heading" numberOfLines={1}>
          {stage.name}
        </GlowText>
        <XStack gap="$2" flexWrap="wrap">
          <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
            <GlowText level="label">{t(`enums.manga.${stage.manga}`)}</GlowText>
          </GlassPanel>
          <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
            <GlowText level="label">{t('stages.orderBadge', { order: stage.order })}</GlowText>
          </GlassPanel>
        </XStack>
      </YStack>

      <GlowText level="label" tone="soft" numberOfLines={2}>
        {stage.description}
      </GlowText>

      <XStack gap="$2" justify="flex-end">
        <GlossButton
          tone="blue"
          btnSize="sm"
          shape="circle"
          onPress={onEdit}
          disabled={isEditBusy}
          accessibilityLabel={t('stages.editA11y', { name: stage.name })}
        >
          {isEditBusy ? <Spinner size="small" color="white" /> : <Pencil size={16} color="white" />}
        </GlossButton>
        <GlossButton
          tone="red"
          btnSize="sm"
          shape="circle"
          onPress={onDelete}
          accessibilityLabel={t('stages.deleteA11y', { name: stage.name })}
        >
          <Trash2 size={16} color="white" />
        </GlossButton>
      </XStack>
    </WiiCard>
  )
}
