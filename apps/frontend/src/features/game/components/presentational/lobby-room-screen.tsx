import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import type { BannableItem } from '@/features/game/components/presentational/fields/banlist-field'
import { ConnectionBanner } from '@/features/game/components/presentational/connection-banner'
import { JoinCodeCard } from '@/features/game/components/presentational/join-code-card'
import { LobbyConfigPanel } from '@/features/game/components/presentational/lobby-config-panel'
import { LobbyLockRow } from '@/features/game/components/presentational/lobby-lock-row'
import { MangaRow } from '@/features/game/components/presentational/manga-row'
import { MatchScreen } from '@/features/game/components/presentational/match/match-screen'
import { SquadRoster } from '@/features/game/components/presentational/squad-roster'
import { StartBar } from '@/features/game/components/presentational/start-bar'
import { TeamColumn } from '@/features/game/components/presentational/team-column'
import { teamTone, type Gate } from '@/features/game/lib/lobby-rules'
import type { RevealPhaseKind } from '@/features/game/lib/loadout-reveal'
import type { GameSnapshot, GameViewer, PoolFilter } from '@/features/game/types/game.types'
import type { GameMode, LobbyVisibility, Manga } from '@/shared/lib/zod'
import type { LiveMatchState, SocketStatus } from '@/features/game/stores/game-socket.store'
import { ConfirmSheet } from '@/shared/components/presentational/confirm-sheet'
import { FilterDisclosure } from '@/shared/components/presentational/filter-disclosure'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'

export type ConfirmSheetState = {
  title: string
  message: string
  confirmLabel: string
  onConfirm: () => void
} | null

type Props = {
  snapshot: GameSnapshot
  you: GameViewer
  socketStatus: SocketStatus
  nextRetryAt: number | null
  onRetryNow: () => void
  gate: Gate
  starting: boolean
  onStart: () => void
  onLeave: () => void
  onAbort: () => void
  onJoinTeam: (teamId: string) => void
  onKick: (participantId: string) => void
  onTransferHost: (participantId: string) => void
  onToggleLock: () => void
  onCopyCode: () => Promise<'copied' | 'shared' | 'failed'>
  onShareCode: () => Promise<'copied' | 'shared' | 'failed'>
  confirmSheet: ConfirmSheetState
  confirming: boolean
  onCancelConfirm: () => void
  configMode: GameMode
  onChangeConfigMode: (mode: GameMode) => void
  configMangas: Manga[]
  onToggleConfigManga: (manga: Manga) => void
  configTeamSize: number
  configTeamSizeMin: number
  configTeamSizeMax: number
  onChangeConfigTeamSize: (size: number) => void
  configAllowBots: boolean
  onToggleConfigAllowBots: () => void
  configVisibility: LobbyVisibility
  onToggleConfigVisibility: () => void
  configVotingWindowSeconds: number
  onChangeConfigVotingWindow: (seconds: number) => void
  configPoolFilter: PoolFilter
  configPoolActiveCount: number
  configBanlistItems: BannableItem[]
  onAddConfigBan: (id: string) => void
  onRemoveConfigBan: (id: string) => void
  onBanMatchingConfig: (ids: string[]) => void
  onClearConfigPoolFilter: () => void
  configSaving: boolean
  configSaved: boolean
  configError?: string
  onSubmitConfig: () => void
  live: LiveMatchState
  revealPhase: RevealPhaseKind
  revealSlotIndex: number
  revealTotalSlots: number
  isRevealing: boolean
  onSkipReveal: () => void
  reducedMotion: boolean
}

export function LobbyRoomScreen({
  snapshot,
  you,
  socketStatus,
  nextRetryAt,
  onRetryNow,
  gate,
  starting,
  onStart,
  onLeave,
  onAbort,
  onJoinTeam,
  onKick,
  onTransferHost,
  onToggleLock,
  onCopyCode,
  onShareCode,
  confirmSheet,
  confirming,
  onCancelConfirm,
  configMode,
  onChangeConfigMode,
  configMangas,
  onToggleConfigManga,
  configTeamSize,
  configTeamSizeMin,
  configTeamSizeMax,
  onChangeConfigTeamSize,
  configAllowBots,
  onToggleConfigAllowBots,
  configVisibility,
  onToggleConfigVisibility,
  configVotingWindowSeconds,
  onChangeConfigVotingWindow,
  configPoolFilter,
  configPoolActiveCount,
  configBanlistItems,
  onAddConfigBan,
  onRemoveConfigBan,
  onBanMatchingConfig,
  onClearConfigPoolFilter,
  configSaving,
  configSaved,
  configError,
  onSubmitConfig,
  live,
  revealPhase,
  revealSlotIndex,
  revealTotalSlots,
  isRevealing,
  onSkipReveal,
  reducedMotion,
}: Props) {
  const { t } = useTranslation()
  const capacity = snapshot.config.teamSize
  const [configExpanded, setConfigExpanded] = useState(false)

  return (
    <PageShell align="top" scroll maxWidth={1080}>
      {snapshot.state === 'LOBBY' ? (
        <>
          <ConnectionBanner status={socketStatus} nextRetryAt={nextRetryAt} onRetryNow={onRetryNow} />

          <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$2">
            <GlowText level="title">{t(`enums.gameMode.${snapshot.mode}`)}</GlowText>
            <LobbyLockRow locked={snapshot.locked} isHost={you.isHost} onToggle={onToggleLock} />
          </XStack>

          <JoinCodeCard
            code={snapshot.code}
            isPublic={snapshot.config.visibility === 'PUBLIC'}
            onCopy={onCopyCode}
            onShare={onShareCode}
          />

          <MangaRow
            mangas={configMangas}
            isHost={you.isHost}
            onToggle={onToggleConfigManga}
            saving={configSaving}
            saved={configSaved}
          />

          {snapshot.mode === 'VERSUS' ? (
            <YStack width="100%" gap="$3" $md={{ flexDirection: 'row' }}>
              {snapshot.teams.map((team, index) => (
                <TeamColumn
                  key={team.id}
                  team={team}
                  tone={teamTone(index)}
                  participants={snapshot.participants.filter((p) => p.teamId === team.id)}
                  hostId={snapshot.hostId}
                  selfId={you.participantId}
                  capacity={capacity}
                  canJoin={you.teamId !== team.id && snapshot.state === 'LOBBY'}
                  isHost={you.isHost}
                  onJoin={() => onJoinTeam(team.id)}
                  onKick={onKick}
                  onTransferHost={onTransferHost}
                />
              ))}
            </YStack>
          ) : (
            <SquadRoster
              participants={snapshot.participants}
              hostId={snapshot.hostId}
              selfId={you.participantId}
              capacity={capacity}
              isHost={you.isHost}
              onKick={onKick}
              onTransferHost={onTransferHost}
            />
          )}

          <FilterDisclosure
            label={t('game.config.title')}
            activeCount={0}
            expanded={configExpanded}
            onToggle={() => setConfigExpanded((v) => !v)}
            clearLabel=""
          >
            <LobbyConfigPanel
              isHost={you.isHost}
              mode={configMode}
              onChangeMode={onChangeConfigMode}
              mangas={configMangas}
              teamSize={configTeamSize}
              teamSizeMin={configTeamSizeMin}
              teamSizeMax={configTeamSizeMax}
              onChangeTeamSize={onChangeConfigTeamSize}
              allowBots={configAllowBots}
              onToggleAllowBots={onToggleConfigAllowBots}
              visibility={configVisibility}
              onToggleVisibility={onToggleConfigVisibility}
              votingWindowSeconds={configVotingWindowSeconds}
              onChangeVotingWindow={onChangeConfigVotingWindow}
              poolFilter={configPoolFilter}
              poolActiveCount={configPoolActiveCount}
              banlistItems={configBanlistItems}
              onAddBan={onAddConfigBan}
              onRemoveBan={onRemoveConfigBan}
              onBanMatching={onBanMatchingConfig}
              onClearPoolFilter={onClearConfigPoolFilter}
              saving={configSaving}
              saved={configSaved}
              error={configError}
              onSubmit={onSubmitConfig}
            />
          </FilterDisclosure>

          <StartBar isHost={you.isHost} gate={gate} starting={starting} onStart={onStart} onLeave={onLeave} onAbort={onAbort} />
        </>
      ) : (
        <MatchScreen
          snapshot={snapshot}
          you={you}
          socketStatus={socketStatus}
          nextRetryAt={nextRetryAt}
          onRetryNow={onRetryNow}
          live={live}
          revealPhase={revealPhase}
          revealSlotIndex={revealSlotIndex}
          revealTotalSlots={revealTotalSlots}
          isRevealing={isRevealing}
          onSkipReveal={onSkipReveal}
          reducedMotion={reducedMotion}
          onAbort={onAbort}
        />
      )}

      <ConfirmSheet
        visible={!!confirmSheet}
        title={confirmSheet?.title ?? ''}
        message={confirmSheet?.message ?? ''}
        confirmLabel={confirmSheet?.confirmLabel ?? ''}
        isConfirming={confirming}
        onConfirm={confirmSheet?.onConfirm ?? (() => {})}
        onCancel={onCancelConfirm}
      />
    </PageShell>
  )
}
