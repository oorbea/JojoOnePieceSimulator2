import { useTranslation } from 'react-i18next'
import { useMedia } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { useNavInsets } from '@/shared/lib/nav-insets'
import { isWeb } from '@/shared/lib/web-blur'

import { MANUAL_SECTIONS, type ManualSectionId } from '../../lib/sections'

type Props = {
  active: ManualSectionId | null
  onSelect: (id: ManualSectionId) => void
}

// Table of contents: a wrapped row of chips on narrow screens, a side column
// on wide ones, where it also sticks to the top on web while the sections
// scroll past. Native has no sticky, so there it just sits above the content.
//
// Sticky is for the wide side column only: on a phone the index is a tall
// block of chips above the sections, and pinning it would keep it covering the
// text the player is trying to read.
export function ManualIndex({ active, onSelect }: Props) {
  const { t } = useTranslation()
  // The app's floating top bar covers the first ~navInsets.top px of the
  // viewport, so the sticky index has to stop below it, not at the very top.
  const navInsets = useNavInsets()
  const media = useMedia()
  const sticky = isWeb && media.md
  return (
    <GlassPanel
      tone="plastic"
      elevate={1}
      p="$3"
      gap="$2"
      width="100%"
      flexDirection="row"
      flexWrap="wrap"
      self="flex-start"
      $md={{ flexDirection: 'column', flexWrap: 'nowrap', width: 220 }}
      testID="manual-index"
      style={sticky ? ({ position: 'sticky', top: navInsets.top + 8 } as object) : undefined}
    >
      {MANUAL_SECTIONS.map((id) => {
        const label = t(`manual.sections.${id}`)
        return (
          <GlossButton
            key={id}
            tone={active === id ? 'blue' : 'glass'}
            btnSize="sm"
            onPress={() => onSelect(id)}
            accessibilityLabel={label}
            tooltip={t('manual.index.jump', { section: label })}
          >
            {label}
          </GlossButton>
        )
      })}
    </GlassPanel>
  )
}
