import { useTranslation } from 'react-i18next'
import { XStack } from 'tamagui'

import { LoadoutSummaryStage } from '@/features/game/components/presentational/match/loadout-summary-stage'
import { MatchRoster } from '@/features/game/components/presentational/match/match-roster'
import { RevealStage } from '@/features/game/components/presentational/match/reveal-stage'
import { RoundResultPanel } from '@/features/game/components/presentational/match/round-result-panel'
import { StageBanner } from '@/features/game/components/presentational/match/stage-banner'
import { VoteBar } from '@/features/game/components/presentational/match/vote-bar'
import { VotingStatusBar } from '@/features/game/components/presentational/match/voting-status-bar'
import { ConnectionBanner } from '@/features/game/components/presentational/connection-banner'
import { currentRound, voteProgress } from '@/features/game/lib/match-rules'
import type { RevealPhaseKind } from '@/features/game/lib/loadout-reveal'
import { voteOptions } from '@/features/game/lib/vote-options'
import type { LiveMatchState, SocketStatus } from '@/features/game/stores/game-socket.store'
import type { GameSnapshot, GameViewer } from '@/features/game/types/game.types'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { useNow } from '@/shared/hooks/use-now'

type Props = {
  snapshot: GameSnapshot
  you: GameViewer
  socketStatus: SocketStatus
  nextRetryAt: number | null
  onRetryNow: () => void
  live: LiveMatchState
  revealPhase: RevealPhaseKind
  revealParticipantIndex: number
  revealSlotIndex: number
  revealTotalSlots: number
  isRevealing: boolean
  onSkipReveal: () => void
  onSummaryReady: () => void
  reducedMotion: boolean
  onAbort: () => void
  onVote: (optionId: string) => void
  onSkipResult: () => void
}

// Renders once the lobby has moved past LOBBY. While isRevealing, the
// sorteo overlay (RevealStage) takes over instead of the roster/stage
// banner - ASSIGNING is now genuinely observable (see GameService.
// scheduleRevealDelay), and this is exactly the state that overlay covers.
export function MatchScreen({
  snapshot,
  you,
  socketStatus,
  nextRetryAt,
  onRetryNow,
  live,
  revealPhase,
  revealParticipantIndex,
  revealSlotIndex,
  revealTotalSlots,
  isRevealing,
  onSkipReveal,
  onSummaryReady,
  reducedMotion,
  onAbort,
  onVote,
  onSkipResult,
}: Props) {
  const { t } = useTranslation()
  const round = currentRound(snapshot)
  const votingOpen = snapshot.state === 'VOTING' || snapshot.state === 'TIEBREAK'
  const now = useNow(1000, votingOpen && live.votingClosesAt !== null)
  const options = votingOpen ? voteOptions(snapshot, you) : []
  const progress = voteProgress(snapshot, live)
  const showResolvedPanel =
    snapshot.state === 'RESOLVING' && !!round?.result && !live.resultDismissed
  const showTiedPanel = snapshot.state === 'TIEBREAK' && !!round?.tiedVotes && !live.resultDismissed

  return (
    <>
      <ConnectionBanner status={socketStatus} nextRetryAt={nextRetryAt} onRetryNow={onRetryNow} />

      <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$2">
        <GlowText level="title">
          {t(`enums.gameMode.${snapshot.mode}`)} · {t(`enums.gameState.${snapshot.state}`)}
        </GlowText>
        {you.isHost ? (
          <GlossButton
            tone="red"
            btnSize="sm"
            onPress={onAbort}
            accessibilityLabel={t('game.abort.action')}
            tooltip={t('game.abort.title')}
          >
            {t('game.abort.action')}
          </GlossButton>
        ) : null}
      </XStack>

      {isRevealing ? (
        <RevealStage
          snapshot={snapshot}
          selfId={you.participantId}
          phase={revealPhase}
          participantIndex={revealParticipantIndex}
          slotIndex={revealSlotIndex}
          totalSlots={revealTotalSlots}
          readyCount={live.revealReadyCount}
          readyTotal={live.revealReadyTotal}
          onSkip={onSkipReveal}
          reducedMotion={reducedMotion}
        />
      ) : snapshot.state === 'SUMMARY' ? (
        <LoadoutSummaryStage
          snapshot={snapshot}
          selfId={you.participantId}
          summaryEndsAt={live.summaryEndsAt}
          readyCount={live.summaryReadyCount}
          readyTotal={live.summaryReadyTotal}
          onSkip={onSummaryReady}
        />
      ) : (
        <>
          {round ? <StageBanner stage={round.stage} roundIndex={round.index} /> : null}

          <VotingStatusBar
            isRevealing={false}
            onSkip={onSkipReveal}
            tiebreak={live.tiebreak}
            votingClosesAt={live.votingClosesAt}
            revealEndsAt={live.revealEndsAt}
            gameState={snapshot.state}
          />

          <MatchRoster snapshot={snapshot} selfId={you.participantId} />

          {showTiedPanel && round ? (
            <RoundResultPanel
              snapshot={snapshot}
              you={you}
              round={round}
              variant="tie"
              onSkip={onSkipResult}
            />
          ) : null}

          {showResolvedPanel && round ? (
            <RoundResultPanel
              snapshot={snapshot}
              you={you}
              round={round}
              variant="result"
              onSkip={onSkipResult}
            />
          ) : votingOpen ? (
            <VoteBar
              options={options}
              selectedOptionId={you.vote ?? null}
              cast={progress.cast}
              total={progress.total}
              closesAt={live.votingClosesAt}
              windowMs={snapshot.config.votingWindowSeconds * 1000}
              now={now}
              tiebreak={live.tiebreak}
              onVote={onVote}
            />
          ) : null}
        </>
      )}
    </>
  )
}
