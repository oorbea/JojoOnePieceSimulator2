import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { MatchRoster } from '@/features/game/components/presentational/match/match-roster'
import { VoteTallyRow } from '@/features/game/components/presentational/match/vote-tally-row'
import { matchRecap } from '@/features/game/lib/game-result'
import { teamTone, teamToneColor } from '@/features/game/lib/lobby-rules'
import type { GameSnapshot, GameViewer } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { useRovingGroup } from '@/shared/hooks/use-roving-group'

type Props = {
  snapshot: GameSnapshot
  you: GameViewer
  isHost: boolean
  onBackToLobbies: () => void
  onRematch: () => void
  /** Message from a rejected REMATCH, already translated. */
  rematchError?: string | null
  onModalOpenChange?: (open: boolean) => void
}

const ACTIONS_GROUP_ID = 'match-result-actions'

// The final screen of a match, rendered in place on /play/[id] once the
// snapshot goes FINISHED or ABORTED - the same route the match itself uses,
// following the precedent MatchScreen already set. Aborted games land here
// too (owner decision): no winner, an explicit notice instead, and the
// recap of whatever rounds were actually played first.
//
// Deliberately renders no ConnectionBanner: the socket is closed on purpose
// by the time this shows (the store stops reconnecting once terminal), so a
// "reconnecting..." banner would be reporting a problem that isn't one.
export function MatchResultScreen({
  snapshot,
  you,
  isHost,
  onBackToLobbies,
  onRematch,
  rematchError,
  onModalOpenChange,
}: Props) {
  const { t } = useTranslation()
  const recap = matchRecap(snapshot, you)

  // Both actions share one Tab stop, arrows move between them - the roving
  // group pattern this project uses for every small fixed control cluster
  // (see norma-teclado). The host has two, everyone else only "back".
  const actions = isHost ? 2 : 1
  const { getItemProps } = useRovingGroup({
    groupId: ACTIONS_GROUP_ID,
    count: actions,
    onActivate: (index) => {
      if (index === 0) onBackToLobbies()
      else onRematch()
    },
  })

  const winnerTeamIndex =
    recap.mode === 'VERSUS' && recap.winnerOptionId
      ? snapshot.teams.findIndex((team) => team.id === recap.winnerOptionId)
      : -1
  const winnerColor = winnerTeamIndex >= 0 ? teamToneColor(teamTone(winnerTeamIndex)) : undefined

  return (
    <YStack gap="$4" width="100%">
      <GlassPanel tone="strong" rounded="$panel" p="$4" gap="$2" width="100%">
        <GlowText level="label" tone="soft">
          {t('game.result.title')}
        </GlowText>

        {recap.aborted ? (
          <GlowText level="heading">{t('game.result.aborted')}</GlowText>
        ) : recap.mode === 'VERSUS' ? (
          <GlowText level="title" color={winnerColor as never}>
            {t('game.result.winnerTeam', { team: recap.winnerTeamName ?? recap.winnerOptionId })}
          </GlowText>
        ) : (
          <GlowText level="title">
            {recap.squadSurvived ? t('game.result.survived') : t('game.result.fell')}
          </GlowText>
        )}

        {recap.aborted ? (
          <GlowText level="label" tone="soft">
            {t('game.result.abortedNotice')}
          </GlowText>
        ) : null}

        <GlowText level="label" tone="soft">
          {t('game.result.roundsPlayed', { count: recap.roundsPlayed })}
        </GlowText>
      </GlassPanel>

      {recap.rounds.length > 0 ? (
        <GlassPanel tone="strong" rounded="$panel" p="$4" gap="$3" width="100%">
          <GlowText level="label">{t('game.result.recapTitle')}</GlowText>
          {recap.rounds.map((round) => (
            <YStack key={round.index} gap="$2">
              <GlowText level="label" tone="soft">
                {t('game.result.roundLabel', { index: round.index + 1, stage: round.stageName })}
              </GlowText>
              {round.decidedByCoinFlip ? (
                <GlassPanel
                  tone="plastic"
                  px="$2.5"
                  py="$1"
                  rounded="$pill"
                  elevate={0}
                  self="flex-start"
                >
                  <GlowText level="label">{t('game.match.result.coinFlip')}</GlowText>
                </GlassPanel>
              ) : null}
              {round.entries.map((entry) => (
                <VoteTallyRow
                  key={entry.id}
                  entry={entry}
                  maxCount={round.maxCount}
                  isWinner={round.winnerOptionId === entry.id}
                  participants={snapshot.participants}
                  selfParticipantId={you.participantId}
                />
              ))}
            </YStack>
          ))}
        </GlassPanel>
      ) : null}

      <GlassPanel tone="strong" rounded="$panel" p="$4" gap="$3" width="100%">
        <GlowText level="label">{t('game.result.outcomesTitle')}</GlowText>
        {recap.mode === 'VERSUS' ? (
          <GlowText level="label" tone="soft">
            {recap.outcomes.find((o) => o.isSelf)?.won
              ? t('game.result.youWon')
              : t('game.result.youLost')}
          </GlowText>
        ) : null}
        <MatchRoster
          snapshot={snapshot}
          selfId={you.participantId}
          onModalOpenChange={onModalOpenChange}
        />
      </GlassPanel>

      {rematchError ? (
        <GlowText level="label" color="$strawHatRed">
          {rematchError}
        </GlowText>
      ) : null}

      <XStack gap="$3" flexWrap="wrap">
        <GlossButton
          tone="glass"
          onPress={onBackToLobbies}
          accessibilityLabel={t('game.result.backA11y')}
          tooltip={t('game.result.backA11y')}
          {...getItemProps(0)}
        >
          {t('game.result.back')}
        </GlossButton>
        {isHost ? (
          <GlossButton
            tone="blue"
            flare
            onPress={onRematch}
            accessibilityLabel={t('game.result.rematchA11y')}
            tooltip={t('game.result.rematchA11y')}
            {...getItemProps(1)}
          >
            {t('game.result.rematch')}
          </GlossButton>
        ) : null}
      </XStack>
    </YStack>
  )
}
