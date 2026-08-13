import { Ban } from '@tamagui/lucide-icons-2'
import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { FilterDisclosure } from '@/shared/components/presentational/filter-disclosure'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  activeCount: number
  editable: boolean
  onClearAll: () => void
  children?: ReactNode
}

// Collapsible power-pool restriction section - banning only (rarity/fruit-
// type/stand-stat *whitelisting* was cut from the UI: the owner confirmed
// only banning specific powers is needed, not restricting the pool to only
// those categories, even though the backend's `PoolFilter` still carries
// those allowlist fields for a future pass). `children` is
// `BanByFilterFields` + `BanlistField` - kept as a slot instead of rendered
// directly here so this component owns only the disclosure chrome.
export function PowerPoolFields({ editable, activeCount, onClearAll, children }: Props) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)

  return (
    <FilterDisclosure
      label={t('game.pool.title')}
      activeCount={activeCount}
      expanded={expanded}
      onToggle={() => setExpanded((v) => !v)}
      onClearAll={editable && activeCount > 0 ? onClearAll : undefined}
      clearLabel={t('game.pool.clearAll')}
    >
      <YStack width="100%" gap="$2">
        <XStack items="center" gap="$1.5">
          <Ban size={14} color="$panelTextSoft" />
          <GlowText level="label">{t('game.pool.bannedHint')}</GlowText>
        </XStack>
        {children}
      </YStack>
    </FilterDisclosure>
  )
}
