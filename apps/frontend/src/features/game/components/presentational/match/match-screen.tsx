import { useTranslation } from 'react-i18next'
import { XStack } from 'tamagui'

import { MatchRoster } from '@/features/game/components/presentational/match/match-roster'
import { RevealStage } from '@/features/game/components/presentational/match/reveal-stage'
import { StageBanner } from '@/features/game/components/presentational/match/stage-banner'
import { VotingStatusBar } from '@/features/game/components/presentational/match/voting-status-bar'
import { ConnectionBanner } from '@/features/game/components/presentational/connection-banner'
import { currentRound } from '@/features/game/lib/match-rules'
import type { RevealPhaseKind } from '@/features/game/lib/loadout-reveal'
import type { LiveMatchState, SocketStatus } from '@/features/game/stores/game-socket.store'
import type { GameSnapshot, GameViewer } from '@/features/game/types/game.types'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  snapshot: GameSnapshot
  you: GameViewer
  socketStatus: SocketStatus
  nextRetryAt: number | null
  onRetryNow: () => void
  live: LiveMatchState
  revealPhase: RevealPhaseKind
  revealSlotIndex: number
  revealTotalSlots: number
  isRevealing: boolean
  onSkipReveal: () => void
  reducedMotion: boolean
  onAbort: () => void
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
  revealSlotIndex,
  revealTotalSlots,
  isRevealing,
  onSkipReveal,
  reducedMotion,
  onAbort,
}: Props) {
  const { t } = useTranslation()
  const round = currentRound(snapshot)

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
          slotIndex={revealSlotIndex}
          totalSlots={revealTotalSlots}
          onSkip={onSkipReveal}
          reducedMotion={reducedMotion}
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
        </>
      )}
    </>
  )
}
