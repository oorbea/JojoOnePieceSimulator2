import type { LayoutChangeEvent, ScrollView } from 'react-native'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'

import type { ManualSectionId } from '../../lib/sections'
import type { ManualRules } from '../../types/manual'
import { AbilitiesSection } from './abilities-section'
import { BattleIQSection } from './battle-iq-section'
import { ConventionsSection } from './conventions-section'
import { EffectsSection } from './effects-section'
import { EtiquetteSection } from './etiquette-section'
import { HowSection } from './how-section'
import { ManualIndex } from './manual-index'
import { Body } from './manual-ui'
import { ModesSection } from './modes-section'
import { OddsSection } from './odds-section'

type Props = {
  rules: ManualRules
  active: ManualSectionId | null
  onSelectSection: (id: ManualSectionId) => void
  /** Layout reports the container needs to scroll to a section: the sections
   * column's and the row's offsets, and each section's offset in its column. */
  onMeasureRow: (y: number) => void
  onMeasureColumn: (y: number) => void
  onMeasureSection: (id: ManualSectionId, y: number) => void
  scrollRef?: React.Ref<ScrollView>
}

// The whole manual: title, table of contents, then every section. Pure UI -
// the rules arrive as props (generated from the game's own tables, see
// contracts/rules.ts), nothing here states a rule of its own.
export function ManualScreen({
  rules,
  active,
  onSelectSection,
  onMeasureRow,
  onMeasureColumn,
  onMeasureSection,
  scrollRef,
}: Props) {
  const { t } = useTranslation()
  const measure = (id: ManualSectionId) => (y: number) => onMeasureSection(id, y)

  return (
    <PageShell align="top" scroll maxWidth={1040} scrollRef={scrollRef}>
      <YStack width="100%" gap="$5" testID="manual-screen">
        <YStack gap="$2">
          <GlowText level="title" role="heading">
            {t('manual.title')}
          </GlowText>
          <Body>{t('manual.intro')}</Body>
        </YStack>

        <XStack
          width="100%"
          flexDirection="column"
          gap="$4"
          items="stretch"
          $md={{ flexDirection: 'row', gap: '$5', items: 'flex-start' }}
          onLayout={(e: LayoutChangeEvent) => onMeasureRow(e.nativeEvent.layout.y)}
        >
          <ManualIndex active={active} onSelect={onSelectSection} />
          <YStack
            flex={1}
            width="100%"
            gap="$4"
            onLayout={(e: LayoutChangeEvent) => onMeasureColumn(e.nativeEvent.layout.y)}
          >
            <HowSection limits={rules.limits} onMeasure={measure('how')} />
            <ModesSection limits={rules.limits} onMeasure={measure('modes')} />
            <AbilitiesSection odds={rules.odds} onMeasure={measure('abilities')} />
            <OddsSection odds={rules.odds} onMeasure={measure('odds')} />
            <EffectsSection
              tieRule={rules.tieRule}
              evolutions={rules.evolutions}
              statFloors={rules.statFloors}
              onMeasure={measure('effects')}
            />
            <BattleIQSection bands={rules.battleIQ} onMeasure={measure('battleIQ')} />
            <ConventionsSection
              conventions={rules.conventions}
              onMeasure={measure('conventions')}
            />
            <EtiquetteSection onMeasure={measure('etiquette')} />
          </YStack>
        </XStack>
      </YStack>
    </PageShell>
  )
}
