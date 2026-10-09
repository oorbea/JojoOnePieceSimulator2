import { useTranslation } from 'react-i18next'

import { Body, SectionCard, Topic } from './manual-ui'

type Props = { onMeasure?: (y: number) => void }

const GROUPS = ['honest', 'respect', 'host', 'spoilers'] as const

// Etiquette is the voters' code of honour: pure copy, no rule data behind it.
export function EtiquetteSection({ onMeasure }: Props) {
  const { t } = useTranslation()
  return (
    <SectionCard id="etiquette" title={t('manual.sections.etiquette')} onMeasure={onMeasure}>
      <Body>{t('manual.etiquette.intro')}</Body>
      {GROUPS.map((group) => (
        <Topic key={group} title={t(`manual.etiquette.${group}.title`)}>
          <Body>{t(`manual.etiquette.${group}.a`)}</Body>
          <Body>{t(`manual.etiquette.${group}.b`)}</Body>
        </Topic>
      ))}
    </SectionCard>
  )
}
