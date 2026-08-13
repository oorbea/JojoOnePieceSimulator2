import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { FilterDisclosure } from '@/shared/components/presentational/filter-disclosure'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import type { FruitType, Rarity } from '@/shared/lib/zod'

const RARITIES: Rarity[] = ['COMMON', 'RARE', 'EPIC', 'LEGENDARY']
const FRUIT_TYPES: FruitType[] = [
  'PARAMECIA',
  'ZOAN',
  'LOGIA',
  'SPECIAL_PARAMECIA',
  'ANCIENT_ZOAN',
  'MYTHICAL_ZOAN',
]

type Props = {
  editable: boolean
  // A rarity toggle drives both standRarities and fruitRarities at once (one
  // "Rarities" chip row restricts both power kinds identically) - there is a
  // single game.pool.rarities label, not a standRarities/fruitRarities pair,
  // matching how the create-lobby form presents every other filter as one
  // control regardless of how many backend fields it fans out to.
  standRarities: Rarity[]
  fruitRarities: Rarity[]
  fruitTypes: FruitType[]
  activeCount: number
  onToggleRarity: (rarity: Rarity) => void
  onToggleFruitType: (fruitType: FruitType) => void
  onClearAll: () => void
  children?: ReactNode
}

// Collapsible rarity/fruit-type restriction fields for a lobby's power pool -
// same toggle-chip pattern as create-lobby-screen.tsx's manga selector,
// wrapped in the shared FilterDisclosure primitive so it stays collapsed by
// default. `children` is a slot for BanlistField so both power-pool concerns
// (rarity/fruit-type + banlist) share one disclosure and one activeCount
// badge instead of stacking two separate collapsible sections.
export function PowerPoolFields({
  editable,
  standRarities,
  fruitRarities,
  fruitTypes,
  activeCount,
  onToggleRarity,
  onToggleFruitType,
  onClearAll,
  children,
}: Props) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)
  const isRarityActive = (rarity: Rarity) => standRarities.includes(rarity) || fruitRarities.includes(rarity)

  return (
    <FilterDisclosure
      label={t('game.pool.title')}
      activeCount={activeCount}
      expanded={expanded}
      onToggle={() => setExpanded((v) => !v)}
      onClearAll={editable && activeCount > 0 ? onClearAll : undefined}
      clearLabel={t('game.pool.clearAll')}
    >
      <YStack width="100%" gap="$4">
        <YStack gap="$1.5">
          <GlowText level="label">{t('game.pool.rarities')}</GlowText>
          <XStack gap="$2" flexWrap="wrap">
            {RARITIES.map((rarity) => (
              <GlossButton
                key={rarity}
                tone={isRarityActive(rarity) ? 'blue' : 'glass'}
                btnSize="sm"
                disabled={!editable}
                onPress={() => onToggleRarity(rarity)}
                accessibilityLabel={t(`enums.rarity.${rarity}`)}
              >
                {t(`enums.rarity.${rarity}`)}
              </GlossButton>
            ))}
          </XStack>
        </YStack>

        <YStack gap="$1.5">
          <GlowText level="label">{t('game.pool.fruitTypes')}</GlowText>
          <XStack gap="$2" flexWrap="wrap">
            {FRUIT_TYPES.map((fruitType) => (
              <GlossButton
                key={fruitType}
                tone={fruitTypes.includes(fruitType) ? 'blue' : 'glass'}
                btnSize="sm"
                disabled={!editable}
                onPress={() => onToggleFruitType(fruitType)}
                accessibilityLabel={t(`enums.fruitType.${fruitType}`)}
              >
                {t(`enums.fruitType.${fruitType}`)}
              </GlossButton>
            ))}
          </XStack>
        </YStack>

        {children}
      </YStack>
    </FilterDisclosure>
  )
}
