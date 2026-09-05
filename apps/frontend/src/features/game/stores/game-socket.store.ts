import { create } from 'zustand'
import { z } from 'zod'

import { reconnectDelay } from '@/features/game/lib/backoff'
import { buildGameSocketUrl } from '@/features/game/lib/socket-url'
import type { ClientCommandType } from '@/shared/contracts/ws'
import { SERVER_FRAME, serverFrameSchema } from '@/shared/contracts/ws'
import type { GameResult, GameStateResponse } from '@/features/game/types/game.types'
import { useSessionStore } from '@/shared/stores/session.store'
import { mintGameSocketTicket } from '@/shared/api/stream-tickets'
import { toAppError } from '@/shared/api/errors'

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
  /** Epoch ms deadline for the in-flight sorteo - the frame's own
   * authoritative closesAt (LOADOUTS_ASSIGNED, stamped server-side from
   * GameService.revealEnds), or adopted from a STATE frame's
   * game.revealEndsAt for a client that missed that frame (a reconnect
   * mid-reveal). Also useLoadoutReveal's pacing input: the reveal's own
   * locally-computed timeline is scaled to fit whatever time remains until
   * this instant. null once voting has actually opened. */
  revealEndsAt: number | null
  /** Absolute values off the latest REVEAL_READY_CHANGED - how many of how
   * many connected humans have marked the current sorteo ready to skip
   * (see game.RevealReadyProgress). null until the first frame for this
   * ASSIGNING window arrives; reset to null every time a fresh assignment
   * starts (see the LOADOUTS_ASSIGNED case below). */
  revealReadyCount: number | null
  revealReadyTotal: number | null
  /** Epoch ms deadline for the in-flight loadout-summary screen (2026-08-30,
   * between the sorteo and the actual vote) - derived locally
   * (Date.now() + closesAt-implied ms) on SUMMARY_OPENED, or adopted from a
   * STATE frame's game.summaryEndsAt for a reconnect mid-summary. null once
   * voting has actually opened (or the round never reassigned and skipped
   * SUMMARY entirely). */
  summaryEndsAt: number | null
  /** Absolute values off the latest SUMMARY_READY_CHANGED, mirroring
   * revealReadyCount/revealReadyTotal exactly but for the summary screen's
   * own skip vote. */
  summaryReadyCount: number | null
  summaryReadyTotal: number | null
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
  /** Epoch ms deadline for the in-flight round-result display - the
   * frame's own authoritative closesAt (ROUND_RESOLVED, stamped
   * server-side from GameService.resultEnds), or adopted from a STATE
   * frame's game.resultEndsAt for a reconnect mid-result. null once the
   * display has ended (or hasn't started). */
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
  revealEndsAt: null,
  revealReadyCount: null,
  revealReadyTotal: null,
  summaryEndsAt: null,
  summaryReadyCount: null,
  summaryReadyTotal: null,
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
  /** The id of the new lobby a REMATCH created from this (finished/aborted)
   * game, or null. Set from the REMATCH_READY frame, which the server
   * publishes to EVERY client on the old game - so all of them navigate,
   * not just whoever pressed the button. The container reads this and
   * routes; nothing else consumes it. */
  rematchGameId: string | null

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
// connectGeneration guards the mint-ticket await connect() now has to make:
// bumped inside closeSocket() (attach switching games, detach, reset), so a
// mint that resolves after the caller has moved on can't resurrect a socket
// for the wrong game. pendingConnect prevents a second, concurrent connect()
// for the SAME game (attach() re-mounting while a mint is still in flight) -
// unlike the old synchronous connect(), an async one leaves socket === null
// for a while, and attach()'s `if (!socket) connect(gameId)` alone can no
// longer tell "already connecting" from "not connected".
let connectGeneration = 0
let pendingConnect = false

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
    // Invalidates any mint that's still in flight from a previous connect()
    // call - see connectGeneration's doc above.
    connectGeneration += 1
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

  // scheduleReconnect is shared by ws.onclose and connect's own mint-failure
  // catch, so a failed mint follows exactly the same backoff curve
  // (reconnectAttempts/nextRetryAt/retryNow) a dropped socket always has.
  const scheduleReconnect = () => {
    const attempt = get().reconnectAttempts
    const delay = reconnectDelay(attempt)
    set({
      status: 'reconnecting',
      reconnectAttempts: attempt + 1,
      nextRetryAt: Date.now() + delay,
    })
    reconnectTimer = setTimeout(() => {
      const id = get().gameId
      if (id && !pendingConnect) connect(id)
    }, delay)
  }

  const connect = async (gameId: string) => {
    const token = useSessionStore.getState().session?.accessToken
    if (!token) return
    // Checked before minting (a blank ticket never reaches the URL builder
    // either way) so an unconfigured EXPO_PUBLIC_SOCKET_URL never burns a
    // ticket for nothing.
    if (!buildGameSocketUrl(gameId, '')) {
      set({ status: 'unavailable' })
      return
    }

    const gen = ++connectGeneration
    pendingConnect = true
    set({ status: hasConnectedBefore ? 'reconnecting' : 'connecting' })

    let ticket: string
    try {
      ticket = await mintGameSocketTicket(gameId)
    } catch (err) {
      pendingConnect = false
      // Superseded by a detach/attach-to-another-game while the mint was in
      // flight - the caller that superseded us already owns status/socket.
      if (gen !== connectGeneration) return

      const appErr = toAppError(err)
      if (appErr.status === 401) {
        // The response interceptor already tried one silent refresh-and-
        // retry (shared/api/interceptors.ts) before this error reached here
        // - a 401 surfacing at this layer means that refresh attempt itself
        // failed, so the interceptor has already cleared the session.
        // Nothing left to retry here; the session-gated screens handle it.
        set({ status: 'closed' })
        return
      }
      if (appErr.status === 403 || appErr.status === 404) {
        // Not seated in this game (or it's gone). Terminal: before tickets,
        // this same rejection only surfaced as a WS handshake failure the
        // browser can't inspect, and the store retried it forever.
        set({ status: 'closed', lastError: { message: appErr.message, code: appErr.code } })
        return
      }
      scheduleReconnect()
      return
    }
    pendingConnect = false
    if (gen !== connectGeneration) return

    const url = buildGameSocketUrl(gameId, ticket)
    if (!url) {
      set({ status: 'unavailable' })
      return
    }

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
      let raw: unknown
      try {
        raw = JSON.parse(event.data as string)
      } catch {
        return
      }

      // Validated against the generated discriminated union instead of a
      // blind cast: a backend field ADDED to a payload is safe (zod v4
      // strips unknown keys by default), a field removed/renamed fails the
      // parse for that one frame - handled the same way an already-unknown
      // frame type always has, below (a feed entry, nothing else touched;
      // STATE additionally means the previous snapshot is simply never
      // replaced, since the branch below never runs).
      const parsed = serverFrameSchema.safeParse(raw)
      if (!parsed.success) {
        if (__DEV__) {
          console.error('[game-socket] dropped an unparseable frame:', z.prettifyError(parsed.error))
        }
        const rawType = (raw as { type?: unknown } | null)?.type
        pushFeed(typeof rawType === 'string' ? rawType : 'UNKNOWN')
        return
      }
      const frame = parsed.data

      switch (frame.type) {
        case SERVER_FRAME.STATE: {
          const payload = frame.payload
          set((state) => {
            let live = state.live

            // Adopt a snapshot's own revealEndsAt only when we aren't
            // already tracking one locally - a genuine reconnect mid-sorteo
            // (missed the LOADOUTS_ASSIGNED frame entirely), not the normal
            // STATE resend that follows every frame we DID see.
            if (live.revealEndsAt === null && payload.game.revealEndsAt) {
              const revealEndsAt = Date.parse(payload.game.revealEndsAt) || null
              if (revealEndsAt !== null) {
                live = { ...live, revealEndsAt }
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
            // Same shape and reasoning again: ROUND_RESOLVED now carries its
            // own closesAt (set on the live path below), so this guard only
            // ever fires for a reconnect mid-RESOLVING - the STATE that
            // follows RESYNC when no ROUND_RESOLVED frame was seen.
            if (live.resultEndsAt === null && payload.game.resultEndsAt) {
              const resultEndsAt = Date.parse(payload.game.resultEndsAt) || null
              if (resultEndsAt !== null) {
                live = { ...live, resultEndsAt }
              }
            }
            // Same shape and reasoning again: a reconnect mid-SUMMARY
            // adopts the deadline from this STATE frame, same as
            // revealEndsAt/votingClosesAt/resultEndsAt above.
            if (live.summaryEndsAt === null && payload.game.summaryEndsAt) {
              const summaryEndsAt = Date.parse(payload.game.summaryEndsAt) || null
              if (summaryEndsAt !== null) {
                live = { ...live, summaryEndsAt }
              }
            }

            return live === state.live ? { snapshot: payload } : { snapshot: payload, live }
          })
          break
        }
        case SERVER_FRAME.PLAYER_KICKED: {
          const payload = frame.payload
          const self = get().snapshot?.you.participantId
          if (payload.participantId && payload.participantId === self) {
            set({ terminal: { kind: 'KICKED' } })
          } else {
            pushFeed(frame.type)
          }
          break
        }
        case SERVER_FRAME.GAME_FINISHED: {
          const payload = frame.payload
          set({ terminal: { kind: 'FINISHED', result: payload.result } })
          break
        }
        case SERVER_FRAME.GAME_ABORTED: {
          const payload = frame.payload
          set({ terminal: { kind: 'ABORTED', reason: payload.reason ?? '' } })
          break
        }
        case SERVER_FRAME.REMATCH_READY: {
          // Published on THIS (old) game's stream by the server, to every
          // connected client - so the whole roster follows the host over to
          // the new lobby instead of only the person who pressed Rematch.
          const payload = frame.payload
          if (payload.gameId) set({ rematchGameId: payload.gameId })
          break
        }
        case SERVER_FRAME.ERROR: {
          const payload = frame.payload
          set({
            lastError: {
              message: payload.error ?? 'Unknown error',
              code: payload.code,
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
          const payload = frame.payload
          set((state) => {
            if (payload.roundIndex !== state.live.votingRoundIndex) {
              return {}
            }
            return {
              live: {
                ...state.live,
                votesCast: payload.votesCast,
                voters: payload.voters,
              },
            }
          })
          break
        }
        case SERVER_FRAME.REVEAL_READY_CHANGED: {
          // Absolute values, never an increment - same shape as VOTE_CAST.
          const payload = frame.payload
          set((state) => ({
            live: {
              ...state.live,
              revealReadyCount: payload.ready,
              revealReadyTotal: payload.total,
            },
          }))
          break
        }
        case SERVER_FRAME.SUMMARY_OPENED: {
          const payload = frame.payload
          set((state) => ({
            live: {
              ...state.live,
              summaryEndsAt: Date.parse(payload.closesAt) || null,
              // Fresh SUMMARY window, fresh ready set - mirrors
              // Game.OpenSummary resetting summaryReady server-side.
              summaryReadyCount: null,
              summaryReadyTotal: null,
            },
          }))
          break
        }
        case SERVER_FRAME.SUMMARY_READY_CHANGED: {
          // Absolute values, never an increment - same shape as VOTE_CAST/
          // REVEAL_READY_CHANGED.
          const payload = frame.payload
          set((state) => ({
            live: {
              ...state.live,
              summaryReadyCount: payload.ready,
              summaryReadyTotal: payload.total,
            },
          }))
          break
        }
        case SERVER_FRAME.LOADOUTS_ASSIGNED: {
          // Sent BEFORE its own pushCurrentState, so `snapshot` here may
          // still be pre-assignment - never touch snapshot/feed from this
          // case. hasAllLoadouts/currentRound (match-rules.ts) gate the
          // actual reveal on the snapshot catching up.
          const payload = frame.payload
          set((state) => ({
            live: {
              ...state.live,
              assignmentSeq: state.live.assignmentSeq + 1,
              assignedRoundIndex: payload.roundIndex,
              revealEndsAt: Date.parse(payload.closesAt) || null,
              // Fresh ASSIGNING window, fresh ready set - mirrors
              // Game.AssignLoadouts resetting revealReady server-side.
              revealReadyCount: null,
              revealReadyTotal: null,
            },
          }))
          break
        }
        case SERVER_FRAME.VOTING_OPENED: {
          const payload = frame.payload
          set((state) => ({
            live: {
              ...state.live,
              votingRoundIndex: payload.roundIndex,
              votingClosesAt: Date.parse(payload.closesAt) || null,
              tiebreak: false,
              revealEndsAt: null,
              revealReadyCount: null,
              revealReadyTotal: null,
              summaryEndsAt: null,
              summaryReadyCount: null,
              summaryReadyTotal: null,
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
          const payload = frame.payload
          set((state) => ({
            live: {
              ...state.live,
              votingRoundIndex: payload.roundIndex,
              votingClosesAt: Date.parse(payload.closesAt) || null,
              tiebreak: true,
              revealEndsAt: null,
              summaryEndsAt: null,
              summaryReadyCount: null,
              summaryReadyTotal: null,
              votesCast: 0,
              voters: null,
              resultEndsAt: null,
              resultDismissed: false,
            },
          }))
          break
        }
        case SERVER_FRAME.ROUND_RESOLVED: {
          const payload = frame.payload
          set((state) => ({
            live: {
              ...state.live,
              votingRoundIndex: null,
              votingClosesAt: null,
              votesCast: null,
              voters: null,
              // Authoritative now - stamped server-side from
              // GameService.resultEnds - so the STATE-adoption guard above
              // correctly declines to overwrite it; it only ever fills in
              // for a reconnect that missed this frame entirely.
              resultEndsAt: Date.parse(payload.closesAt) || null,
              resultDismissed: false,
            },
          }))
          pushFeed(frame.type)
          break
        }
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
      scheduleReconnect()
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
    rematchGameId: null,

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
          rematchGameId: null,
        })
      }
      refCount += 1
      if (!socket && !pendingConnect) connect(gameId)
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
      if (id && !pendingConnect) connect(id)
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
        rematchGameId: null,
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
