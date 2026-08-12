import { create } from 'zustand'

import { reconnectDelay } from '@/features/game/lib/backoff'
import { buildGameSocketUrl } from '@/features/game/lib/socket-url'
import type { ClientCommandType } from '@/features/game/types/game-ws.types'
import { SERVER_FRAME } from '@/features/game/types/game-ws.types'
import type { GameResult, GameStateResponse } from '@/features/game/types/game.types'
import { useSessionStore } from '@/shared/stores/session.store'

export type SocketStatus = 'idle' | 'connecting' | 'open' | 'reconnecting' | 'closed' | 'unavailable'

export type TerminalInfo =
  | { kind: 'FINISHED'; result: GameResult }
  | { kind: 'ABORTED'; reason: string }
  | { kind: 'KICKED' }

type FeedEntry = { id: number; type: string; at: number }

type LastError = { code?: string; message: string; requestId?: string }

type SocketFactory = (url: string) => WebSocket

type GameSocketState = {
  gameId: string | null
  status: SocketStatus
  snapshot: GameStateResponse | null
  terminal: TerminalInfo | null
  lastError: LastError | null
  feed: FeedEntry[]
  reconnectAttempts: number
  nextRetryAt: number | null

  attach: (gameId: string, socketFactory?: SocketFactory) => void
  detach: () => void
  send: (type: ClientCommandType, payload?: Record<string, unknown>) => string
  retryNow: () => void
  reset: () => void
}

// Module-level (not zustand state, deliberately): a live WebSocket isn't
// serializable state, and only one connection should ever exist regardless
// of how many components render the room. Refcounted the same way the
// backend's own connRegistry is (game_ws_endpoints.go) - a second mount for
// the same gameId reuses the socket instead of opening a new one.
let socket: WebSocket | null = null
let socketFactoryRef: SocketFactory = (url) => new WebSocket(url)
let refCount = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let hasConnectedBefore = false
let feedCounter = 0

function nextFeedId() {
  feedCounter += 1
  return feedCounter
}

function genRequestId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

export const useGameSocketStore = create<GameSocketState>((set, get) => {
  const clearReconnectTimer = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  const closeSocket = () => {
    clearReconnectTimer()
    if (socket) {
      const s = socket
      socket = null
      s.onopen = null
      s.onmessage = null
      s.onerror = null
      s.onclose = null
      s.close()
    }
  }

  const pushFeed = (type: string) => {
    const feed = [...get().feed, { id: nextFeedId(), type, at: Date.now() }]
    set({ feed: feed.slice(-8) })
  }

  const connect = (gameId: string) => {
    const token = useSessionStore.getState().session?.accessToken
    if (!token) return
    const url = buildGameSocketUrl(gameId, token)
    if (!url) {
      set({ status: 'unavailable' })
      return
    }

    set({ status: hasConnectedBefore ? 'reconnecting' : 'connecting' })
    const ws = socketFactoryRef(url)
    socket = ws

    ws.onopen = () => {
      if (socket !== ws) return
      set({ status: 'open', reconnectAttempts: 0, nextRetryAt: null })
      if (hasConnectedBefore) {
        get().send('RESYNC' as ClientCommandType)
      }
      hasConnectedBefore = true
    }

    ws.onmessage = (event: MessageEvent) => {
      if (socket !== ws) return
      let frame: { type?: string; requestId?: string; payload?: unknown }
      try {
        frame = JSON.parse(event.data as string)
      } catch {
        return
      }
      if (!frame || typeof frame.type !== 'string') return

      switch (frame.type) {
        case SERVER_FRAME.STATE:
          set({ snapshot: frame.payload as GameStateResponse })
          break
        case SERVER_FRAME.PLAYER_KICKED: {
          const payload = frame.payload as { participantId?: string } | undefined
          const self = get().snapshot?.you.participantId
          if (payload?.participantId && payload.participantId === self) {
            set({ terminal: { kind: 'KICKED' } })
          } else {
            pushFeed(frame.type)
          }
          break
        }
        case SERVER_FRAME.GAME_FINISHED: {
          const payload = frame.payload as { result?: GameResult } | undefined
          if (payload?.result) set({ terminal: { kind: 'FINISHED', result: payload.result } })
          break
        }
        case SERVER_FRAME.GAME_ABORTED: {
          const payload = frame.payload as { reason?: string } | undefined
          set({ terminal: { kind: 'ABORTED', reason: payload?.reason ?? '' } })
          break
        }
        case SERVER_FRAME.ERROR: {
          const payload = frame.payload as { error?: string; code?: string } | undefined
          set({
            lastError: {
              message: payload?.error ?? 'Unknown error',
              code: payload?.code,
              requestId: frame.requestId,
            },
          })
          break
        }
        case SERVER_FRAME.RESYNC_REQUIRED:
          get().send('RESYNC' as ClientCommandType)
          break
        case SERVER_FRAME.VOTE_CAST:
          // High-frequency, self-describing, no snapshot impact - the
          // in-match tanda will render live.votesCast from here.
          break
        default:
          pushFeed(frame.type)
      }
    }

    ws.onerror = () => {
      // onclose always follows onerror for a WebSocket - the reconnect
      // logic lives there so it only ever runs once per failure.
    }

    ws.onclose = () => {
      if (socket !== ws) return
      socket = null
      if (get().terminal) {
        set({ status: 'closed' })
        return
      }
      const attempt = get().reconnectAttempts
      const delay = reconnectDelay(attempt)
      set({ status: 'reconnecting', reconnectAttempts: attempt + 1, nextRetryAt: Date.now() + delay })
      reconnectTimer = setTimeout(() => {
        const id = get().gameId
        if (id) connect(id)
      }, delay)
    }
  }

  return {
    gameId: null,
    status: 'idle',
    snapshot: null,
    terminal: null,
    lastError: null,
    feed: [],
    reconnectAttempts: 0,
    nextRetryAt: null,

    attach: (gameId, socketFactory) => {
      if (socketFactory) socketFactoryRef = socketFactory
      const current = get()
      if (current.gameId !== gameId) {
        closeSocket()
        hasConnectedBefore = false
        refCount = 0
        set({
          gameId,
          status: 'idle',
          snapshot: null,
          terminal: null,
          lastError: null,
          feed: [],
          reconnectAttempts: 0,
          nextRetryAt: null,
        })
      }
      refCount += 1
      if (!socket) connect(gameId)
    },

    detach: () => {
      refCount = Math.max(0, refCount - 1)
      if (refCount === 0) {
        closeSocket()
      }
    },

    send: (type, payload) => {
      const requestId = genRequestId()
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type, requestId, payload }))
      }
      return requestId
    },

    retryNow: () => {
      clearReconnectTimer()
      set({ reconnectAttempts: 0 })
      const id = get().gameId
      if (id) connect(id)
    },

    reset: () => {
      closeSocket()
      hasConnectedBefore = false
      refCount = 0
      set({
        gameId: null,
        status: 'idle',
        snapshot: null,
        terminal: null,
        lastError: null,
        feed: [],
        reconnectAttempts: 0,
        nextRetryAt: null,
      })
    },
  }
})
