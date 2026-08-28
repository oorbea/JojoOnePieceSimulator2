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
import { formatCode } from '@/features/game/lib/game-code'
import { startGate } from '@/features/game/lib/lobby-rules'
import { shouldReveal } from '@/features/game/lib/loadout-reveal'
import { shareJoinCode } from '@/features/game/lib/share'
import { voteOptions } from '@/features/game/lib/vote-options'
import { useGameSocketStore } from '@/features/game/stores/game-socket.store'
import type { GameConfig, PoolFilter } from '@/features/game/types/game.types'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import { useReducedMotion } from '@/shared/hooks/use-reduced-motion'
import type { GameMode, LobbyVisibility, Manga } from '@/shared/lib/zod'
import { showErrorToast, showSuccessToast } from '@/shared/lib/toast'
import { AppError } from '@/shared/api/errors'
import { useDevilFruits } from '@/features/devil-fruits'
import { useStands } from '@/features/stands'

const GAUNTLET_MIN = 1
const GAUNTLET_MAX = 10
const VERSUS_MIN = 1
const VERSUS_MAX = 5

// Local edit-form state for the config panel, seeded from snapshot.config
// (and snapshot.mode, which lives outside GameConfig) whenever the lobby
// hasn't been locally edited yet. Kept separate from the live snapshot so
// typing/toggling doesn't fight incoming STATE frames from other clients.
type ConfigFormState = {
  mode: GameMode
  stageMangas: Manga[]
  powerMangas: Manga[]
  teamSize: number
  allowBots: boolean
  visibility: LobbyVisibility
  votingWindowSeconds: number
  poolFilter: PoolFilter
}

function configFormFromSnapshot(mode: GameMode, config: GameConfig): ConfigFormState {
  return {
    mode,
    stageMangas: config.stageMangas,
    powerMangas: config.powerMangas,
    teamSize: config.teamSize,
    allowBots: config.allowBots,
    visibility: config.visibility,
    votingWindowSeconds: config.votingWindowSeconds,
    poolFilter: config.poolFilter,
  }
}

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
    mangas: revealMangas,
    active: revealActive,
    markRevealed: socket.markAssignmentRevealed,
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
    showErrorToast(new AppError(socket.lastError.message, { code: socket.lastError.code }))
  }, [socket.lastError])

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
    setConfigForm((current) => {
      const base = current ?? form
      if (mode === 'GAUNTLET') {
        return {
          ...base,
          mode,
          allowBots: false,
          teamSize: Math.min(Math.max(base.teamSize, GAUNTLET_MIN), GAUNTLET_MAX),
        }
      }
      return { ...base, mode, teamSize: Math.min(Math.max(base.teamSize, VERSUS_MIN), VERSUS_MAX) }
    })
  }

  // Submits a full config replacement built from `next` (not from whatever
  // `form` happens to be at call time) - UPDATE_CONFIG is a full replacement,
  // and reading `form` here would race React's state batching for callers
  // that also just called setConfigForm in the same event (see
  // handleToggleConfigManga below).
  const submitConfigForm = (next: ConfigFormState) => {
    setConfigSaving(true)
    setConfigSaved(false)
    commands.updateConfig({
      mode: next.mode,
      stageMangas: next.stageMangas,
      powerMangas: next.powerMangas,
      abilitySource: snapshot.config.abilitySource,
      teamSize: next.teamSize,
      allowBots: next.allowBots,
      visibility: next.visibility,
      votingWindowSeconds: next.votingWindowSeconds,
      poolFilter: next.poolFilter,
    })
    setTimeout(() => {
      setConfigSaving(false)
      setConfigSaved(true)
    }, 500)
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
      configTeamSizeMin={form.mode === 'GAUNTLET' ? GAUNTLET_MIN : VERSUS_MIN}
      configTeamSizeMax={form.mode === 'GAUNTLET' ? GAUNTLET_MAX : VERSUS_MAX}
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
      configPoolFilter={form.poolFilter}
      configPoolActiveCount={configPoolActiveCount}
      configBanlistItems={banlistItems}
      onAddConfigBan={handleAddConfigBan}
      onRemoveConfigBan={handleRemoveConfigBan}
      onBanMatchingConfig={handleBanMatchingConfig}
      onClearConfigPoolFilter={handleClearConfigPoolFilter}
      configSaving={configSaving}
      configSaved={configSaved}
      onSubmitConfig={handleSubmitConfig}
      live={socket.live}
      revealPhase={loadoutReveal.phase}
      revealSlotIndex={loadoutReveal.slotIndex}
      revealTotalSlots={loadoutReveal.totalSlots}
      isRevealing={loadoutReveal.isRevealing}
      onSkipReveal={loadoutReveal.skip}
      reducedMotion={reducedMotion}
      onVote={handleVote}
      onSkipResult={socket.dismissResult}
    />
  )
}
