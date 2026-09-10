import { Users } from '@tamagui/lucide-icons-2'
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
import type { PictureStatus, PowerRarity } from '@/shared/contracts/enums'
import { CharacterStatBlock } from '@/features/characters/components/presentational/character-stat-block'
import type { CharacterStatRow } from '@/features/characters/lib/character-stats'

type CharacterLike = {
  name: string
  description: string
  rarity: PowerRarity
  picture: string
  pictureThumb: string
  pictureLqip?: string
  pictureStatus: PictureStatus
}

type Props<T extends CharacterLike> = {
  character: T
  rows: CharacterStatRow<T>[]
}

// Read-only breakdown a card's detail modal opens - same recipe as
// StageDetail, plus the stat block Stages never had.
export function CharacterDetail<T extends CharacterLike>({ character, rows }: Props<T>) {
  const { t } = useTranslation()
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [imageState, setImageState] = useState<LazyImageState>('queued')
  const [retryToken, setRetryToken] = useState(0)
  const uri = fullSource(character)
  const isImageError = imageState === 'error'

  return (
    <YStack gap="$4">
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
          height={220}
          contentFit="contain"
          pictureStatus={character.pictureStatus}
          retryToken={retryToken}
          onStateChange={setImageState}
          fallback={<Users size={40} color="$wiiBlue" />}
        />
      </Pressable>
      <ImageLightbox
        visible={isPreviewOpen}
        uri={uri}
        lqip={lqipSource(character)}
        onClose={() => setIsPreviewOpen(false)}
      />

      <XStack gap="$2" flexWrap="wrap">
        <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
          <GlowText level="label">{t(`enums.rarity.${character.rarity}`)}</GlowText>
        </GlassPanel>
      </XStack>

      <CharacterStatBlock character={character} rows={rows} />

      {character.description ? (
        <YStack gap="$1">
          <GlowText level="label" tone="soft">
            {t('characters.description')}
          </GlowText>
          <GlowText level="label">{character.description}</GlowText>
        </YStack>
      ) : null}
    </YStack>
  )
}
