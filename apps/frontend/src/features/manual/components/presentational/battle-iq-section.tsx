import { useTranslation } from 'react-i18next'
import { YStack } from 'tamagui'

import type { ManualRules } from '../../types/manual'
import { Body, Muted, PctRow, SectionCard, Topic } from './manual-ui'

type Props = {
  bands: ManualRules['battleIQ']
  onMeasure?: (y: number) => void
}

export function BattleIQSection({ bands, onMeasure }: Props) {
  const { t } = useTranslation()
  const top = bands[bands.length - 1]?.hi ?? 255
  return (
    <SectionCard id="battleIQ" title={t('manual.sections.battleIQ')} onMeasure={onMeasure}>
      <Body>{t('manual.battleIQ.intro')}</Body>
      <Topic title={t('manual.battleIQ.bands.title')}>
        <Muted>{t('manual.battleIQ.bands.text', { top })}</Muted>
        <YStack gap="$2">
          {bands.map((band) => (
            <PctRow
              key={band.key}
              label={t(`enums.battleIQCategory.${band.key}`)}
              detail={`${band.lo}-${band.hi}`}
              percent={band.percent}
            />
          ))}
        </YStack>
      </Topic>
      <Topic title={t('manual.battleIQ.convention.title')}>
        <Body>{t('manual.battleIQ.convention.text')}</Body>
      </Topic>
    </SectionCard>
  )
}
