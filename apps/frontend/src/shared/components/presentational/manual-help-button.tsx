import { BookOpen } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'

import { useOpenManual } from '@/shared/lib/manual-overlay'
import type { ManualSectionId } from '@/shared/lib/manual-sections'

import { GlossButton } from './gloss-button'

type Props = { section: ManualSectionId }

// Small book icon that opens the manual's `section` in an overlay, so a player
// can check a rule without leaving the lobby or the match. Every button gets a
// tooltip (project norm); here it also names the section it opens.
export function ManualHelpButton({ section }: Props) {
  const { t } = useTranslation()
  const openManual = useOpenManual()
  const label = t('manual.help.open', { section: t(`manual.sections.${section}`) })
  return (
    <GlossButton
      tone="glass"
      btnSize="sm"
      shape="circle"
      tooltip={label}
      accessibilityLabel={label}
      onPress={() => openManual(section)}
    >
      <BookOpen size={14} color="$panelTextSoft" />
    </GlossButton>
  )
}
