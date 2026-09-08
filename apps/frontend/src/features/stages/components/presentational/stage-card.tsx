import { Map, Pencil, Trash2 } from '@tamagui/lucide-icons-2'
import { forwardRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Pressable, type View } from 'react-native'
import { Spinner, XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { ImageLightbox } from '@/shared/components/presentational/image-lightbox'
import { LazyImage, type LazyImageState } from '@/shared/components/presentational/lazy-image'
import { WiiCard } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import { cardSource, fullSource, lqipSource } from '@/shared/lib/picture-source'
import type { StageResponse } from '@/features/stages/types/stages.types'

type Props = {
  stage: StageResponse
  onOpenDetail: () => void
  readOnly?: boolean
  onEdit?: () => void
  onDelete?: () => void
  isEditBusy?: boolean
}

// Same grid-card recipe as StandCard - thumb well, name + badges, actions -
// swapping the stat grid (Stands have none here) for the stage's
// description, since that's the field an admin actually wants to preview.
// forwardRef targets the detail Pressable (the card's main tab stop) - see
// StandCard's identical doc for why "Cargar más" needs it.
export const StageCard = forwardRef<View, Props>(function StageCard(
  { stage, onOpenDetail, readOnly, onEdit, onDelete, isEditBusy },
  ref
) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [imageState, setImageState] = useState<LazyImageState>('queued')
  const [retryToken, setRetryToken] = useState(0)
  const uri = cardSource(stage)
  const isImageError = imageState === 'error'
  return (
    <WiiCard padded width={280} gap="$3">
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
          height={140}
          pictureStatus={stage.pictureStatus}
          retryToken={retryToken}
          onStateChange={setImageState}
          fallback={<Map size={32} color="$wiiBlue" />}
        />
      </Pressable>
      <ImageLightbox
        visible={isPreviewOpen}
        uri={fullSource(stage)}
        lqip={lqipSource(stage)}
        onClose={() => setIsPreviewOpen(false)}
      />

      <Pressable
        ref={ref}
        onPress={onOpenDetail}
        {...a11yProps(t('stages.detailA11y', { name: stage.name }), 'button')}
      >
        <YStack gap="$3">
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
      )}
    </WiiCard>
  )
})
