import { useEffect } from 'react'
import { AppState, Platform } from 'react-native'

import { foregroundAction, onlineAction } from '@/features/game/lib/reconnect-policy'
import { useGameSocketStore } from '@/features/game/stores/game-socket.store'
import type { ClientCommandType } from '@/shared/contracts/ws'

// Binds the module-level socket store to this component's lifecycle: attach
// on mount, detach on unmount (refcounted, so a second room mount for the
// same gameId reuses the open socket instead of reconnecting). Also reacts
// to the device coming back (foreground / network): retries a dead socket
// immediately and resyncs one that looks open - a backgrounded mobile socket
// can look alive to the OS long after the server gave up on it or while
// frames were missed. See ../lib/reconnect-policy.ts.
export function useGameSocket(gameId: string | null) {
  const status = useGameSocketStore((s) => s.status)
  const snapshot = useGameSocketStore((s) => s.snapshot)
  const terminal = useGameSocketStore((s) => s.terminal)
  const lastError = useGameSocketStore((s) => s.lastError)
  const nextRetryAt = useGameSocketStore((s) => s.nextRetryAt)
  const live = useGameSocketStore((s) => s.live)
  const rematchGameId = useGameSocketStore((s) => s.rematchGameId)
  const send = useGameSocketStore((s) => s.send)
  const retryNow = useGameSocketStore((s) => s.retryNow)
  const attach = useGameSocketStore((s) => s.attach)
  const detach = useGameSocketStore((s) => s.detach)
  const markAssignmentRevealed = useGameSocketStore((s) => s.markAssignmentRevealed)
  const dismissResult = useGameSocketStore((s) => s.dismissResult)

  useEffect(() => {
    if (!gameId) return
    attach(gameId)
    return () => detach()
    // eslint-disable-next-line react-hooks/exhaustive-deps -- attach/detach are stable store actions
  }, [gameId])

  useEffect(() => {
    const sub = AppState.addEventListener('change', (state) => {
      if (state !== 'active' || !gameId) return
      if (foregroundAction(status) === 'resync') send('RESYNC' as ClientCommandType)
      else retryNow()
    })
    return () => sub.remove()
  }, [status, gameId, retryNow, send])

  // AppState only covers visibility; a connection lost while the app stays
  // in front (tunnel, lift, airplane mode toggled off) is signalled by the
  // browser's own online event, which beats waiting out the backoff.
  useEffect(() => {
    if (Platform.OS !== 'web' || !gameId) return
    const onOnline = () => {
      if (onlineAction(status) === 'retry') retryNow()
    }
    window.addEventListener('online', onOnline)
    return () => window.removeEventListener('online', onOnline)
  }, [status, gameId, retryNow])

  return {
    status,
    snapshot: snapshot?.game ?? null,
    you: snapshot?.you ?? null,
    isHost: snapshot?.you.isHost ?? false,
    terminal,
    lastError,
    nextRetryAt,
    live,
    rematchGameId,
    send,
    retryNow,
    markAssignmentRevealed,
    dismissResult,
  }
}
