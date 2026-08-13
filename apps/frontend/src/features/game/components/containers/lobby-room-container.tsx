import { useRouter, useLocalSearchParams } from 'expo-router'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQueryClient } from '@tanstack/react-query'

import type { BannableItem } from '@/features/game/components/presentational/fields/banlist-field'
import { LobbyRoomScreen, type ConfirmSheetState } from '@/features/game/components/presentational/lobby-room-screen'
import { gameKeys } from '@/features/game/api/game.keys'
import { useGameCommands } from '@/features/game/hooks/use-game-commands'
import { useGameDetail } from '@/features/game/hooks/use-game-detail'
import { useGameSocket } from '@/features/game/hooks/use-game-socket'
import { formatCode } from '@/features/game/lib/game-code'
import { startGate } from '@/features/game/lib/lobby-rules'
import { shareJoinCode } from '@/features/game/lib/share'
import { useGameSocketStore } from '@/features/game/stores/game-socket.store'
import type { GameConfig, PoolFilter } from '@/features/game/types/game.types'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
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
  mangas: Manga[]
  teamSize: number
  allowBots: boolean
  visibility: LobbyVisibility
  votingWindowSeconds: number
  poolFilter: PoolFilter
}

function configFormFromSnapshot(mode: GameMode, config: GameConfig): ConfigFormState {
  return {
    mode,
    mangas: config.mangas,
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
      ...(standsQuery.data ?? []).map((s) => ({ id: s.id, name: s.name, kind: 'STAND' as const })),
      ...(devilFruitsQuery.data ?? []).map((f) => ({ id: f.id, name: f.name, kind: 'DEVIL_FRUIT' as const })),
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

  // Reseed the edit form whenever a fresh CONFIG_UPDATED/STATE snapshot
  // lands for this game (tracked by a config-version key), never on every
  // render - otherwise the host's in-progress edits would be clobbered by
  // their own optimistic-less round trip or another client's STATE push.
  const configVersionKey = snapshot
    ? `${snapshot.id}:${JSON.stringify(snapshot.config)}:${snapshot.mode}`
    : null
  useEffect(() => {
    if (!snapshot || configVersionKey === configSeededFor) return
    setConfigForm(configFormFromSnapshot(snapshot.mode, snapshot.config))
    setConfigSeededFor(configVersionKey)
    // eslint-disable-next-line react-hooks/exhaustive-deps -- keyed on configVersionKey, not snapshot identity
  }, [configVersionKey])

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
    // eslint-disable-next-line react-hooks/exhaustive-deps -- fires once per new error
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

  const handleToggleConfigManga = (manga: Manga) => {
    setConfigForm((current) => {
      const base = current ?? form
      const mangas = base.mangas.includes(manga)
        ? base.mangas.filter((m) => m !== manga)
        : [...base.mangas, manga]
      return { ...base, mangas }
    })
  }

  const handleAddConfigBan = (powerId: string) => {
    setConfigForm((current) => {
      const base = current ?? form
      if (base.poolFilter.banned.includes(powerId)) return base
      return { ...base, poolFilter: { ...base.poolFilter, banned: [...base.poolFilter.banned, powerId] } }
    })
  }

  const handleRemoveConfigBan = (powerId: string) => {
    setConfigForm((current) => {
      const base = current ?? form
      return {
        ...base,
        poolFilter: { ...base.poolFilter, banned: base.poolFilter.banned.filter((b) => b !== powerId) },
      }
    })
  }

  const handleClearConfigPoolFilter = () => {
    setConfigForm((current) => {
      const base = current ?? form
      return { ...base, poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] } }
    })
  }

  // Whitelisting by rarity/fruit-type isn't exposed in the UI (owner call -
  // only banning specific powers is needed), so the active-count badge only
  // ever reflects the banlist.
  const configPoolActiveCount = form.poolFilter.banned.length

  const handleSubmitConfig = () => {
    setConfigSaving(true)
    setConfigSaved(false)
    commands.updateConfig({
      mode: form.mode,
      mangas: form.mangas,
      abilitySource: snapshot.config.abilitySource,
      teamSize: form.teamSize,
      allowBots: form.allowBots,
      visibility: form.visibility,
      votingWindowSeconds: form.votingWindowSeconds,
      poolFilter: form.poolFilter,
    })
    setTimeout(() => {
      setConfigSaving(false)
      setConfigSaved(true)
    }, 500)
  }

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
      configMangas={form.mangas}
      onToggleConfigManga={handleToggleConfigManga}
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
      onChangeConfigVotingWindow={(votingWindowSeconds) => setConfigForm({ ...form, votingWindowSeconds })}
      configPoolFilter={form.poolFilter}
      configPoolActiveCount={configPoolActiveCount}
      configBanlistItems={banlistItems}
      onAddConfigBan={handleAddConfigBan}
      onRemoveConfigBan={handleRemoveConfigBan}
      onClearConfigPoolFilter={handleClearConfigPoolFilter}
      configSaving={configSaving}
      configSaved={configSaved}
      onSubmitConfig={handleSubmitConfig}
    />
  )
}
