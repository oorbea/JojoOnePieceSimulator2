import { useRouter, useLocalSearchParams } from 'expo-router'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQueryClient } from '@tanstack/react-query'

import { LobbyRoomScreen, type ConfirmSheetState } from '@/features/game/components/presentational/lobby-room-screen'
import { gameKeys } from '@/features/game/api/game.keys'
import { useGameCommands } from '@/features/game/hooks/use-game-commands'
import { useGameDetail } from '@/features/game/hooks/use-game-detail'
import { useGameSocket } from '@/features/game/hooks/use-game-socket'
import { formatCode } from '@/features/game/lib/game-code'
import { startGate } from '@/features/game/lib/lobby-rules'
import { shareJoinCode } from '@/features/game/lib/share'
import { useGameSocketStore } from '@/features/game/stores/game-socket.store'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import { showErrorToast, showSuccessToast } from '@/shared/lib/toast'
import { AppError } from '@/shared/api/errors'

export function LobbyRoomContainer() {
  const router = useRouter()
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { id } = useLocalSearchParams<{ id: string }>()

  const socket = useGameSocket(id ?? null)
  const detail = useGameDetail(id ?? null, socket.status)
  const commands = useGameCommands()
  const resetSocket = useGameSocketStore((s) => s.reset)

  const [starting, setStarting] = useState(false)
  const [confirmSheet, setConfirmSheet] = useState<ConfirmSheetState>(null)

  const snapshot = socket.snapshot ?? detail.data?.game ?? null
  const you = socket.you ?? detail.data?.you ?? null

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
    />
  )
}
