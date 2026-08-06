import { Apple, Pencil, Trash2 } from '@tamagui/lucide-icons-2'
import { Image } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { InsetRing, WiiCard } from '@/shared/components/presentational/wii-card'
import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'

type Props = {
  devilFruit: DevilFruitResponse
  onEdit: () => void
  onDelete: () => void
}

export function DevilFruitCard({ devilFruit, onEdit, onDelete }: Props) {
  return (
    <WiiCard padded width={280} gap="$3">
      <YStack width="100%" height={140} rounded="$card" overflow="hidden" position="relative" bg="$plasticEdge">
        <InsetRing rounded="$card" />
        {devilFruit.pictureThumb ? (
          <Image source={{ uri: devilFruit.pictureThumb }} style={{ width: '100%', height: '100%' }} />
        ) : (
          <YStack flex={1} items="center" justify="center">
            <Apple size={32} color="$strawHatRed" />
          </YStack>
        )}
      </YStack>

      <YStack gap="$1">
        <GlowText level="heading" numberOfLines={1}>
          {devilFruit.name}
        </GlowText>
        <XStack gap="$2" flexWrap="wrap">
          <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
            <GlowText level="label">{devilFruit.rarity}</GlowText>
          </GlassPanel>
          <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
            <GlowText level="label">{devilFruit.fruitType}</GlowText>
          </GlassPanel>
        </XStack>
      </YStack>

      <XStack gap="$2" justify="flex-end">
        <GlossButton
          tone="blue"
          btnSize="sm"
          shape="circle"
          onPress={onEdit}
          accessibilityLabel={`Edit ${devilFruit.name}`}
        >
          <Pencil size={16} color="white" />
        </GlossButton>
        <GlossButton
          tone="red"
          btnSize="sm"
          shape="circle"
          onPress={onDelete}
          accessibilityLabel={`Delete ${devilFruit.name}`}
        >
          <Trash2 size={16} color="white" />
        </GlossButton>
      </XStack>
    </WiiCard>
  )
}
