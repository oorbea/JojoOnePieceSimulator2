import { useTranslation } from 'react-i18next'
import { YStack } from 'tamagui'

import type { LoadoutSlot } from '@/shared/contracts/enums'
import { slotLevelKey, slotTraitKey } from '@/shared/lib/loadout-slots'

import type { ManualRules } from '../../types/manual'
import { Body, ChipRow, SectionCard, Topic } from './manual-ui'

type Props = {
  odds: ManualRules['odds']
  onMeasure?: (y: number) => void
}

// The level scales come from the odds tables (every level the draw can
// produce), so a level added in the backend shows up here by itself.
export function AbilitiesSection({ odds, onMeasure }: Props) {
  const { t } = useTranslation()
  const scales: { slot: LoadoutSlot; levels: readonly string[] }[] = [
    { slot: 'SPIN', levels: odds.spin.map((p) => p.level) },
    { slot: 'HAMON', levels: odds.hamon.map((p) => p.level) },
    { slot: 'FRUIT_MASTERY', levels: odds.fruitMastery.map((p) => p.level) },
    { slot: 'PHYSICAL_FORM', levels: odds.physicalForm.map((p) => p.level) },
    { slot: 'ARMAMENT_HAKI', levels: odds.hakiMastery.map((p) => p.level) },
  ]
  return (
    <SectionCard id="abilities" title={t('manual.sections.abilities')} onMeasure={onMeasure}>
      <Topic title={t('manual.abilities.jojo.title')}>
        <Body>{t('manual.abilities.jojo.text')}</Body>
      </Topic>
      <Topic title={t('manual.abilities.onePiece.title')}>
        <Body>{t('manual.abilities.onePiece.text')}</Body>
      </Topic>
      <Body>{t('manual.abilities.both')}</Body>
      <Topic title={t('manual.abilities.none.title')}>
        <Body>{t('manual.abilities.none.text')}</Body>
      </Topic>
      <Topic title={t('manual.abilities.coupled.title')}>
        <Body>{t('manual.abilities.coupled.text')}</Body>
      </Topic>
      <Topic title={t('manual.abilities.scales.title')}>
        <YStack gap="$3">
          {scales.map(({ slot, levels }) => (
            <ChipRow
              key={slot}
              label={t(
                slot === 'ARMAMENT_HAKI' ? 'manual.abilities.scales.haki' : slotTraitKey(slot)
              )}
              chips={levels.map((level) => t(slotLevelKey(slot, level)))}
            />
          ))}
        </YStack>
      </Topic>
    </SectionCard>
  )
}
