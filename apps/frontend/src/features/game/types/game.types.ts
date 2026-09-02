// Every type below is generated (apps/backend/internal/infrastructure/api/dto)
// - Go decides the shape, these are rename re-exports so ~40 existing
// importers keep working under the names already in use. See
// ObsidianVault/contratos-tipos-generados.md.
//
// This also retires two things the hand-written version had to work around:
// GameStage used to be redeclared here specifically to avoid a cross-feature
// import from @/features/stages (see the deleted comment above the old
// GameStage), and GameLoadout deep-imported @/features/devil-fruits/types
// and @/features/stands/types directly, both violations of the
// cross-feature-only-via-barrel norm (src/features/README.md). The
// generated GameLoadoutResponse/GameStageResponse compose StandResponse/
// DevilFruitResponse/StageResponse from '@/shared/contracts/dto' instead,
// so neither workaround is needed anymore.
export type {
  PoolFilterResponse as PoolFilter,
  GameConfigResponse as GameConfig,
  GameTeamResponse as GameTeam,
  GameLoadoutResponse as GameLoadout,
  GameStageResponse as GameStage,
  GameParticipantResponse as GameParticipant,
  GameRoundResponse as GameRound,
  ParticipantOutcomeResponse as ParticipantOutcome,
  GameResultResponse as GameResult,
  GameSnapshotResponse as GameSnapshot,
  GameViewerResponse as GameViewer,
  GameStateResponse,
  PublicLobbyResponse as PublicLobby,
  PublicLobbyListResponse as PublicLobbyList,
  LobbyPreviewResponse as LobbyPreview,
  CreateGameRequest as CreateGameInput,
  UpdateConfigPayload as UpdateGameConfigInput,
} from '@/shared/contracts/dto'
