import { ChevronDown, ChevronUp, SlidersHorizontal, X } from '@tamagui/lucide-icons-2'
import type { ReactNode } from 'react'
import { XStack, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'

import { GlassPanel } from './glass-panel'
import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'

type Props = {
  label: string
  /** Number of filters currently set inside the disclosure - shown as a
   * badge on the header, 0 hides it. */
  activeCount: number
  expanded: boolean
  onToggle: () => void
  /** Omit to hide the "clear all" action entirely (e.g. nothing is active). */
  onClearAll?: () => void
  clearLabel: string
  children: ReactNode
}

// Collapsible "more filters" section - keeps a wide filter set (Stand's six
// stats + evolvesFrom, on top of search/rarity) from permanently occupying
// screen space when nobody's using them. Genuinely generic, not
// Stand-specific, in case Devil Fruits or Stages ever grow past their
// always-visible row.
export function FilterDisclosure({
  label,
  activeCount,
  expanded,
  onToggle,
  onClearAll,
  clearLabel,
  children,
}: Props) {
  return (
    <YStack width="100%" gap="$3">
      <XStack items="center" gap="$2" flexWrap="wrap">
        <XStack
          items="center"
          gap="$2"
          onPress={onToggle}
          cursor="pointer"
          {...a11yProps(label, 'button')}
        >
          <SlidersHorizontal size={16} color="$panelTextSoft" />
          <GlowText level="label">{label}</GlowText>
          {activeCount > 0 ? (
            <GlassPanel tone="plastic" px="$2" py="$0.5" rounded="$pill" elevate={0}>
              <GlowText level="label">{activeCount}</GlowText>
            </GlassPanel>
          ) : null}
          {expanded ? (
            <ChevronUp size={16} color="$panelTextSoft" />
          ) : (
            <ChevronDown size={16} color="$panelTextSoft" />
          )}
        </XStack>

        {onClearAll && activeCount > 0 ? (
          <GlossButton
            tone="glass"
            btnSize="sm"
            onPress={onClearAll}
            accessibilityLabel={clearLabel}
          >
            <X size={14} color="$panelText" /> {clearLabel}
          </GlossButton>
        ) : null}
      </XStack>

      {expanded ? (
        <XStack width="100%" flexWrap="wrap" gap="$3">
          {children}
        </XStack>
      ) : null}
    </YStack>
  )
}
