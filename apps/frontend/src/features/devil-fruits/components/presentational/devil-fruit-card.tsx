import { Apple, Pencil, Trash2 } from '@tamagui/lucide-icons-2'
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
import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'

type Props = {
  devilFruit: DevilFruitResponse
  onOpenDetail: () => void
  readOnly?: boolean
  onEdit?: () => void
  onDelete?: () => void
  isEditBusy?: boolean
}

// forwardRef targets the detail Pressable (the card's main tab stop) - see
// StandCard's identical doc for why "Cargar más" needs it.
export const DevilFruitCard = forwardRef<View, Props>(function DevilFruitCard(
  { devilFruit, onOpenDetail, readOnly, onEdit, onDelete, isEditBusy },
  ref
) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [imageState, setImageState] = useState<LazyImageState>('queued')
  const [retryToken, setRetryToken] = useState(0)
  const uri = cardSource(devilFruit)
  const isImageError = imageState === 'error'
  return (
    <WiiCard padded width={280} gap="$3">
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
          height={140}
          pictureStatus={devilFruit.pictureStatus}
          retryToken={retryToken}
          onStateChange={setImageState}
          fallback={<Apple size={32} color="$strawHatRed" />}
        />
      </Pressable>
      <ImageLightbox
        visible={isPreviewOpen}
        uri={fullSource(devilFruit)}
        lqip={lqipSource(devilFruit)}
        onClose={() => setIsPreviewOpen(false)}
      />

      <Pressable
        ref={ref}
        onPress={onOpenDetail}
        {...a11yProps(t('devilFruits.detailA11y', { name: devilFruit.name }), 'button')}
      >
        <YStack gap="$1">
          <GlowText level="heading" numberOfLines={1}>
            {devilFruit.name}
          </GlowText>
          <XStack gap="$2" flexWrap="wrap">
            <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
              <GlowText level="label">{t(`enums.rarity.${devilFruit.rarity}`)}</GlowText>
            </GlassPanel>
            <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
              <GlowText level="label">{t(`enums.fruitType.${devilFruit.fruitType}`)}</GlowText>
            </GlassPanel>
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
            accessibilityLabel={t('devilFruits.editA11y', { name: devilFruit.name })}
          >
            {isEditBusy ? <Spinner size="small" color="white" /> : <Pencil size={16} color="white" />}
          </GlossButton>
          <GlossButton
            tone="red"
            btnSize="sm"
            shape="circle"
            onPress={onDelete}
            accessibilityLabel={t('devilFruits.deleteA11y', { name: devilFruit.name })}
          >
            <Trash2 size={16} color="white" />
          </GlossButton>
        </XStack>
      )}
    </WiiCard>
  )
})
