import type { GameStateResponse, GameResult } from '@/features/game/types/game.types'

// Mirrors dto.game_ws.go's client command constants. Create/join/get are
// deliberately not commands (plain HTTP), and Disconnect/Reconnect are
// socket-lifecycle only (driven by open/close, never sent by a client).
export const CLIENT_COMMAND = {
  LEAVE: 'LEAVE',
  ADD_BOT: 'ADD_BOT',
  REMOVE_BOT: 'REMOVE_BOT',
  START: 'START',
  ABORT: 'ABORT',
  VOTE: 'VOTE',
  RESYNC: 'RESYNC',
  SWITCH_TEAM: 'SWITCH_TEAM',
  MOVE_PLAYER: 'MOVE_PLAYER',
  KICK: 'KICK',
  TRANSFER_HOST: 'TRANSFER_HOST',
  SET_LOCK: 'SET_LOCK',
  UPDATE_CONFIG: 'UPDATE_CONFIG',
  // REVEAL_READY is the sorteo's own skip vote (owner decision,
  // 2026-08-30) - no payload, always applies to the caller.
  REVEAL_READY: 'REVEAL_READY',
  // SUMMARY_READY is the loadout-summary screen's own skip vote (owner
  // decision, 2026-08-30), mirroring REVEAL_READY exactly - no payload,
  // always applies to the caller.
  SUMMARY_READY: 'SUMMARY_READY',
  // REMATCH asks for a brand new lobby carrying this finished/aborted
  // game's whole roster over. Host-only, terminal-state-only, no payload -
  // the new game's id comes back to every connected client as a
  // REMATCH_READY frame, not as a reply to the requester alone.
  REMATCH: 'REMATCH',
} as const

export type ClientCommandType = (typeof CLIENT_COMMAND)[keyof typeof CLIENT_COMMAND]

export type ClientCommand = {
  type: ClientCommandType
  requestId?: string
  payload?: Record<string, unknown>
}

// Mirrors dto.game_ws.go's server frame type constants, reusing the
// domain-event names verbatim (see entities/game/events.go).
export const SERVER_FRAME = {
  STATE: 'STATE',
  PLAYER_JOINED: 'PLAYER_JOINED',
  PLAYER_LEFT: 'PLAYER_LEFT',
  HOST_REASSIGNED: 'HOST_REASSIGNED',
  GAME_STARTED: 'GAME_STARTED',
  LOADOUTS_ASSIGNED: 'LOADOUTS_ASSIGNED',
  VOTING_OPENED: 'VOTING_OPENED',
  VOTE_CAST: 'VOTE_CAST',
  TIEBREAK_OPENED: 'TIEBREAK_OPENED',
  ROUND_RESOLVED: 'ROUND_RESOLVED',
  GAME_FINISHED: 'GAME_FINISHED',
  GAME_ABORTED: 'GAME_ABORTED',
  ERROR: 'ERROR',
  RESYNC_REQUIRED: 'RESYNC_REQUIRED',
  TEAM_CHANGED: 'TEAM_CHANGED',
  PLAYER_KICKED: 'PLAYER_KICKED',
  LOBBY_LOCK_CHANGED: 'LOBBY_LOCK_CHANGED',
  CONFIG_UPDATED: 'CONFIG_UPDATED',
  REVEAL_READY_CHANGED: 'REVEAL_READY_CHANGED',
  SUMMARY_OPENED: 'SUMMARY_OPENED',
  SUMMARY_READY_CHANGED: 'SUMMARY_READY_CHANGED',
  REMATCH_READY: 'REMATCH_READY',
} as const

export type ServerFrameType = (typeof SERVER_FRAME)[keyof typeof SERVER_FRAME]

export type ServerFrame = {
  type: string
  requestId?: string
  payload?: unknown
}

export type StatePayload = GameStateResponse
export type ErrorPayload = { error: string; code?: string; details?: string[] }
export type PlayerKickedPayload = { participantId: string }
export type VotingOpenedPayload = { roundIndex: number; closesAt: string }
export type VoteCastPayload = { roundIndex: number; votesCast: number; voters: number }
export type LoadoutsAssignedPayload = { roundIndex: number; revealMs: number }
export type TiebreakOpenedPayload = { roundIndex: number; closesAt: string }
export type RoundResolvedPayload = { roundIndex: number; winner: string; decidedByCoinFlip: boolean }
export type GameFinishedPayload = { result: GameResult }
export type GameAbortedPayload = { reason: string }
export type ResyncRequiredPayload = { reason: string }
// Published on the OLD (finished/aborted) game's stream - gameId is the NEW
// lobby to navigate to, so every client on the result screen follows the
// roster over, not just whoever pressed Rematch.
export type RematchReadyPayload = { gameId: string }
