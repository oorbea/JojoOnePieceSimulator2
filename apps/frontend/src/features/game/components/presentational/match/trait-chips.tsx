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
  if (!slot.i18nKey) return null
  if (slot.value === undefined && slot.numeric === undefined) return null

  const isNone = slot.value === 'NONE'
  const label = slot.numeric
    ? `${slot.numeric.score} · ${t(slot.numeric.categoryKey)}`
    : t(`enums.${enumNamespace(slot.key)}.${slot.value}`)

  return (
    <GlassPanel tone="plastic" px="$2" py="$1" rounded="$pill" elevate={0} opacity={isNone ? 0.55 : 1}>
      <GlowText level="label" fontSize="$1" tone={isNone ? 'soft' : undefined}>
        {t(slot.i18nKey)}: {label}
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
