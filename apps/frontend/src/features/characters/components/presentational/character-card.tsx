import { Pencil, Trash2, Users } from '@tamagui/lucide-icons-2'
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
import type { PictureStatus, PowerRarity } from '@/shared/contracts/enums'
import { CharacterStatBlock } from '@/features/characters/components/presentational/character-stat-block'
import type { CharacterStatRow } from '@/features/characters/lib/character-stats'

type CharacterLike = {
  id: string
  name: string
  rarity: PowerRarity
  picture: string
  pictureThumb: string
  pictureCard?: string
  pictureLqip?: string
  pictureStatus: PictureStatus
}

type Props<T extends CharacterLike> = {
  character: T
  rows: CharacterStatRow<T>[]
  onOpenDetail: () => void
  readOnly?: boolean
  onEdit?: () => void
  onDelete?: () => void
  isEditBusy?: boolean
}

// Same grid-card recipe as StageCard - thumb well, name + rarity badge,
// stat block, actions. Kind-agnostic: the caller (characters-screen.tsx)
// picks which `rows` descriptor to pass, this component never branches on
// manga itself.
function CharacterCardInner<T extends CharacterLike>(
  { character, rows, onOpenDetail, readOnly, onEdit, onDelete, isEditBusy }: Props<T>,
  ref: React.ForwardedRef<View>
) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [imageState, setImageState] = useState<LazyImageState>('queued')
  const [retryToken, setRetryToken] = useState(0)
  const uri = cardSource(character)
  const isImageError = imageState === 'error'

  return (
    <WiiCard padded width={280} gap="$3">
      <Pressable
        onPress={() => (isImageError ? setRetryToken((n) => n + 1) : setIsPreviewOpen(true))}
        disabled={!uri}
        {...a11yProps(
          t(isImageError ? 'common.imageRetry' : 'characters.previewA11y', {
            name: character.name,
          }),
          'imagebutton'
        )}
      >
        <LazyImage
          uri={uri}
          lqip={lqipSource(character)}
          height={140}
          pictureStatus={character.pictureStatus}
          retryToken={retryToken}
          onStateChange={setImageState}
          fallback={<Users size={32} color="$wiiBlue" />}
        />
      </Pressable>
      <ImageLightbox
        visible={isPreviewOpen}
        uri={fullSource(character)}
        lqip={lqipSource(character)}
        onClose={() => setIsPreviewOpen(false)}
      />

      <Pressable
        ref={ref}
        onPress={onOpenDetail}
        {...a11yProps(t('characters.detailA11y', { name: character.name }), 'button')}
      >
        <YStack gap="$2">
          <YStack gap="$1">
            <GlowText level="heading" numberOfLines={1}>
              {character.name}
            </GlowText>
            <GlassPanel
              tone="plastic"
              px="$2.5"
              py="$1"
              rounded="$pill"
              elevate={0}
              self="flex-start"
            >
              <GlowText level="label">{t(`enums.rarity.${character.rarity}`)}</GlowText>
            </GlassPanel>
          </YStack>

          <CharacterStatBlock character={character} rows={rows} />
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
            accessibilityLabel={t('characters.editA11y', { name: character.name })}
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
            accessibilityLabel={t('characters.deleteA11y', { name: character.name })}
          >
            <Trash2 size={16} color="white" />
          </GlossButton>
        </XStack>
      )}
    </WiiCard>
  )
}

export const CharacterCard = forwardRef(CharacterCardInner) as <T extends CharacterLike>(
  props: Props<T> & { ref?: React.ForwardedRef<View> }
) => ReturnType<typeof CharacterCardInner>
