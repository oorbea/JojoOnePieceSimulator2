import type {
  AbilitySource,
  FruitMastery,
  GameMode,
  GameState,
  HakiLevel,
  HamonLevel,
  LobbyVisibility,
  Manga,
  ParticipantKind,
  PhysicalForm,
  PictureStatus,
  SpinLevel,
} from '@/shared/lib/zod'
import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'
import type { StandResponse } from '@/features/stands/types/stands.types'

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

// Mirrors dto.GameLoadoutResponse. stand/devilFruit come straight off the
// shared feature types - don't redeclare the power shapes here. Note:
// stand.description/skills and devilFruit.description/skills are frozen in
// enums.EnGB by RepoPowerPool at draw time, but GameEndpoints re-resolves
// both to the viewer's own locale at serialization (standTextResolver/
// devilFruitTextResolver, mirroring GameStage's description below) - safe to
// render both directly, same as any other server-localized text.
export type GameLoadout = {
  stand?: StandResponse
  devilFruit?: DevilFruitResponse
  spin: SpinLevel
  hamon: HamonLevel
  fruitMastery: FruitMastery
  armamentHaki: HakiLevel
  observationHaki: HakiLevel
  conquerorHaki: HakiLevel
  physicalForm: PhysicalForm
}

// Mirrors dto.GameStageResponse. Deliberately NOT imported from
// @/features/stages - that would cross a feature boundary this codebase
// keeps clean. description is re-resolved per viewer locale server-side
// (fully localized, unlike a loadout's stand/devilFruit text above).
export type GameStage = {
  id: string
  manga: Manga
  order: number
  name: string
  description: string
  picture: string
  pictureThumb: string
  pictureStatus: PictureStatus
}

// Mirrors dto.GameParticipantResponse. avatarThumb is "" for a bot or a
// human with neither a self-uploaded avatar nor a Google-synced picture -
// never absent, so a falsy check is enough to fall back to the initial/
// robot-icon placeholder (see ParticipantTile).
export type GameParticipant = {
  id: string
  userId?: string
  displayName: string
  teamId: string
  kind: ParticipantKind
  connected: boolean
  avatarThumb: string
  loadout?: GameLoadout
}

// Mirrors dto.GameRoundResponse. Not rendered by the lobby yet - carried
// through so the socket store already has the shape ready for the
// in-match tanda.
export type GameRound = {
  index: number
  stage: GameStage
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
  /** RFC3339, present only while ASSIGNING with a pending reveal - lets a
   * (re)connecting client resume the sorteo countdown instead of restarting
   * it (see dto.GameSnapshotResponse.RevealEndsAt). */
  revealEndsAt?: string
  /** RFC3339, present only while VOTING/TIEBREAK with a pending window -
   * lets a (re)connecting client resume the vote countdown instead of
   * showing a dead bar (see dto.GameSnapshotResponse.VotingEndsAt). At most
   * one of revealEndsAt/votingEndsAt is ever set. */
  votingEndsAt?: string
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
  poolFilter?: PoolFilter
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
