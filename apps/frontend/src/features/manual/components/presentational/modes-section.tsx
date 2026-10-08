import { useTranslation } from 'react-i18next'

import type { ManualRules } from '../../types/manual'
import { Body, SectionCard, Topic } from './manual-ui'

type Props = {
  limits: ManualRules['limits']
  onMeasure?: (y: number) => void
}

export function ModesSection({ limits, onMeasure }: Props) {
  const { t } = useTranslation()
  return (
    <SectionCard id="modes" title={t('manual.sections.modes')} onMeasure={onMeasure}>
      <Topic title={t('manual.modes.gauntlet.title')}>
        <Body>
          {t('manual.modes.gauntlet.text', {
            min: limits.gauntletMinPlayers,
            max: limits.gauntletMaxPlayers,
          })}
        </Body>
      </Topic>
      <Topic title={t('manual.modes.versus.title')}>
        <Body>
          {t('manual.modes.versus.text', {
            min: limits.versusMinTeamSize,
            max: limits.versusMaxTeamSize,
            teams: limits.versusTeamCount,
            rounds: limits.versusRounds,
          })}
        </Body>
      </Topic>
      <Topic title={t('manual.modes.bots.title')}>
        <Body>{t('manual.modes.bots.text')}</Body>
      </Topic>
      <Topic title={t('manual.modes.unique.title')}>
        <Body>{t('manual.modes.unique.text')}</Body>
      </Topic>
    </SectionCard>
  )
}
