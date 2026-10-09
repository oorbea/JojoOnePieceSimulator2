import { useTranslation } from 'react-i18next'

import type { ManualRules } from '../../types/manual'
import { Body, SectionCard, Topic } from './manual-ui'

type Props = {
  limits: ManualRules['limits']
  onMeasure?: (y: number) => void
}

export function HowSection({ limits, onMeasure }: Props) {
  const { t } = useTranslation()
  const vote = {
    default: limits.votingDefaultSeconds,
    min: limits.votingMinSeconds,
    max: limits.votingMaxSeconds,
  }
  return (
    <SectionCard id="how" title={t('manual.sections.how')} onMeasure={onMeasure}>
      <Body>{t('manual.how.intro')}</Body>
      <Topic title={t('manual.how.host.title')}>
        <Body>{t('manual.how.host.text')}</Body>
      </Topic>
      <Topic step={1} title={t('manual.how.draw.title')}>
        <Body>{t('manual.how.draw.text')}</Body>
      </Topic>
      <Topic step={2} title={t('manual.how.vote.title')}>
        <Body>{t('manual.how.vote.text', vote)}</Body>
      </Topic>
      <Topic step={3} title={t('manual.how.resolve.title')}>
        <Body>{t('manual.how.resolve.text', { ext: limits.votingExtensionSeconds })}</Body>
      </Topic>
      <Topic title={t('manual.how.tie.title')}>
        <Body>{t('manual.how.tie.text')}</Body>
      </Topic>
      <Topic title={t('manual.how.skip.title')}>
        <Body>{t('manual.how.skip.text')}</Body>
      </Topic>
    </SectionCard>
  )
}
