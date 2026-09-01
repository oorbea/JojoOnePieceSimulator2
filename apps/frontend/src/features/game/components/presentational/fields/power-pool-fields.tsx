import { AlertTriangle, Ban } from '@tamagui/lucide-icons-2'
import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import type { PoolCounts, PoolShortfall } from '@/features/game/lib/pool-stats'
import { FilterDisclosure } from '@/shared/components/presentational/filter-disclosure'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { a11yProps } from '@/shared/lib/a11y'
import type { Manga } from '@/shared/lib/zod'

type Props = {
  activeCount: number
  editable: boolean
  onClearAll: () => void
  /** Per-kind total/remaining split over the current catalog after the
   * banlist is applied - see lib/pool-stats.ts's computePoolCounts. */
  counts: PoolCounts
  /** Non-empty when a selected power manga's remaining pool can't seat
   * this lobby's configured teamSize - see poolShortfalls. Rendered as a
   * banner ABOVE the FilterDisclosure so it stays visible while collapsed. */
  shortfalls: PoolShortfall[]
  powerMangas: Manga[]
  children?: ReactNode
}

// Collapsible power-pool restriction section - banning only (rarity/fruit-
// type/stand-stat *whitelisting* was cut from the UI: the owner confirmed
// only banning specific powers is needed, not restricting the pool to only
// those categories, even though the backend's `PoolFilter` still carries
// those allowlist fields for a future pass). `children` is
// `BanByFilterFields` + `BanlistField` - kept as a slot instead of rendered
// directly here so this component owns only the disclosure chrome.
export function PowerPoolFields({
  editable,
  activeCount,
  onClearAll,
  counts,
  shortfalls,
  powerMangas,
  children,
}: Props) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)

  return (
    <YStack width="100%" gap="$2">
      {shortfalls.length > 0 ? (
        <GlassPanel tone="plastic" p="$3" gap="$2" width="100%">
          {shortfalls.map((shortfall) => (
            <XStack key={shortfall.kind} items="center" gap="$2" {...a11yProps(undefined, 'alert')}>
              <AlertTriangle size={16} color="$strawHatRedDeep" />
              <GlowText level="label" color="$strawHatRedDeep">
                {t(
                  shortfall.kind === 'STAND' ? 'game.pool.insufficientStands' : 'game.pool.insufficientFruits',
                  { remaining: shortfall.remaining, required: shortfall.required }
                )}
              </GlowText>
            </XStack>
          ))}
        </GlassPanel>
      ) : null}

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

          <YStack width="100%" gap="$1">
            <GlowText level="label">{t('game.pool.remainingTitle')}</GlowText>
            {powerMangas.includes('JOJO') ? (
              <GlowText level="label">
                {t('game.pool.remainingStands', { remaining: counts.STAND.remaining, total: counts.STAND.total })}
              </GlowText>
            ) : null}
            {powerMangas.includes('ONE_PIECE') ? (
              <GlowText level="label">
                {t('game.pool.remainingFruits', {
                  remaining: counts.DEVIL_FRUIT.remaining,
                  total: counts.DEVIL_FRUIT.total,
                })}
              </GlowText>
            ) : null}
          </YStack>

          {children}
        </YStack>
      </FilterDisclosure>
    </YStack>
  )
}
