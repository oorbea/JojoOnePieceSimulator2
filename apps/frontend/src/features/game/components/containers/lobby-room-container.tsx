import { useRouter, useLocalSearchParams } from 'expo-router'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQueryClient } from '@tanstack/react-query'

import type { BannableItem } from '@/features/game/components/presentational/fields/banlist-field'
import {
  LobbyRoomScreen,
  type ConfirmSheetState,
} from '@/features/game/components/presentational/lobby-room-screen'
import { gameKeys } from '@/features/game/api/game.keys'
import { useGameCommands } from '@/features/game/hooks/use-game-commands'
import { useGameDetail } from '@/features/game/hooks/use-game-detail'
import { useGameSocket } from '@/features/game/hooks/use-game-socket'
import { useLoadoutReveal } from '@/features/game/hooks/use-loadout-reveal'
import { useMatchHotkeys } from '@/features/game/hooks/use-match-hotkeys'
import {
  applyModeChange,
  buildUpdateConfigPayload,
  configFormFromSnapshot,
  TEAM_SIZE_LIMITS,
  type ConfigFormState,
} from '@/features/game/lib/config-form'
import { formatCode } from '@/features/game/lib/game-code'
import { startGate } from '@/features/game/lib/lobby-rules'
import { shouldReveal } from '@/features/game/lib/loadout-reveal'
import { shareJoinCode } from '@/features/game/lib/share'
import { voteOptions } from '@/features/game/lib/vote-options'
import { useGameSocketStore } from '@/features/game/stores/game-socket.store'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import { useReducedMotion } from '@/shared/hooks/use-reduced-motion'
import type { GameMode, RevealSpeed, Manga } from '@/shared/lib/zod'
import { showErrorToast, showSuccessToast } from '@/shared/lib/toast'
import { AppError } from '@/shared/api/errors'
import { useDevilFruits } from '@/features/devil-fruits'
import { useStands } from '@/features/stands'

// REVEAL_SPEED_CYCLE fixes the order onCycleRevealSpeed steps through -
// slowest to fastest, so repeated presses read as "speeding the sorteo up"
// rather than a confusing wraparound.
const REVEAL_SPEED_CYCLE: RevealSpeed[] = ['RELAXED', 'NORMAL', 'SWIFT']

export function LobbyRoomContainer() {
  const router = useRouter()
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { id } = useLocalSearchParams<{ id: string }>()

  const socket = useGameSocket(id ?? null)
  const detail = useGameDetail(id ?? null, socket.status)
  const commands = useGameCommands()
  const resetSocket = useGameSocketStore((s) => s.reset)
  const standsQuery = useStands()
  const devilFruitsQuery = useDevilFruits()

  const banlistItems: BannableItem[] = useMemo(
    () => [
      ...(standsQuery.data ?? []).map((s) => ({
        id: s.id,
        name: s.name,
        kind: 'STAND' as const,
        rarity: s.rarity,
        stats: {
          attackPower: s.attackPower,
          speed: s.speed,
          attackRange: s.attackRange,
          endurance: s.endurance,
          precision: s.precision,
          potential: s.potential,
        },
      })),
      ...(devilFruitsQuery.data ?? []).map((f) => ({
        id: f.id,
        name: f.name,
        kind: 'DEVIL_FRUIT' as const,
        rarity: f.rarity,
        fruitType: f.fruitType,
      })),
    ],
    [standsQuery.data, devilFruitsQuery.data]
  )

  const [starting, setStarting] = useState(false)
  const [confirmSheet, setConfirmSheet] = useState<ConfirmSheetState>(null)
  const [configForm, setConfigForm] = useState<ConfigFormState | null>(null)
  const [configSeededFor, setConfigSeededFor] = useState<string | null>(null)
  const [configSaving, setConfigSaving] = useState(false)
  const [configSaved, setConfigSaved] = useState(false)
  const [configRequestId, setConfigRequestId] = useState<string | null>(null)
  const [configError, setConfigError] = useState<string | undefined>(undefined)

  const snapshot = socket.snapshot ?? detail.data?.game ?? null
  const you = socket.you ?? detail.data?.you ?? null
  const reducedMotion = useReducedMotion()

  // Defined unconditionally (before the !snapshot early return below) so
  // useMatchHotkeys - itself called before that same return, hooks can't be
  // conditional - can safely close over it regardless of render order.
  // Guards internally: nothing to vote for before snapshot/you exist.
  const handleVote = (optionId: string) => {
    if (!snapshot || !you) return
    if (!you.vote) {
      commands.vote(optionId)
      return
    }
    if (you.vote === optionId) return
    const options = voteOptions(snapshot, you)
    const chosen = options.find((o) => o.id === optionId)
    const label = chosen
      ? chosen.labelKey
        ? t(chosen.labelKey)
        : (chosen.label ?? optionId)
      : optionId
    setConfirmSheet({
      title: t('game.vote.changeTitle'),
      message: t('game.vote.changeMessage', { option: label }),
      confirmLabel: t('game.vote.changeConfirm'),
      tone: 'blue',
      onConfirm: () => {
        commands.vote(optionId)
        setConfirmSheet(null)
      },
    })
  }

  // Computed unconditionally (before the !snapshot early return below) since
  // hooks can't be called conditionally - mangas/active fall back to safe
  // empty/false values until snapshot actually arrives.
  const revealMangas = snapshot?.config.powerMangas ?? []
  const revealActive = !!snapshot && shouldReveal(socket.live, snapshot)
  const loadoutReveal = useLoadoutReveal({
    gameId: snapshot?.id ?? '',
    roundIndex: snapshot?.rounds.length ?? 0,
    mangas: revealMangas,
    participants: snapshot?.participants ?? [],
    speed: snapshot?.config.revealSpeed ?? 'NORMAL',
    active: revealActive,
    markRevealed: socket.markAssignmentRevealed,
    sendRevealReady: commands.revealReady,
    serverRevealMs: socket.live.revealMs,
  })

  // Computed unconditionally for the same reason as revealMangas/revealActive
  // above - useMatchHotkeys must be called before the early return.
  const votingOpen = !!snapshot && (snapshot.state === 'VOTING' || snapshot.state === 'TIEBREAK')
  const hotkeyOptions = snapshot && you ? voteOptions(snapshot, you) : []
  useMatchHotkeys({
    votingOpen,
    optionCount: hotkeyOptions.length,
    revealing: loadoutReveal.isRevealing,
    summaryOpen: snapshot?.state === 'SUMMARY',
    // ConfirmSheet is the one overlay this container itself owns and can
    // check synchronously. LoadoutModal's open/closed state lives inside
    // MatchRoster (deliberately - it's the component that already has
    // `mangas` in hand for the modal's content), so hotkeys aren't
    // suppressed while it's open yet - a small known gap, not a silent one.
    blocked: !!confirmSheet,
    onVote: (index) => {
      const option = hotkeyOptions[index]
      if (option) handleVote(option.id)
    },
    onSkipReveal: loadoutReveal.skip,
    onSkipSummary: commands.summaryReady,
  })

  // Reseed the edit form whenever a fresh CONFIG_UPDATED/STATE snapshot
  // lands for this game (tracked by a config-version key), never on every
  // render - otherwise the host's in-progress edits would be clobbered by
  // their own optimistic-less round trip or another client's STATE push.
  // Adjusted directly during render (React's own recommended pattern for
  // "reset state when some derived key changes") rather than in an effect -
  // no extra commit-then-re-render, and avoids `react-hooks/set-state-in-
  // effect` flagging the unconditional `setState` inside a `useEffect` body.
  const configVersionKey = snapshot
    ? `${snapshot.id}:${JSON.stringify(snapshot.config)}:${snapshot.mode}`
    : null
  if (snapshot && configVersionKey !== configSeededFor) {
    setConfigForm(configFormFromSnapshot(snapshot.mode, snapshot.config))
    setConfigSeededFor(configVersionKey)
    // A fresh snapshot landing is also how a successful UPDATE_CONFIG round
    // trip surfaces (there's no separate ack) - so treat it as "save
    // succeeded" too, clearing any in-flight saving/error state.
    if (configSeededFor !== null) {
      setConfigSaving(false)
      setConfigSaved(true)
      setConfigError(undefined)
    }
  }

  useEffect(() => {
    if (!socket.terminal) return
    const cleanupAndLeave = (message: string) => {
      queryClient.removeQueries({ queryKey: gameKeys.detail(id ?? '') })
      resetSocket()
      router.replace('/play' as never)
      showErrorToast(new AppError(message))
    }
    if (socket.terminal.kind === 'KICKED') cleanupAndLeave(t('game.terminal.kicked'))
    else if (socket.terminal.kind === 'ABORTED') cleanupAndLeave(t('game.terminal.abortedByHost'))
    else if (socket.terminal.kind === 'FINISHED') cleanupAndLeave(t('game.terminal.finished'))
    // eslint-disable-next-line react-hooks/exhaustive-deps -- fires once per terminal transition
  }, [socket.terminal])

  useEffect(() => {
    if (!socket.lastError) return
    // Toast fires for every WS error, config-related or not. Separately, if
    // this error answers our own in-flight config submit (requestId match),
    // also surface it inline on the config panel. Not clearing
    // configRequestId here - a later successful reseed supersedes it, and a
    // stale id from an already-completed submit can't spuriously re-match
    // since lastError itself only changes on a new incoming error.
    showErrorToast(new AppError(socket.lastError.message, { code: socket.lastError.code }))
    if (configRequestId && socket.lastError.requestId === configRequestId) {
      setConfigError(socket.lastError.message)
      setConfigSaving(false)
      setConfigSaved(false)
    }
  }, [socket.lastError, configRequestId])

  if (!snapshot || !you) {
    return <LoadingScreen />
  }

  const gate = startGate(snapshot, you)

  const handleStart = () => {
    setStarting(true)
    commands.start()
    setTimeout(() => setStarting(false), 3000)
  }

  const handleLeave = () => {
    setConfirmSheet({
      title: t('game.leave.title'),
      message: t('game.leave.message'),
      confirmLabel: t('game.leave.confirm'),
      onConfirm: () => {
        commands.leave()
        queryClient.removeQueries({ queryKey: gameKeys.detail(id ?? '') })
        resetSocket()
        router.replace('/play' as never)
      },
    })
  }

  const handleAbort = () => {
    setConfirmSheet({
      title: t('game.abort.title'),
      message: t('game.abort.message'),
      confirmLabel: t('game.abort.confirm'),
      onConfirm: () => {
        commands.abort()
        setConfirmSheet(null)
      },
    })
  }

  const handleKick = (participantId: string) => {
    setConfirmSheet({
      title: t('game.kick.title'),
      message: t('game.kick.message'),
      confirmLabel: t('game.kick.confirm'),
      onConfirm: () => {
        commands.kick(participantId)
        setConfirmSheet(null)
      },
    })
  }

  const handleTransferHost = (participantId: string) => {
    setConfirmSheet({
      title: t('game.transferHost.title'),
      message: t('game.transferHost.message'),
      confirmLabel: t('game.transferHost.confirm'),
      onConfirm: () => {
        commands.transferHost(participantId)
        setConfirmSheet(null)
      },
    })
  }

  const form = configForm ?? configFormFromSnapshot(snapshot.mode, snapshot.config)

  const handleChangeConfigMode = (mode: GameMode) => {
    setConfigForm((current) => applyModeChange(current ?? form, mode))
  }

  const configOffline = socket.status !== 'open'

  // Submits a full config replacement built from `next` (not from whatever
  // `form` happens to be at call time) - UPDATE_CONFIG is a full replacement,
  // and reading `form` here would race React's state batching for callers
  // that also just called setConfigForm in the same event (see
  // handleToggleConfigManga below). Saving/saved/error state now derives
  // purely from the real WS round trip: error via the lastError/requestId
  // match above, success via the reseed-on-CONFIG_UPDATED block above.
  const submitConfigForm = (next: ConfigFormState) => {
    if (configOffline) return
    setConfigSaving(true)
    setConfigSaved(false)
    setConfigError(undefined)
    const requestId = commands.updateConfig(buildUpdateConfigPayload(next, snapshot.config.abilitySource))
    setConfigRequestId(requestId)
  }

  // The manga selectors now live on the main lobby screen (MangaRow), always
  // visible instead of tucked inside "Lobby settings" - so, unlike every
  // other field in that settings panel, they have no separate "Save" step
  // and autosave on every toggle. Two independent axes (stages vs. powers).
  const handleToggleConfigStageManga = (manga: Manga) => {
    const stageMangas = form.stageMangas.includes(manga)
      ? form.stageMangas.filter((m) => m !== manga)
      : [...form.stageMangas, manga]
    const next = { ...form, stageMangas }
    setConfigForm(next)
    submitConfigForm(next)
  }

  const handleToggleConfigPowerManga = (manga: Manga) => {
    const powerMangas = form.powerMangas.includes(manga)
      ? form.powerMangas.filter((m) => m !== manga)
      : [...form.powerMangas, manga]
    const next = { ...form, powerMangas }
    setConfigForm(next)
    submitConfigForm(next)
  }

  const handleAddConfigBan = (powerId: string) => {
    setConfigForm((current) => {
      const base = current ?? form
      if (base.poolFilter.banned.includes(powerId)) return base
      return {
        ...base,
        poolFilter: { ...base.poolFilter, banned: [...base.poolFilter.banned, powerId] },
      }
    })
  }

  const handleRemoveConfigBan = (powerId: string) => {
    setConfigForm((current) => {
      const base = current ?? form
      return {
        ...base,
        poolFilter: {
          ...base.poolFilter,
          banned: base.poolFilter.banned.filter((b) => b !== powerId),
        },
      }
    })
  }

  const handleBanMatchingConfig = (powerIds: string[]) => {
    setConfigForm((current) => {
      const base = current ?? form
      const banned = [
        ...base.poolFilter.banned,
        ...powerIds.filter((id) => !base.poolFilter.banned.includes(id)),
      ]
      return { ...base, poolFilter: { ...base.poolFilter, banned } }
    })
  }

  const handleClearConfigPoolFilter = () => {
    setConfigForm((current) => {
      const base = current ?? form
      return {
        ...base,
        poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
      }
    })
  }

  // Whitelisting by rarity/fruit-type isn't exposed in the UI (owner call -
  // only banning specific powers is needed), so the active-count badge only
  // ever reflects the banlist.
  const configPoolActiveCount = form.poolFilter.banned.length

  const handleSubmitConfig = () => submitConfigForm(form)

  const shareMessage = t('game.code.shareMessage', { code: formatCode(snapshot.code) })

  return (
    <LobbyRoomScreen
      snapshot={snapshot}
      you={you}
      socketStatus={socket.status}
      nextRetryAt={socket.nextRetryAt}
      onRetryNow={socket.retryNow}
      gate={gate}
      starting={starting}
      onStart={handleStart}
      onLeave={handleLeave}
      onAbort={handleAbort}
      onJoinTeam={(teamId) => commands.switchTeam(teamId)}
      onMovePlayer={(participantId, teamId) => commands.movePlayer(participantId, teamId)}
      onKick={handleKick}
      onTransferHost={handleTransferHost}
      onToggleLock={() => commands.setLocked(!snapshot.locked)}
      onCopyCode={async () => {
        const result = await shareJoinCode(snapshot.code, shareMessage)
        if (result === 'copied') showSuccessToast(t('game.code.copied'))
        return result
      }}
      onShareCode={async () => {
        const result = await shareJoinCode(snapshot.code, shareMessage)
        if (result === 'shared') showSuccessToast(t('game.code.shared'))
        return result
      }}
      confirmSheet={confirmSheet}
      confirming={false}
      onCancelConfirm={() => setConfirmSheet(null)}
      configMode={form.mode}
      onChangeConfigMode={handleChangeConfigMode}
      configStageMangas={form.stageMangas}
      configPowerMangas={form.powerMangas}
      onToggleConfigStageManga={handleToggleConfigStageManga}
      onToggleConfigPowerManga={handleToggleConfigPowerManga}
      configTeamSize={form.teamSize}
      configTeamSizeMin={TEAM_SIZE_LIMITS[form.mode].min}
      configTeamSizeMax={TEAM_SIZE_LIMITS[form.mode].max}
      onChangeConfigTeamSize={(teamSize) => setConfigForm({ ...form, teamSize })}
      configAllowBots={form.allowBots}
      onToggleConfigAllowBots={() => setConfigForm({ ...form, allowBots: !form.allowBots })}
      configVisibility={form.visibility}
      onToggleConfigVisibility={() =>
        setConfigForm({ ...form, visibility: form.visibility === 'PUBLIC' ? 'PRIVATE' : 'PUBLIC' })
      }
      configVotingWindowSeconds={form.votingWindowSeconds}
      onChangeConfigVotingWindow={(votingWindowSeconds) =>
        setConfigForm({ ...form, votingWindowSeconds })
      }
      configSummaryDurationSeconds={form.summaryDurationSeconds}
      onChangeConfigSummaryDuration={(summaryDurationSeconds) =>
        setConfigForm({ ...form, summaryDurationSeconds })
      }
      configRevealSpeed={form.revealSpeed}
      onCycleConfigRevealSpeed={() => {
        const i = REVEAL_SPEED_CYCLE.indexOf(form.revealSpeed)
        const revealSpeed = REVEAL_SPEED_CYCLE[(i + 1) % REVEAL_SPEED_CYCLE.length]
        setConfigForm({ ...form, revealSpeed })
      }}
      configPoolFilter={form.poolFilter}
      configPoolActiveCount={configPoolActiveCount}
      configBanlistItems={banlistItems}
      onAddConfigBan={handleAddConfigBan}
      onRemoveConfigBan={handleRemoveConfigBan}
      onBanMatchingConfig={handleBanMatchingConfig}
      onClearConfigPoolFilter={handleClearConfigPoolFilter}
      configSaving={configSaving || configOffline}
      configSaved={configSaved}
      configError={configError}
      onSubmitConfig={handleSubmitConfig}
      live={socket.live}
      revealPhase={loadoutReveal.phase}
      revealParticipantIndex={loadoutReveal.participantIndex}
      revealSlotIndex={loadoutReveal.slotIndex}
      revealTotalSlots={loadoutReveal.totalSlots}
      isRevealing={loadoutReveal.isRevealing}
      onSkipReveal={loadoutReveal.skip}
      onSummaryReady={commands.summaryReady}
      reducedMotion={reducedMotion}
      onVote={handleVote}
      onSkipResult={socket.dismissResult}
    />
  )
}
