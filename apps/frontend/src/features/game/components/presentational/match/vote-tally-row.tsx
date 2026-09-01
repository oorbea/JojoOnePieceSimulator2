import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { VoterAvatar } from '@/features/game/components/presentational/match/voter-avatar'
import type { VoteTallyEntry } from '@/features/game/lib/vote-options'
import type { GameParticipant } from '@/features/game/types/game.types'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { MeterBar } from '@/shared/components/presentational/meter-bar'

type Props = {
  entry: VoteTallyEntry
  /** Normalizes the meter across the round - the largest count in it, never
   * below 1 (see voteTally). */
  maxCount: number
  /** Outlines the row the same way VoteBar outlines your own cast vote. */
  isWinner: boolean
  participants: GameParticipant[]
  selfParticipantId: string
}

// One option's row in a vote breakdown: label, a proportional meter, the
// count, and the avatars of whoever voted for it.
//
// Extracted from RoundResultPanel so the live round-result panel and the
// final result screen's per-round recap render an identical row rather than
// two copies that drift - the tones and labels both come from the same
// voteTally entry the vote bar itself used.
export function VoteTallyRow({
  entry,
  maxCount,
  isWinner,
  participants,
  selfParticipantId,
}: Props) {
  const { t } = useTranslation()
  const label = entry.labelKey ? t(entry.labelKey) : (entry.label ?? entry.id)

  return (
    <YStack gap="$1.5">
      <XStack items="center" gap="$2.5">
        <YStack flex={1}>
          <GlowText level="label">{label}</GlowText>
          <MeterBar
            value={entry.count / maxCount}
            tone={entry.tone === 'red' ? 'red' : entry.tone === 'green' ? 'green' : 'blue'}
            a11yLabel={t('game.match.result.votesA11y', { count: entry.count, option: label })}
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
            const participant = participants.find((p) => p.id === id)
            return participant ? (
              <VoterAvatar key={id} participant={participant} isSelf={id === selfParticipantId} />
            ) : null
          })}
        </XStack>
      ) : null}
    </YStack>
  )
}
