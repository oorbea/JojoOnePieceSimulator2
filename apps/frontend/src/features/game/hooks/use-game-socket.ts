import { useEffect } from 'react'
import { AppState } from 'react-native'

import { useGameSocketStore } from '@/features/game/stores/game-socket.store'

// Binds the module-level socket store to this component's lifecycle: attach
// on mount, detach on unmount (refcounted, so a second room mount for the
// same gameId reuses the open socket instead of reconnecting). Also retries
// immediately when the app returns to the foreground while not connected -
// a backgrounded mobile socket can look alive to the OS long after the
// server has given up on it.
export function useGameSocket(gameId: string | null) {
  const status = useGameSocketStore((s) => s.status)
  const snapshot = useGameSocketStore((s) => s.snapshot)
  const terminal = useGameSocketStore((s) => s.terminal)
  const lastError = useGameSocketStore((s) => s.lastError)
  const nextRetryAt = useGameSocketStore((s) => s.nextRetryAt)
  const live = useGameSocketStore((s) => s.live)
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
      if (state === 'active' && status !== 'open' && gameId) retryNow()
    })
    return () => sub.remove()
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
    send,
    retryNow,
    markAssignmentRevealed,
    dismissResult,
  }
}
