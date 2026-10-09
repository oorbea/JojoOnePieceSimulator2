import { useTranslation } from 'react-i18next'

import { DetailModal } from '@/shared/components/presentational/detail-modal'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import type { ManualSectionId } from '@/shared/lib/manual-sections'

import type { ManualRules } from '../../types/manual'
import { AbilitiesSection } from './abilities-section'
import { BattleIQSection } from './battle-iq-section'
import { ConventionsSection } from './conventions-section'
import { EffectsSection } from './effects-section'
import { EtiquetteSection } from './etiquette-section'
import { HowSection } from './how-section'
import { ManualBareContext } from './manual-ui'
import { ModesSection } from './modes-section'
import { OddsSection } from './odds-section'

type Props = {
  /** The section to show; null keeps the overlay closed. */
  section: ManualSectionId | null
  rules: ManualRules
  onClose: () => void
  /** Leaves for the full manual at this section. */
  onOpenFull: (section: ManualSectionId) => void
}

function SectionBody({ section, rules }: { section: ManualSectionId; rules: ManualRules }) {
  switch (section) {
    case 'how':
      return <HowSection limits={rules.limits} />
    case 'modes':
      return <ModesSection limits={rules.limits} />
    case 'abilities':
      return <AbilitiesSection odds={rules.odds} />
    case 'odds':
      return <OddsSection odds={rules.odds} />
    case 'effects':
      return (
        <EffectsSection
          tieRule={rules.tieRule}
          evolutions={rules.evolutions}
          statFloors={rules.statFloors}
        />
      )
    case 'battleIQ':
      return <BattleIQSection bands={rules.battleIQ} />
    case 'conventions':
      return <ConventionsSection conventions={rules.conventions} />
    case 'etiquette':
      return <EtiquetteSection />
  }
}

// One manual section in a modal, opened from the lobby, the vote or the reveal
// without leaving them. It renders the very same section components as the
// manual page, bare (no card, the modal provides the title).
export function ManualOverlay({ section, rules, onClose, onOpenFull }: Props) {
  const { t } = useTranslation()
  return (
    <DetailModal
      visible={section !== null}
      title={section ? t(`manual.sections.${section}`) : ''}
      onClose={onClose}
      closeA11y={t('manual.help.close')}
      footer={
        section ? (
          <GlossButton
            tone="blue"
            btnSize="sm"
            onPress={() => onOpenFull(section)}
            accessibilityLabel={t('manual.help.openFull')}
            tooltip={t('manual.help.openFullHint')}
          >
            {t('manual.help.openFull')}
          </GlossButton>
        ) : undefined
      }
    >
      {section ? (
        <ManualBareContext.Provider value>
          <SectionBody section={section} rules={rules} />
        </ManualBareContext.Provider>
      ) : null}
    </DetailModal>
  )
}
