import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import type { LoadoutSlot } from '@/shared/contracts/enums'
import { slotLevelKey, slotTraitKey } from '@/shared/lib/loadout-slots'

import { subjectLabel } from '../../lib/subject-label'
import type { ManualChain, ManualStatFloor } from '../../types/manual'
import { Body, Chip, Divider, Muted, SectionCard, Topic } from './manual-ui'

type Props = {
  tieRule: string
  evolutions: readonly ManualChain[]
  statFloors: readonly ManualStatFloor[]
  onMeasure?: (y: number) => void
}

const CHANGE_COLOR: Record<string, string> = {
  NONE: '$panelTextSoft',
  LEVEL_RAISED: '$wiiBlue',
  EVOLVED: '$sunYellow',
}

export function EffectsSection({ tieRule, evolutions, statFloors, onMeasure }: Props) {
  const { t } = useTranslation()

  const chainTitle = (chain: ManualChain) => chain.stages.map((s) => s.name).join(' → ')

  const renderChain = (chain: ManualChain) => {
    const driver = chain.driver as LoadoutSlot
    const driverName = t(slotTraitKey(driver))
    return (
      <YStack key={chainTitle(chain)} gap="$3">
        <GlowText level="label" tone="ink" fontSize="$5">
          {chainTitle(chain)}
        </GlowText>
        <XStack flexWrap="wrap" gap="$2">
          {chain.stages
            .filter((s) => s.tier)
            .map((s) => (
              <Chip
                key={s.name}
                label={t('manual.effects.stageNeeds', {
                  stage: s.name,
                  stat: driverName,
                  level: t(slotLevelKey(driver, s.tier ?? '')),
                })}
              />
            ))}
        </XStack>
        <YStack gap="$3">
          {chain.rows.map((row) => (
            <GlassPanel key={row.drawn} tone="plastic" p="$3" gap="$2" elevate={0} rounded="$card">
              <GlowText level="label" tone="ink">
                {t('manual.effects.ifYouDraw', { name: row.drawn })}
              </GlowText>
              {row.outcomes.map((o) => (
                <XStack key={o.level} gap="$2" items="center">
                  <YStack
                    width={10}
                    height={10}
                    rounded="$circle"
                    bg={CHANGE_COLOR[o.change] as never}
                  />
                  <GlowText level="label" flex={1}>
                    {t(`manual.effects.outcome.${o.change}`, {
                      stat: driverName,
                      level: t(slotLevelKey(driver, o.level)),
                      after: t(slotLevelKey(driver, o.levelAfter)),
                      power: o.power,
                    })}
                  </GlowText>
                </XStack>
              ))}
            </GlassPanel>
          ))}
        </YStack>
      </YStack>
    )
  }

  return (
    <SectionCard id="effects" title={t('manual.sections.effects')} onMeasure={onMeasure}>
      <Body>{t('manual.effects.intro')}</Body>

      <Topic title={t('manual.effects.evolutions.title')}>
        <Body>{t('manual.effects.evolutions.text')}</Body>
        <Body>{t(`manual.effects.tie.${tieRule}`)}</Body>
        <Muted>{t('manual.effects.evolutions.banned')}</Muted>
      </Topic>
      <YStack gap="$5">{evolutions.map(renderChain)}</YStack>

      <Divider />

      <Topic title={t('manual.effects.floors.title')}>
        <Body>{t('manual.effects.floors.text')}</Body>
        <YStack gap="$2.5">
          {statFloors.map((rule) => (
            <XStack
              key={`${rule.target}-${rule.floor}-${rule.when.kind}-${(rule.when.names ?? []).join()}-${rule.when.min ?? ''}`}
              gap="$2"
              items="center"
              flexWrap="wrap"
            >
              <Chip label={subjectLabel(t, rule.when)} />
              <Body>
                {t('manual.effects.floors.line', {
                  stat: t(slotTraitKey(rule.target as LoadoutSlot)),
                  level: t(slotLevelKey(rule.target as LoadoutSlot, rule.floor)),
                })}
              </Body>
              {rule.cross ? <Chip label={t('manual.effects.floors.cross')} /> : null}
            </XStack>
          ))}
        </YStack>
        <Muted>{t('manual.effects.floors.crossNote')}</Muted>
      </Topic>

      <Topic title={t('manual.effects.family.title')}>
        <Body>{t('manual.effects.family.text')}</Body>
      </Topic>
    </SectionCard>
  )
}
