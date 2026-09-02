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
  RevealSpeed,
  SpinLevel,
} from '@/shared/contracts/enums'
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

// Mirrors dto.GameConfigResponse. StageMangas and PowerMangas are
// independent - which manga(s) Stages come from vs. which manga(s)
// abilities/powers come from need not match.
export type GameConfig = {
  stageMangas: Manga[]
  powerMangas: Manga[]
  abilitySource: AbilitySource
  teamSize: number
  allowBots: boolean
  visibility: LobbyVisibility
  votingWindowSeconds: number
  poolFilter: PoolFilter
  revealSpeed: RevealSpeed
  summaryDurationSeconds: number
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

// Mirrors dto.GameRoundResponse. votes/result render as the round-result
// panel once the round resolves (see round-result-panel.tsx); tiedVotes is
// the exception to "hidden while live" - it's revealed as soon as a tie
// opens a revote (state TIEBREAK), so the panel can show what tied instead
// of just a "tie - revote" label.
export type GameRound = {
  index: number
  stage: GameStage
  options: string[]
  tiebreakUsed: boolean
  votedParticipantIds: string[]
  votes?: Record<string, string>
  tiedVotes?: Record<string, string>
  result?: { winner: string; decidedByCoinFlip: boolean }
}

// Mirrors dto.ParticipantOutcomeResponse - one seat as it stood the moment
// the game ended. teamId is what a VERSUS client compares against
// GameResult.winner to decide whether that seat won; in GAUNTLET the whole
// squad shares one collective outcome, so this carries identity only.
export type ParticipantOutcome = {
  participantId: string
  displayName: string
  teamId: string
  bot: boolean
}

export type GameResult = {
  mode: GameMode
  winner: string
  roundsPlayed: number
  aborted: boolean
  // Optional: a game finished by a backend predating the result screen
  // carries no participant outcomes.
  participants?: ParticipantOutcome[]
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
   * one of revealEndsAt/votingEndsAt/resultEndsAt is ever set. */
  votingEndsAt?: string
  /** RFC3339, present only while RESOLVING with a pending round-result
   * display - lets a (re)connecting client resume that countdown instead
   * of missing the panel entirely (see dto.GameSnapshotResponse.
   * ResultEndsAt). */
  resultEndsAt?: string
  /** RFC3339, present only while SUMMARY with a pending loadout-summary
   * display - lets a (re)connecting client resume that countdown instead
   * of missing the screen entirely (see dto.GameSnapshotResponse.
   * SummaryEndsAt). At most one of revealEndsAt/votingEndsAt/resultEndsAt/
   * summaryEndsAt is ever set. */
  summaryEndsAt?: string
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

// Mirrors services.CreateGameInput / dto.CreateGameRequest. StageMangas and
// PowerMangas are independent - see GameConfig's doc comment.
export type CreateGameInput = {
  mode: GameMode
  stageMangas: Manga[]
  powerMangas: Manga[]
  abilitySource: AbilitySource
  teamSize: number
  allowBots: boolean
  visibility?: LobbyVisibility
  votingWindowSeconds?: number
  poolFilter?: PoolFilter
  revealSpeed?: RevealSpeed
  summaryDurationSeconds?: number
}

// Mirrors services.ConfigUpdateInput / dto.UpdateConfigPayload - the
// UPDATE_CONFIG command's payload. A full replacement of the lobby's
// Config, not a partial patch, so every field is required (unlike
// CreateGameInput's optional visibility/votingWindowSeconds, which the
// backend only defaults on creation).
export type UpdateGameConfigInput = {
  mode: GameMode
  stageMangas: Manga[]
  powerMangas: Manga[]
  abilitySource: AbilitySource
  teamSize: number
  allowBots: boolean
  visibility: LobbyVisibility
  votingWindowSeconds: number
  poolFilter?: PoolFilter
  revealSpeed: RevealSpeed
  summaryDurationSeconds: number
}
