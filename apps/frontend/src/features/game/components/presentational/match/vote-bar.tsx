import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import type { VoteOption } from '@/features/game/lib/vote-options'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { MeterBar } from '@/shared/components/presentational/meter-bar'
import { a11yProps } from '@/shared/lib/a11y'
import { useRovingGroup } from '@/shared/hooks/use-roving-group'
import { isWeb } from '@/shared/lib/web-blur'

type Props = {
  options: VoteOption[]
  selectedOptionId: string | null
  cast: number
  total: number
  /** Epoch ms - null renders no timer (voting-status-bar's own label still
   * shows the state, e.g. right after a TIEBREAK_OPENED frame before its
   * closesAt/votingEndsAt has resolved). */
  closesAt: number | null
  /** The full window duration in ms, for the draining MeterBar - config's
   * votingWindowSeconds, host-configurable per lobby. */
  windowMs: number
  now: number
  tiebreak: boolean
  onVote: (optionId: string) => void
}

// The fixed bottom bar for an open voting (or revote) window: a draining
// progress bar + seconds, the live human-vote count, then the options
// themselves as a keyboard-navigable radio group (roving tabindex - one Tab
// stop for the whole group, arrows move between options, Enter/Space
// votes). Anchored to the bottom of the match view on web only (native has
// no CSS position:sticky - it just renders last in the scroll content,
// per the a11y/pointerEvents-leak norm: platform branching stays inside
// style{}, never as a top-level prop).
export function VoteBar({ options, selectedOptionId, cast, total, closesAt, windowMs, now, tiebreak, onVote }: Props) {
  const { t } = useTranslation()
  const selectedIndex = Math.max(
    0,
    options.findIndex((o) => o.id === selectedOptionId)
  )
  const { getItemProps } = useRovingGroup({
    groupId: 'vote-option',
    count: options.length,
    initialIndex: selectedIndex,
    onActivate: (index) => onVote(options[index].id),
  })

  if (options.length === 0) return null

  const remainingMs = closesAt !== null ? Math.max(0, closesAt - now) : null
  const progress = remainingMs !== null && windowMs > 0 ? remainingMs / windowMs : null
  const seconds = remainingMs !== null ? Math.round(remainingMs / 1000) : null

  return (
    <GlassPanel
      tone="strong"
      rounded="$panel"
      p="$4"
      gap="$2.5"
      width="100%"
      style={isWeb ? ({ position: 'sticky', bottom: 0 } as object) : undefined}
    >
      {progress !== null ? (
        <XStack items="center" gap="$2.5">
          <YStack flex={1}>
            <MeterBar value={progress} tone={progress < 0.2 ? 'red' : 'blue'} a11yLabel={t('game.vote.timeLeftA11y')} />
          </YStack>
          <GlowText level="label" tone="soft" minW={32}>
            {seconds}s
          </GlowText>
        </XStack>
      ) : null}

      <GlowText level="label">
        {tiebreak ? `${t('game.vote.tiebreakHint')} · ` : ''}
        {t('game.vote.progress', { cast, total })}
      </GlowText>
      {selectedOptionId ? <GlowText level="label" tone="soft">{t('game.vote.youVoted')}</GlowText> : null}

      <XStack
        gap="$2.5"
        flexWrap="wrap"
        {...a11yProps(t('game.vote.groupA11y'), 'radiogroup')}
      >
        {options.map((option, index) => {
          const itemProps = getItemProps(index)
          const isSelected = option.id === selectedOptionId
          const label = option.labelKey ? t(option.labelKey) : (option.label ?? option.id)
          return (
            <GlossButton
              key={option.id}
              tone={option.tone}
              btnSize="md"
              onPress={() => onVote(option.id)}
              accessibilityLabel={
                option.isOwnTeam ? `${label} (${t('game.vote.yourTeam')})` : label
              }
              tooltip={option.isOwnTeam ? t('game.vote.yourTeam') : null}
              a11yRole="radio"
              a11yChecked={isSelected}
              tabIndex={itemProps.tabIndex}
              onKeyDown={itemProps.onKeyDown}
              onFocus={itemProps.onFocus}
              id={itemProps.id}
              style={
                isSelected
                  ? ({ outlineWidth: 3, outlineColor: '$channelActive', outlineStyle: 'solid' } as object)
                  : undefined
              }
            >
              {label}
              {option.isOwnTeam ? ` · ${t('game.vote.yourTeam')}` : ''}
            </GlossButton>
          )
        })}
      </XStack>

      {isWeb ? (
        <GlowText level="label" tone="soft" fontSize="$1">
          {t('game.vote.hint.keys')}
        </GlowText>
      ) : null}
    </GlassPanel>
  )
}
