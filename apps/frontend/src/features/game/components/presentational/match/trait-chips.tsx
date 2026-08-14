import { useTranslation } from 'react-i18next'

import type { LoadoutSlot } from '@/features/game/lib/match-rules'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = { slot: LoadoutSlot }

// A single scalar-trait chip (physicalForm/fruitMastery/hamon/haki/spin) -
// one slot at a time, so the reveal sequence (use-loadout-reveal.ts) can
// bring them in one by one alongside the stand/devilFruit blocks, in the
// same order LoadoutBuilder drew them. The caller (LoadoutCard) is
// responsible for only rendering slots with an i18nKey/value (i.e. never
// 'stand'/'devilFruit', which render their own dedicated block instead).
export function TraitChip({ slot }: Props) {
  const { t } = useTranslation()
  if (!slot.i18nKey || slot.value === undefined) return null

  return (
    <GlassPanel tone="plastic" px="$2" py="$1" rounded="$pill" elevate={0}>
      <GlowText level="label" fontSize="$1">
        {t(slot.i18nKey)}: {t(`enums.${enumNamespace(slot.key)}.${slot.value}`)}
      </GlowText>
    </GlassPanel>
  )
}

// Maps a slot key onto the enums i18n namespace that carries its value
// labels (spinLevel/hamonLevel/etc, added alongside the match feature).
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
