import { apiClient } from '@/shared/api/client'
import type {
  CreateGameInput,
  GameStateResponse,
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
