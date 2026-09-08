import { Pencil, Sparkles, Trash2 } from '@tamagui/lucide-icons-2'
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
//
// forwardRef targets the detail Pressable specifically (the card's main tab
// stop) - "Cargar más" in stands-screen.tsx moves focus there for the first
// newly-appended card after a page loads, per norma-teclado.md's
// keyboard-accessibility requirement.
export const StandCard = forwardRef<View, Props>(function StandCard(
  { stand, onOpenDetail, readOnly, onEdit, onDelete, isEditBusy },
  ref
) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [imageState, setImageState] = useState<LazyImageState>('queued')
  const [retryToken, setRetryToken] = useState(0)
  const uri = cardSource(stand)
  const isImageError = imageState === 'error'
  return (
    <WiiCard padded width={280} gap="$3">
      <Pressable
        onPress={() => (isImageError ? setRetryToken((n) => n + 1) : setIsPreviewOpen(true))}
        disabled={!uri}
        {...a11yProps(
          t(isImageError ? 'common.imageRetry' : 'stands.previewA11y', { name: stand.name }),
          'imagebutton'
        )}
      >
        <LazyImage
          uri={uri}
          lqip={lqipSource(stand)}
          height={140}
          pictureStatus={stand.pictureStatus}
          retryToken={retryToken}
          onStateChange={setImageState}
          fallback={<Sparkles size={32} color="$standPurple" />}
        />
      </Pressable>
      <ImageLightbox
        visible={isPreviewOpen}
        uri={fullSource(stand)}
        onClose={() => setIsPreviewOpen(false)}
      />

      <Pressable
        ref={ref}
        onPress={onOpenDetail}
        {...a11yProps(t('stands.detailA11y', { name: stand.name }), 'button')}
      >
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
            {isEditBusy ? (
              <Spinner size="small" color="white" />
            ) : (
              <Pencil size={16} color="white" />
            )}
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
})
