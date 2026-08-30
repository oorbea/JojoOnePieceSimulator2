import { create } from 'zustand'

import { reconnectDelay } from '@/features/game/lib/backoff'
import { buildGameSocketUrl } from '@/features/game/lib/socket-url'
import type { ClientCommandType } from '@/features/game/types/game-ws.types'
import { SERVER_FRAME } from '@/features/game/types/game-ws.types'
import type { GameResult, GameStateResponse } from '@/features/game/types/game.types'
import { useSessionStore } from '@/shared/stores/session.store'

export type SocketStatus =
  'idle' | 'connecting' | 'open' | 'reconnecting' | 'closed' | 'unavailable'

export type TerminalInfo =
  | { kind: 'FINISHED'; result: GameResult }
  | { kind: 'ABORTED'; reason: string }
  | { kind: 'KICKED' }

type FeedEntry = { id: number; type: string; at: number }

type LastError = { code?: string; message: string; requestId?: string }

// In-match live signal: not part of `snapshot` (which is replaced wholesale
// on STATE and never touched anywhere else) - LOADOUTS_ASSIGNED/
// VOTING_OPENED/TIEBREAK_OPENED/ROUND_RESOLVED arrive as their own frames,
// often before their own pushCurrentState, so they're tracked separately and
// the reveal gating (features/game/lib/loadout-reveal.ts) cross-checks both.
export type LiveMatchState = {
  assignmentSeq: number
  revealedAssignmentSeq: number
  assignedRoundIndex: number | null
  /** The sorteo's own duration for this assignment, straight off
   * LOADOUTS_ASSIGNED's revealMs - authoritative pacing input for
   * useLoadoutReveal (see its doc). null until that frame arrives. */
  revealMs: number | null
  /** Epoch ms deadline for the in-flight sorteo - derived locally
   * (Date.now() + revealMs) on LOADOUTS_ASSIGNED, or adopted from a STATE
   * frame's game.revealEndsAt for a client that missed that frame (a
   * reconnect mid-reveal). null once voting has actually opened. */
  revealEndsAt: number | null
  /** Absolute values off the latest REVEAL_READY_CHANGED - how many of how
   * many connected humans have marked the current sorteo ready to skip
   * (see game.RevealReadyProgress). null until the first frame for this
   * ASSIGNING window arrives; reset to null every time a fresh assignment
   * starts (see the LOADOUTS_ASSIGNED case below). */
  revealReadyCount: number | null
  revealReadyTotal: number | null
  votingRoundIndex: number | null
  votingClosesAt: number | null
  tiebreak: boolean
  /** Absolute values off the latest VOTE_CAST for votingRoundIndex, never
   * an increment - bots cast the instant a window opens, so several frames
   * can read 0/N in a row. null until the first frame for this round
   * arrives (see lib/match-rules.ts's voteProgress for the snapshot-derived
   * fallback that covers that gap, e.g. right after a reconnect). */
  votesCast: number | null
  voters: number | null
  /** Epoch ms deadline for the in-flight round-result display - adopted
   * from a STATE frame's game.resultEndsAt (ROUND_RESOLVED itself carries
   * no closesAt, so this always comes from the STATE that follows it, same
   * as a reconnect mid-result). null once the display has ended (or hasn't
   * started). */
  resultEndsAt: number | null
  /** True once the local viewer has clicked "skip" on the current round's
   * result panel - reset on every VOTING_OPENED/TIEBREAK_OPENED/
   * ROUND_RESOLVED so the next round's panel starts visible again. Purely
   * a client-side convenience: the server keeps holding RESOLVING for its
   * own full duration regardless. */
  resultDismissed: boolean
}

const INITIAL_LIVE: LiveMatchState = {
  assignmentSeq: 0,
  revealedAssignmentSeq: 0,
  assignedRoundIndex: null,
  revealMs: null,
  revealEndsAt: null,
  revealReadyCount: null,
  revealReadyTotal: null,
  votingRoundIndex: null,
  votingClosesAt: null,
  tiebreak: false,
  votesCast: null,
  voters: null,
  resultEndsAt: null,
  resultDismissed: false,
}

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
  live: LiveMatchState

  attach: (gameId: string, socketFactory?: SocketFactory) => void
  detach: () => void
  send: (type: ClientCommandType, payload?: Record<string, unknown>) => string
  retryNow: () => void
  reset: () => void
  markAssignmentRevealed: () => void
  dismissResult: () => void
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
        case SERVER_FRAME.STATE: {
          const payload = frame.payload as GameStateResponse
          set((state) => {
            let live = state.live

            // Adopt a snapshot's own revealEndsAt only when we aren't
            // already tracking one locally - a genuine reconnect mid-sorteo
            // (missed the LOADOUTS_ASSIGNED frame entirely), not the normal
            // STATE resend that follows every frame we DID see.
            if (live.revealEndsAt === null && payload.game.revealEndsAt) {
              const revealEndsAt = Date.parse(payload.game.revealEndsAt) || null
              if (revealEndsAt !== null) {
                live = { ...live, revealEndsAt, revealMs: Math.max(0, revealEndsAt - Date.now()) }
              }
            }
            // Same shape and reasoning as revealEndsAt above: only adopt a
            // reconnecting client's votingEndsAt when nothing local is
            // already tracking it, so a live deadline is never overwritten
            // by a slightly-later STATE resend.
            if (live.votingClosesAt === null && payload.game.votingEndsAt) {
              const votingClosesAt = Date.parse(payload.game.votingEndsAt) || null
              if (votingClosesAt !== null) {
                live = { ...live, votingClosesAt }
              }
            }
            // Same shape and reasoning again: ROUND_RESOLVED carries no
            // closesAt of its own, so resultEndsAt always arrives via the
            // STATE that follows it (or, for a reconnect mid-RESOLVING, the
            // STATE that follows RESYNC).
            if (live.resultEndsAt === null && payload.game.resultEndsAt) {
              const resultEndsAt = Date.parse(payload.game.resultEndsAt) || null
              if (resultEndsAt !== null) {
                live = { ...live, resultEndsAt }
              }
            }

            return live === state.live ? { snapshot: payload } : { snapshot: payload, live }
          })
          break
        }
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
        case SERVER_FRAME.VOTE_CAST: {
          // Absolute values, never an increment - see LiveMatchState's doc.
          // Ignored when it's for a round we're no longer voting on (a late
          // frame arriving after ROUND_RESOLVED already cleared
          // votingRoundIndex, or for a stale previous round).
          const payload = frame.payload as
            { roundIndex?: number; votesCast?: number; voters?: number } | undefined
          set((state) => {
            if (
              payload?.roundIndex === undefined ||
              payload.roundIndex !== state.live.votingRoundIndex
            ) {
              return {}
            }
            return {
              live: {
                ...state.live,
                votesCast: payload.votesCast ?? null,
                voters: payload.voters ?? null,
              },
            }
          })
          break
        }
        case SERVER_FRAME.REVEAL_READY_CHANGED: {
          // Absolute values, never an increment - same shape as VOTE_CAST.
          const payload = frame.payload as { ready?: number; total?: number } | undefined
          set((state) => ({
            live: {
              ...state.live,
              revealReadyCount: payload?.ready ?? null,
              revealReadyTotal: payload?.total ?? null,
            },
          }))
          break
        }
        case SERVER_FRAME.LOADOUTS_ASSIGNED: {
          // Sent BEFORE its own pushCurrentState, so `snapshot` here may
          // still be pre-assignment - never touch snapshot/feed from this
          // case. hasAllLoadouts/currentRound (match-rules.ts) gate the
          // actual reveal on the snapshot catching up.
          const payload = frame.payload as { roundIndex?: number; revealMs?: number } | undefined
          const revealMs = payload?.revealMs ?? null
          set((state) => ({
            live: {
              ...state.live,
              assignmentSeq: state.live.assignmentSeq + 1,
              assignedRoundIndex: payload?.roundIndex ?? null,
              revealMs,
              revealEndsAt: revealMs !== null ? Date.now() + revealMs : null,
              // Fresh ASSIGNING window, fresh ready set - mirrors
              // Game.AssignLoadouts resetting revealReady server-side.
              revealReadyCount: null,
              revealReadyTotal: null,
            },
          }))
          break
        }
        case SERVER_FRAME.VOTING_OPENED: {
          const payload = frame.payload as { roundIndex?: number; closesAt?: string } | undefined
          set((state) => ({
            live: {
              ...state.live,
              votingRoundIndex: payload?.roundIndex ?? null,
              votingClosesAt: payload?.closesAt ? Date.parse(payload.closesAt) || null : null,
              tiebreak: false,
              revealMs: null,
              revealEndsAt: null,
              revealReadyCount: null,
              revealReadyTotal: null,
              // Reset, not cleared to null: the fresh window has cast=0 of
              // an as-yet-unknown total, covered by voteProgress's
              // snapshot-derived fallback until the first VOTE_CAST (or a
              // bot's, which fires immediately) lands.
              votesCast: 0,
              voters: null,
              resultEndsAt: null,
              resultDismissed: false,
            },
          }))
          break
        }
        case SERVER_FRAME.TIEBREAK_OPENED: {
          const payload = frame.payload as { roundIndex?: number; closesAt?: string } | undefined
          set((state) => ({
            live: {
              ...state.live,
              votingRoundIndex: payload?.roundIndex ?? null,
              votingClosesAt: payload?.closesAt ? Date.parse(payload.closesAt) || null : null,
              tiebreak: true,
              revealMs: null,
              revealEndsAt: null,
              votesCast: 0,
              voters: null,
              resultEndsAt: null,
              resultDismissed: false,
            },
          }))
          break
        }
        case SERVER_FRAME.ROUND_RESOLVED:
          set((state) => ({
            live: {
              ...state.live,
              votingRoundIndex: null,
              votingClosesAt: null,
              votesCast: null,
              voters: null,
              // resultEndsAt isn't in this payload (see LiveMatchState's
              // doc) - reset to null so the STATE frame that follows adopts
              // it fresh, same guard as revealEndsAt/votingClosesAt.
              resultEndsAt: null,
              resultDismissed: false,
            },
          }))
          pushFeed(frame.type)
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
      set({
        status: 'reconnecting',
        reconnectAttempts: attempt + 1,
        nextRetryAt: Date.now() + delay,
      })
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
    live: INITIAL_LIVE,

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
          live: INITIAL_LIVE,
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
        live: INITIAL_LIVE,
      })
    },

    markAssignmentRevealed: () => {
      set((state) => ({ live: { ...state.live, revealedAssignmentSeq: state.live.assignmentSeq } }))
    },

    dismissResult: () => {
      set((state) => ({ live: { ...state.live, resultDismissed: true } }))
    },
  }
})
