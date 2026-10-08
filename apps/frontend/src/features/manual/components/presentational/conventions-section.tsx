import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'

import {
  actorsLabelKey,
  conventionText,
  conventionTitle,
  groupByCategory,
  targetsLabelKey,
} from '../../lib/convention-copy'
import { subjectLabel } from '../../lib/subject-label'
import type { ManualConvention, ManualSubject } from '../../types/manual'
import { Body, Chip, ChipRow, Muted, SectionCard, Topic } from './manual-ui'

type Props = {
  conventions: readonly ManualConvention[]
  onMeasure?: (y: number) => void
}

export function ConventionsSection({ conventions, onMeasure }: Props) {
  const { t } = useTranslation()
  const labels = (subjects: readonly ManualSubject[]) => subjects.map((s) => subjectLabel(t, s))

  return (
    <SectionCard id="conventions" title={t('manual.sections.conventions')} onMeasure={onMeasure}>
      <Body>{t('manual.conventions.intro')}</Body>
      <Muted>{t('manual.conventions.judgementNote')}</Muted>
      {groupByCategory(conventions).map((group) => (
        <Topic key={group.category} title={t(`manual.conventions.categories.${group.category}`)}>
          <YStack gap="$3">
            {group.items.map((c) => (
              <GlassPanel
                key={c.id}
                tone="plastic"
                p="$4"
                gap="$2.5"
                elevate={0}
                rounded="$card"
                testID={`manual-convention-${c.id}`}
              >
                <XStack gap="$2" items="center" flexWrap="wrap">
                  <GlowText level="label" tone="ink" fontSize="$5">
                    {conventionTitle(t, c)}
                  </GlowText>
                  {c.judgement ? <Chip label={t('manual.conventions.judgement')} /> : null}
                </XStack>
                <Body>{conventionText(t, c)}</Body>
                <ChipRow label={t(actorsLabelKey(c))} chips={labels(c.actors)} />
                <ChipRow label={t(targetsLabelKey(c))} chips={labels(c.targets)} />
                <ChipRow label={t('manual.conventions.labels.immune')} chips={labels(c.immune)} />
                <ChipRow
                  label={t('manual.conventions.labels.attenuated')}
                  chips={labels(c.attenuated)}
                />
              </GlassPanel>
            ))}
          </YStack>
        </Topic>
      ))}
    </SectionCard>
  )
}
