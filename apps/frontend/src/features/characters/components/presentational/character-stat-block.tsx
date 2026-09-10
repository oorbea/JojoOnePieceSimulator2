import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlowText } from '@/shared/components/presentational/glow-text'
import { formatBattleIQ } from '@/shared/lib/battle-iq'
import type { CharacterStatRow } from '@/features/characters/lib/character-stats'

type Props<T> = {
  character: T
  rows: CharacterStatRow<T>[]
}

// The one leaf in this feature that knows a character's kind - everything
// above it (character-card.tsx, character-detail.tsx) just passes in the
// right descriptor from character-stats.ts and stays generic. Renders each
// row as a stat pill; battleIq (no enumNamespace) goes through the shared
// WAIS-IV formatter instead of an enums.* lookup - same split
// loadout-modal.tsx's scalar-slots row already draws between numeric and
// enum traits.
export function CharacterStatBlock<T>({ character, rows }: Props<T>) {
  const { t } = useTranslation()

  return (
    <XStack flexWrap="wrap" gap="$2">
      {rows.map((row) => {
        const raw = row.value(character)
        const label =
          row.enumNamespace != null
            ? t(`enums.${row.enumNamespace}.${raw}`)
            : (formatBattleIQ(t, raw as number) ?? String(raw))

        return (
          <YStack key={row.key} flexBasis={96} grow={1} items="center" gap="$0.5">
            <GlowText level="label" tone="soft" fontSize="$3">
              {t(row.labelKey)}
            </GlowText>
            <GlowText level="heading" fontSize="$4" align="center">
              {label}
            </GlowText>
          </YStack>
        )
      })}
    </XStack>
  )
}
