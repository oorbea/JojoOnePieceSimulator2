import type {
  AbilitySource,
  GameMode,
  GameState,
  LobbyVisibility,
  Manga,
  ParticipantKind,
} from '@/shared/lib/zod'

// Mirrors dto.PoolFilterResponse (apps/backend .../api/dto/game_response.go).
// Empty arrays mean "no restriction", exactly like the domain type.
export type PoolFilter = {
  standRarities: string[]
  fruitRarities: string[]
  fruitTypes: string[]
  banned: string[]
}

// Mirrors dto.GameConfigResponse.
export type GameConfig = {
  mangas: Manga[]
  abilitySource: AbilitySource
  teamSize: number
  allowBots: boolean
  visibility: LobbyVisibility
  votingWindowSeconds: number
  poolFilter: PoolFilter
}

// Mirrors dto.GameTeamResponse.
export type GameTeam = {
  id: string
  name: string
  color: number
  memberIds: string[]
}

// Mirrors dto.GameLoadoutResponse. Kept minimal - the in-match round UI
// (loadout cards, vote timer) is the next tanda; the lobby only needs to
// know a loadout exists, not render it.
export type GameLoadout = Record<string, unknown>

// Mirrors dto.GameParticipantResponse.
export type GameParticipant = {
  id: string
  userId?: string
  displayName: string
  teamId: string
  kind: ParticipantKind
  connected: boolean
  loadout?: GameLoadout
}

// Mirrors dto.GameRoundResponse. Not rendered by the lobby yet - carried
// through so the socket store already has the shape ready for the
// in-match tanda.
export type GameRound = {
  index: number
  options: string[]
  tiebreakUsed: boolean
  votedParticipantIds: string[]
  votes?: Record<string, string>
  result?: { winner: string; decidedByCoinFlip: boolean }
}

export type GameResult = {
  mode: GameMode
  winner: string
  roundsPlayed: number
  aborted: boolean
}

// Mirrors dto.GameSnapshotResponse.
export type GameSnapshot = {
  id: string
  code: string
  state: GameState
  mode: GameMode
  hostId: string
  locked: boolean
  config: GameConfig
  teams: GameTeam[]
  participants: GameParticipant[]
  rounds: GameRound[]
  result?: GameResult
}

// Mirrors dto.GameViewerResponse.
export type GameViewer = {
  participantId: string
  teamId: string
  isHost: boolean
  hasVoted: boolean
  vote?: string
}

// Mirrors dto.GameStateResponse - the shape returned by every REST game
// route and pushed as the WebSocket STATE frame.
export type GameStateResponse = {
  game: GameSnapshot
  you: GameViewer
}

// Mirrors dto.PublicLobbyResponse.
export type PublicLobby = {
  gameId: string
  mode: GameMode
  hostDisplayName: string
  playerCount: number
  maxPlayers: number
  mangas: Manga[]
  abilitySource: AbilitySource
  allowBots: boolean
  votingWindowSeconds: number
  locked: boolean
}

// Mirrors dto.PublicLobbyListResponse.
export type PublicLobbyList = { items: PublicLobby[] }

// Mirrors dto.LobbyPreviewResponse.
export type LobbyPreview = PublicLobby & { code: string; visibility: LobbyVisibility }

// Mirrors services.CreateGameInput / dto.CreateGameRequest.
export type CreateGameInput = {
  mode: GameMode
  mangas: Manga[]
  abilitySource: AbilitySource
  teamSize: number
  allowBots: boolean
  visibility?: LobbyVisibility
  votingWindowSeconds?: number
}

// Mirrors services.ConfigUpdateInput / dto.UpdateConfigPayload - the
// UPDATE_CONFIG command's payload. A full replacement of the lobby's
// Config, not a partial patch, so every field is required (unlike
// CreateGameInput's optional visibility/votingWindowSeconds, which the
// backend only defaults on creation).
export type UpdateGameConfigInput = {
  mode: GameMode
  mangas: Manga[]
  abilitySource: AbilitySource
  teamSize: number
  allowBots: boolean
  visibility: LobbyVisibility
  votingWindowSeconds: number
  poolFilter?: PoolFilter
}
