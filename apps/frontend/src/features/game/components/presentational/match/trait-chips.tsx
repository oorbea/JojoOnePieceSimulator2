import { useTranslation } from 'react-i18next'
import { XStack } from 'tamagui'

import { loadoutTraits } from '@/features/game/lib/match-rules'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import type { GameLoadout } from '@/features/game/types/game.types'

type Props = { loadout: GameLoadout }

export function TraitChips({ loadout }: Props) {
  const { t } = useTranslation()

  return (
    <XStack gap="$1.5" flexWrap="wrap">
      {loadoutTraits(loadout).map((trait) => (
        <GlassPanel key={trait.key} tone="plastic" px="$2" py="$1" rounded="$pill" elevate={0}>
          <GlowText level="label" fontSize="$1">
            {t(trait.i18nKey)}: {t(`enums.${enumNamespace(trait.key)}.${trait.value}`)}
          </GlowText>
        </GlassPanel>
      ))}
    </XStack>
  )
}

// Maps a loadoutTraits key onto the enums i18n namespace that carries its
// value labels (added alongside spinLevel/hamonLevel/etc in the locale
// catalogs).
function enumNamespace(key: string): string {
  switch (key) {
    case 'spin':
      return 'spinLevel'
    case 'hamon':
      return 'hamonLevel'
    case 'fruitMastery':
      return 'fruitMastery'
    case 'physicalForm':
      return 'physicalForm'
    default:
      return 'hakiLevel'
  }
}
