import { Pencil, Sparkles, Trash2 } from '@tamagui/lucide-icons-2'
import { Image } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { InsetRing, WiiCard } from '@/shared/components/presentational/wii-card'
import type { StandResponse } from '@/features/stands/types/stands.types'

type Props = {
  stand: StandResponse
  onEdit: () => void
  onDelete: () => void
}

const STAT_LABELS: { key: keyof Pick<StandResponse, 'attackPower' | 'speed' | 'attackRange' | 'endurance' | 'precision' | 'potential'>; label: string }[] = [
  { key: 'attackPower', label: 'PWR' },
  { key: 'speed', label: 'SPD' },
  { key: 'attackRange', label: 'RNG' },
  { key: 'endurance', label: 'END' },
  { key: 'precision', label: 'PRE' },
  { key: 'potential', label: 'DEV' },
]

export function StandCard({ stand, onEdit, onDelete }: Props) {
  return (
    <WiiCard padded width={280} gap="$3">
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

      <YStack gap="$1">
        <GlowText level="heading" numberOfLines={1}>
          {stand.name}
        </GlowText>
        <XStack gap="$2" flexWrap="wrap">
          <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
            <GlowText level="label">{stand.rarity}</GlowText>
          </GlassPanel>
          {stand.evolvesFrom ? (
            <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0}>
              <GlowText level="label" numberOfLines={1}>
                From: {stand.evolvesFrom.name}
              </GlowText>
            </GlassPanel>
          ) : null}
        </XStack>
      </YStack>

      <XStack flexWrap="wrap" gap="$2">
        {STAT_LABELS.map(({ key, label }) => (
          <YStack key={key} width={80} items="center" gap="$0.5">
            <GlowText level="label" tone="soft" fontSize="$1">
              {label}
            </GlowText>
            <GlowText level="label" fontSize="$4">
              {stand[key]}
            </GlowText>
          </YStack>
        ))}
      </XStack>

      <XStack gap="$2" justify="flex-end">
        <GlossButton tone="blue" btnSize="sm" shape="circle" onPress={onEdit} accessibilityLabel={`Edit ${stand.name}`}>
          <Pencil size={16} color="white" />
        </GlossButton>
        <GlossButton tone="red" btnSize="sm" shape="circle" onPress={onDelete} accessibilityLabel={`Delete ${stand.name}`}>
          <Trash2 size={16} color="white" />
        </GlossButton>
      </XStack>
    </WiiCard>
  )
}
