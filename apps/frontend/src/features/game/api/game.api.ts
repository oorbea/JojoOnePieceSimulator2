import { apiClient } from '@/shared/api/client'
import { assertContract } from '@/shared/api/assert-contract'
import { gameStateResponseSchema } from '@/shared/contracts/dto'
import type {
  CreateGameInput,
  GameInvite,
  GameStateResponse,
  InvitePreview,
  InviteStatusResult,
  LobbyPreview,
  PublicLobbyList,
} from '@/features/game/types/game.types'

export async function createGame(input: CreateGameInput): Promise<GameStateResponse> {
  const response = await apiClient.post<GameStateResponse>('/games', input)
  return response.data
}

export async function joinGameByCode(code: string): Promise<GameStateResponse> {
  const response = await apiClient.post<GameStateResponse>('/games/join', { code })
  return response.data
}

export async function joinGameById(gameId: string): Promise<GameStateResponse> {
  const response = await apiClient.post<GameStateResponse>(`/games/${gameId}/join`)
  return response.data
}

export async function getGame(gameId: string): Promise<GameStateResponse> {
  const response = await apiClient.get<GameStateResponse>(`/games/${gameId}`)
  // Dev-only: proves the REST snapshot still matches the generated
  // contract, same shape the WS STATE frame already validates for real -
  // see assert-contract.ts's doc for why this never runs in production.
  if (__DEV__) assertContract(gameStateResponseSchema, response.data, 'GET /games/:id')
  return response.data
}

export async function previewGameByCode(code: string): Promise<LobbyPreview> {
  const response = await apiClient.get<LobbyPreview>('/games/preview', { params: { code } })
  return response.data
}

export async function getPublicLobbies(): Promise<PublicLobbyList> {
  const response = await apiClient.get<PublicLobbyList>('/games/public')
  return response.data
}

// getMyGame resumes the caller's active game (see GET /games/me's own doc):
// null means there is none to resume, not an error - the client's own
// validateStatus already treats 204 as success, so callers never need to
// retry it as if it failed.
export async function getMyGame(): Promise<GameStateResponse | null> {
  const response = await apiClient.get<GameStateResponse | ''>('/games/me')
  if (response.status === 204 || !response.data) return null
  if (__DEV__) assertContract(gameStateResponseSchema, response.data, 'GET /games/me')
  return response.data
}

// createGameInvite mints a fresh share-link invite token for gameId. Any
// seated human participant may call this, not only the host.
export async function createGameInvite(gameId: string): Promise<GameInvite> {
  const response = await apiClient.post<GameInvite>(`/games/${gameId}/invite`)
  return response.data
}

// getInviteStatus hits the PUBLIC status route - no bearer token required,
// safe to call before the visitor has signed in (or hasn't at all).
export async function getInviteStatus(token: string): Promise<InviteStatusResult> {
  const response = await apiClient.get<InviteStatusResult>(
    `/games/invite/${encodeURIComponent(token)}/status`
  )
  return response.data
}

// getInvitePreview is the authenticated counterpart of previewGameByCode,
// reachable with a token instead of a join code - works for PRIVATE
// lobbies too.
export async function getInvitePreview(token: string): Promise<InvitePreview> {
  const response = await apiClient.get<InvitePreview>(`/games/invite/${encodeURIComponent(token)}`)
  return response.data
}

export async function joinGameByInvite(token: string): Promise<GameStateResponse> {
  const response = await apiClient.post<GameStateResponse>('/games/join-invite', { token })
  return response.data
}

// leaveGameById backs the invite-link "leave your current game and join
// this one" confirm - there's no open WS socket to the game being left at
// that point, so this REST route (mirroring the WS LEAVE command) exists
// specifically for it.
export async function leaveGameById(gameId: string): Promise<void> {
  await apiClient.post(`/games/${gameId}/leave`)
}
