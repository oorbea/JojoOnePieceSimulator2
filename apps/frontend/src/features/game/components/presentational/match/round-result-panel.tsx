import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { VoteTallyRow } from '@/features/game/components/presentational/match/vote-tally-row'
import { secondsUntil } from '@/features/game/lib/match-rules'
import { voteTally } from '@/features/game/lib/vote-options'
import type { GameRound, GameSnapshot, GameViewer } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { useNow } from '@/shared/hooks/use-now'

type Props = {
  snapshot: GameSnapshot
  you: GameViewer
  round: GameRound
  /** 'result': the round has a Result (winner, maybe coin flip) - votes
   * come off round.votes. 'tie': the round just tied and opened a revote -
   * votes come off round.tiedVotes, and there is no winner yet. */
  variant: 'result' | 'tie'
  onSkip: () => void
  /** The server's own RESOLVING deadline (live.resultEndsAt) - only ever
   * set alongside variant 'result' (see roundResultPanelVariant); null for
   * 'tie', which has no result timer of its own. Purely informational, same
   * spirit as VotingStatusBar's revealEndsAt countdown - the server alone
   * decides when RESOLVING actually ends, "skip" still only hides the panel
   * locally (see dismissResult), it never accelerates this deadline. */
  resultEndsAt: number | null
}

// Replaces VoteBar in the exact same slot once a round resolves (variant
// 'result', state RESOLVING) or ties (variant 'tie', state TIEBREAK, shown
// above the revote's own VoteBar) - the owner's explicit call (2026-08-28):
// per-option vote counts plus who voted for what (avatars only, name on
// hover/tooltip), not just a bare "X wins" label.
export function RoundResultPanel({ snapshot, you, round, variant, onSkip, resultEndsAt }: Props) {
  const { t } = useTranslation()
  const votes = (variant === 'result' ? round.votes : round.tiedVotes) ?? {}
  const { entries, maxCount } = voteTally(snapshot, you, votes)
  const noVotes = Object.keys(votes).length === 0
  const now = useNow(1000, resultEndsAt !== null)
  const secondsLeft = variant === 'result' ? secondsUntil(resultEndsAt, now) : null

  return (
    <GlassPanel tone="strong" rounded="$panel" p="$4" gap="$2.5" width="100%">
      <XStack items="center" justify="space-between" gap="$2">
        <GlowText level="label">
          {variant === 'tie'
            ? t('game.match.result.tie')
            : round.result?.winner
              ? t('game.match.result.winner', {
                  option:
                    entries.find((e) => e.id === round.result?.winner)?.label ??
                    round.result.winner,
                })
              : t('game.match.result.title', { index: round.index + 1 })}
        </GlowText>
        <GlossButton
          tone="glass"
          btnSize="sm"
          onPress={onSkip}
          accessibilityLabel={t('game.match.result.skipA11y')}
          tooltip={t('game.match.result.skipA11y')}
        >
          {t('game.match.result.skip')}
        </GlossButton>
      </XStack>

      {variant === 'result' && round.result?.decidedByCoinFlip ? (
        <GlassPanel tone="plastic" px="$2.5" py="$1" rounded="$pill" elevate={0} self="flex-start">
          <GlowText level="label">{t('game.match.result.coinFlip')}</GlowText>
        </GlassPanel>
      ) : null}

      {secondsLeft !== null ? (
        <GlowText level="label" tone="soft">
          {t('game.match.result.nextIn', { seconds: secondsLeft })}
        </GlowText>
      ) : null}

      {noVotes ? (
        <GlowText level="label" tone="soft">
          {t('game.match.result.noVotes')}
        </GlowText>
      ) : (
        <YStack gap="$2.5">
          {entries.map((entry) => (
            <VoteTallyRow
              key={entry.id}
              entry={entry}
              maxCount={maxCount}
              isWinner={variant === 'result' && round.result?.winner === entry.id}
              participants={snapshot.participants}
              selfParticipantId={you.participantId}
            />
          ))}
        </YStack>
      )}
    </GlassPanel>
  )
}
