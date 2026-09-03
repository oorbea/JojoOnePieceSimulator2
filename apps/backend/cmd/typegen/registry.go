package main

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// restTypes is the allowlist of REST DTO structs cmd/typegen walks via
// reflection to emit dto.ts (plus errors.ts for ErrorResponse - see
// fileFor). Every exported struct in package dto must appear here, in
// wsOnlyTypes, or in nonWireTypeNames - registry_test.go AST-scans the
// package and fails on any exported struct that appears in none of the
// three, so a new DTO added without being registered fails `go test`
// instead of silently being missing from the generated TypeScript.
var restTypes = []any{
	dto.GoogleLoginRequest{},
	dto.LoginResponse{},
	dto.UserResponse{},
	dto.PublicUserResponse{},
	dto.UpdateProfileRequest{},
	dto.AdminUpdateUserRequest{},
	dto.UpdateRoleRequest{},
	dto.StandRequest{},
	dto.StandResponse{},
	dto.DevilFruitRequest{},
	dto.DevilFruitResponse{},
	dto.StageRequest{},
	dto.StageResponse{},
	dto.TranslationRequest{},
	dto.StageTranslationRequest{},
	dto.TranslationResponse{},
	dto.PowerTranslationsResponse{},
	dto.StageTranslationResponse{},
	dto.StageTranslationsResponse{},
	dto.CreateGameRequest{},
	dto.JoinGameRequest{},
	dto.GameConfigResponse{},
	dto.PoolFilterResponse{},
	dto.GameTeamResponse{},
	dto.GameLoadoutResponse{},
	dto.GameParticipantResponse{},
	dto.GameStageResponse{},
	dto.GameRoundResultResponse{},
	dto.GameRoundResponse{},
	dto.GameResultResponse{},
	dto.ParticipantOutcomeResponse{},
	dto.GameSnapshotResponse{},
	dto.GameViewerResponse{},
	dto.GameStateResponse{},
	dto.PublicLobbyResponse{},
	dto.PublicLobbyListResponse{},
	dto.LobbyPreviewResponse{},
	dto.ErrorResponse{},
	dto.PictureEventPayload{},
	// PoolFilterPayload/UpdateConfigPayload are shared: also WS command
	// payloads (see wsOnlyTypes' doc comment), but registered here because
	// they're also REST bodies (CreateGameRequest.PoolFilter, and
	// UpdateConfigPayload is itself the PATCH /games/{id}/config body).
	dto.PoolFilterPayload{},
	dto.UpdateConfigPayload{},
	dto.StreamTicketResponse{},
}

// wsOnlyTypes are payload structs that only ever appear as a WebSocket
// command or frame payload, never a REST body - they belong in ws.ts, not
// dto.ts. A struct referenced from both a WS payload and a REST DTO (like
// UpdateConfigPayload) goes in restTypes instead, and ws.ts imports it from
// dto.ts.
var wsOnlyTypes = []any{
	dto.AddBotPayload{},
	dto.RemoveBotPayload{},
	dto.VotePayload{},
	dto.SwitchTeamPayload{},
	dto.MovePlayerPayload{},
	dto.KickPayload{},
	dto.TransferHostPayload{},
	dto.SetLockPayload{},
	dto.RematchReadyPayload{},
	dto.VotingOpenedPayload{},
	dto.TiebreakOpenedPayload{},
	dto.VoteCastPayload{},
	dto.RevealReadyChangedPayload{},
	dto.SummaryOpenedPayload{},
	dto.SummaryReadyChangedPayload{},
	dto.PlayerJoinedPayload{},
	dto.PlayerLeftPayload{},
	dto.HostReassignedPayload{},
	dto.LoadoutsAssignedPayload{},
	dto.RoundResolvedPayload{},
	dto.GameFinishedPayload{},
	dto.GameAbortedPayload{},
	dto.ResyncRequiredPayload{},
	dto.TeamChangedPayload{},
	dto.PlayerKickedPayload{},
	dto.LobbyLockChangedPayload{},
}

// nonWireTypeNames are exported package-dto type names that are
// deliberately NOT walked: helper/resolver types with no json tags
// (GameStateDeadlines, the three func-typed resolvers, ValidationError),
// and the envelope structs (ClientCommand, ServerFrame) plus the table
// element types (FrameSpec, CommandSpec) - typegen builds ws.ts's
// discriminated unions by hand from dto.FramePayloads/CommandPayloads
// instead of reflecting over the envelopes themselves.
var nonWireTypeNames = map[string]bool{
	"GameStateDeadlines": true,
	"PictureURLResolver": true,
	"StageTextResolver":  true,
	"PowerTextResolver":  true,
	"ValidationError":    true,
	"ClientCommand":      true,
	"ServerFrame":        true,
	"FrameSpec":          true,
	"CommandSpec":        true,
}
