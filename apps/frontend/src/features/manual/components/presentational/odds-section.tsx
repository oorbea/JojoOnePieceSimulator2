import { useTranslation } from 'react-i18next'
import { YStack } from 'tamagui'

import type { LoadoutSlot } from '@/shared/contracts/enums'
import { slotLevelKey, slotTraitKey } from '@/shared/lib/loadout-slots'

import type { ManualPct, ManualRules } from '../../types/manual'
import { Body, Muted, PctRow, SectionCard, Topic } from './manual-ui'

type Props = {
  odds: ManualRules['odds']
  onMeasure?: (y: number) => void
}

export function OddsSection({ odds, onMeasure }: Props) {
  const { t } = useTranslation()

  const levelRows = (slot: LoadoutSlot, pcts: readonly ManualPct[]) => (
    <YStack gap="$2">
      {pcts.map((p) => (
        <PctRow key={p.level} label={t(slotLevelKey(slot, p.level))} percent={p.percent} />
      ))}
    </YStack>
  )

  const hakiLabel = (set: { armament: boolean; observation: boolean; conqueror: boolean }) => {
    const names = [
      set.armament ? t('game.match.hakiType.armament') : null,
      set.observation ? t('game.match.hakiType.observation') : null,
      set.conqueror ? t('game.match.hakiType.conqueror') : null,
    ].filter((n): n is string => n !== null)
    return names.length > 0 ? names.join(' + ') : t('game.match.hakiType.none')
  }

  return (
    <SectionCard id="odds" title={t('manual.sections.odds')} onMeasure={onMeasure}>
      <Body>{t('manual.odds.intro')}</Body>

      <Topic title={t('manual.odds.pool.title')}>
        <Body>
          {t('manual.odds.pool.text', { stand: odds.noStandWeight, fruit: odds.noFruitWeight })}
        </Body>
      </Topic>

      <Topic title={t(slotTraitKey('SPIN'))}>{levelRows('SPIN', odds.spin)}</Topic>
      <Topic title={t(slotTraitKey('HAMON'))}>{levelRows('HAMON', odds.hamon)}</Topic>
      <Topic title={t(slotTraitKey('FRUIT_MASTERY'))}>
        <Muted>{t('manual.odds.mastery.note')}</Muted>
        {levelRows('FRUIT_MASTERY', odds.fruitMastery)}
      </Topic>
      <Topic title={t(slotTraitKey('PHYSICAL_FORM'))}>
        {levelRows('PHYSICAL_FORM', odds.physicalForm)}
      </Topic>

      <Topic title={t('manual.odds.haki.title')}>
        <Body>{t('manual.odds.haki.text')}</Body>
        <Muted>{t('manual.odds.haki.presence')}</Muted>
        <YStack gap="$2">
          <PctRow label={t('game.match.hakiType.armament')} percent={odds.hakiPresence.armament} />
          <PctRow
            label={t('game.match.hakiType.observation')}
            percent={odds.hakiPresence.observation}
          />
          <PctRow
            label={t('game.match.hakiType.conqueror')}
            percent={odds.hakiPresence.conqueror}
          />
        </YStack>
        <Muted>{t('manual.odds.haki.combos')}</Muted>
        <YStack gap="$2">
          {odds.hakiSets.map((set) => (
            <PctRow key={hakiLabel(set)} label={hakiLabel(set)} percent={set.percent} />
          ))}
        </YStack>
        <Muted>{t('manual.odds.haki.level')}</Muted>
        {levelRows('ARMAMENT_HAKI', odds.hakiMastery)}
      </Topic>

      <Body>{t('manual.odds.battleIQ')}</Body>
    </SectionCard>
  )
}
