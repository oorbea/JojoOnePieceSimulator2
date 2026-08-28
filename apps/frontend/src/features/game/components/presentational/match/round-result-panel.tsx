import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { VoterAvatar } from '@/features/game/components/presentational/match/voter-avatar'
import { voteTally } from '@/features/game/lib/vote-options'
import type { GameRound, GameSnapshot, GameViewer } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { MeterBar } from '@/shared/components/presentational/meter-bar'

type Props = {
  snapshot: GameSnapshot
  you: GameViewer
  round: GameRound
  /** 'result': the round has a Result (winner, maybe coin flip) - votes
   * come off round.votes. 'tie': the round just tied and opened a revote -
   * votes come off round.tiedVotes, and there is no winner yet. */
  variant: 'result' | 'tie'
  onSkip: () => void
}

// Replaces VoteBar in the exact same slot once a round resolves (variant
// 'result', state RESOLVING) or ties (variant 'tie', state TIEBREAK, shown
// above the revote's own VoteBar) - the owner's explicit call (2026-08-28):
// per-option vote counts plus who voted for what (avatars only, name on
// hover/tooltip), not just a bare "X wins" label. The server alone decides
// when RESOLVING actually ends (see game.ResultDuration/CompleteRound) -
// this panel has no countdown of its own; "skip" only hides it locally
// (see dismissResult), it never accelerates the server's own timer.
export function RoundResultPanel({ snapshot, you, round, variant, onSkip }: Props) {
  const { t } = useTranslation()
  const votes = (variant === 'result' ? round.votes : round.tiedVotes) ?? {}
  const { entries, maxCount } = voteTally(snapshot, you, votes)
  const noVotes = Object.keys(votes).length === 0

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

      {noVotes ? (
        <GlowText level="label" tone="soft">
          {t('game.match.result.noVotes')}
        </GlowText>
      ) : (
        <YStack gap="$2.5">
          {entries.map((entry) => {
            const label = entry.labelKey ? t(entry.labelKey) : (entry.label ?? entry.id)
            const isWinner = variant === 'result' && round.result?.winner === entry.id
            return (
              <YStack key={entry.id} gap="$1.5">
                <XStack items="center" gap="$2.5">
                  <YStack flex={1}>
                    <GlowText level="label">{label}</GlowText>
                    <MeterBar
                      value={entry.count / maxCount}
                      tone={
                        entry.tone === 'red' ? 'red' : entry.tone === 'green' ? 'green' : 'blue'
                      }
                      a11yLabel={t('game.match.result.votesA11y', {
                        count: entry.count,
                        option: label,
                      })}
                    />
                  </YStack>
                  <GlowText level="label" tone="soft" minW={24}>
                    {entry.count}
                  </GlowText>
                </XStack>
                {entry.voterIds.length > 0 ? (
                  <XStack
                    flexWrap="wrap"
                    gap="$1.5"
                    style={
                      isWinner
                        ? ({
                            outlineWidth: 2,
                            outlineColor: '$channelActive',
                            outlineStyle: 'solid',
                          } as object)
                        : undefined
                    }
                  >
                    {entry.voterIds.map((id) => {
                      const participant = snapshot.participants.find((p) => p.id === id)
                      return participant ? (
                        <VoterAvatar
                          key={id}
                          participant={participant}
                          isSelf={id === you.participantId}
                        />
                      ) : null
                    })}
                  </XStack>
                ) : null}
              </YStack>
            )
          })}
        </YStack>
      )}
    </GlassPanel>
  )
}
