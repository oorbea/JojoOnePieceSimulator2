package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

func TestCreateInvite_ByHost_Success(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, expiresAt, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if token == "" {
		t.Fatal("CreateInvite returned an empty token")
	}
	if !expiresAt.After(deps.clock.Now()) {
		t.Fatalf("expiresAt %v is not in the future", expiresAt)
	}
}

func TestCreateInvite_ByNonHostMember_Success(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	_, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.JoinByCode(context.Background(), code, joinerID); err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	g, err := svc.GetGameByCode(context.Background(), code)
	if err != nil {
		t.Fatalf("GetGameByCode: %v", err)
	}
	if _, _, err := svc.CreateInvite(context.Background(), g.ID(), joinerID); err != nil {
		t.Fatalf("CreateInvite by non-host member: %v", err)
	}
}

func TestCreateInvite_ByStranger_Forbidden(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	strangerID := mustTestUser(t, deps, "stranger")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, _, err := svc.CreateInvite(context.Background(), g.ID(), strangerID); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestCreateInvite_GameAlreadyStarted_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), hostOf(t, g)); err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	if _, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID); !errors.Is(err, services.ErrGameAlreadyStarted) {
		t.Fatalf("err = %v, want ErrGameAlreadyStarted", err)
	}
}

func TestJoinByInvite_Success(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	g, err = svc.JoinByInvite(context.Background(), token, joinerID)
	if err != nil {
		t.Fatalf("JoinByInvite: %v", err)
	}
	if len(g.Participants()) != 2 {
		t.Fatalf("participants = %d, want 2", len(g.Participants()))
	}
}

// TestJoinByInvite_MultiUse_TwoDifferentUsersBothSucceed is the headline
// assertion for this feature: unlike a join code's one-shot semantics at
// the domain level (each caller still seats separately, but the *token*
// itself is never consumed), the same invite admits more than one person.
func TestJoinByInvite_MultiUse_TwoDifferentUsersBothSucceed(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerAID := mustTestUser(t, deps, "joinerA")
	joinerBID := mustTestUser(t, deps, "joinerB")

	input := gauntletInput()
	input.TeamSize = 5
	g, _, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := svc.JoinByInvite(context.Background(), token, joinerAID); err != nil {
		t.Fatalf("JoinByInvite(joinerA): %v", err)
	}
	g, err = svc.JoinByInvite(context.Background(), token, joinerBID)
	if err != nil {
		t.Fatalf("JoinByInvite(joinerB): %v", err)
	}
	if len(g.Participants()) != 3 {
		t.Fatalf("participants = %d, want 3", len(g.Participants()))
	}
}

func TestJoinByInvite_AlreadySeated_IdempotentSuccess(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	g, err = svc.JoinByInvite(context.Background(), token, hostID)
	if err != nil {
		t.Fatalf("JoinByInvite(already seated): %v", err)
	}
	if len(g.Participants()) != 1 {
		t.Fatalf("participants = %d, want 1 (idempotent, no duplicate seat)", len(g.Participants()))
	}
}

// TestJoinByInvite_AfterRegenerateGameCode_Revoked is the headline test for
// the revocation mechanism: rotating the join code must kill every invite
// minted against the old one, with no separate invalidation call.
func TestJoinByInvite_AfterRegenerateGameCode_Revoked(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := svc.RegenerateGameCode(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("RegenerateGameCode: %v", err)
	}
	if _, err := svc.JoinByInvite(context.Background(), token, joinerID); !errors.Is(err, services.ErrInviteRevoked) {
		t.Fatalf("err = %v, want ErrInviteRevoked", err)
	}
}

func TestJoinByInvite_UnknownOrExpiredToken_ReturnsInviteInvalid(t *testing.T) {
	svc, deps := newTestGameService(t)
	joinerID := mustTestUser(t, deps, "joiner")

	if _, err := svc.JoinByInvite(context.Background(), "does-not-exist", joinerID); !errors.Is(err, ports.ErrInviteInvalid) {
		t.Fatalf("err = %v, want ErrInviteInvalid", err)
	}
}

func TestJoinByInvite_PrivateLobby_Allowed(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	// gauntletInput leaves Visibility at its zero value, i.e. PRIVATE (see
	// CreateGameInput's doc) - JoinByID would 403 this with
	// game.ErrLobbyPrivate; JoinByInvite must not.
	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if g.Config().Visibility() != enums.Private {
		t.Fatalf("test setup: lobby visibility = %v, want Private", g.Config().Visibility())
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := svc.JoinByInvite(context.Background(), token, joinerID); err != nil {
		t.Fatalf("JoinByInvite against a private lobby: %v", err)
	}
}

func TestJoinByInvite_GameAlreadyStarted_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), hostOf(t, g)); err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	if _, err := svc.JoinByInvite(context.Background(), token, joinerID); !errors.Is(err, services.ErrGameAlreadyStarted) {
		t.Fatalf("err = %v, want ErrGameAlreadyStarted", err)
	}
}

func TestJoinByInvite_LobbyLocked_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := svc.SetLobbyLocked(context.Background(), g.ID(), g.HostID(), true); err != nil {
		t.Fatalf("SetLobbyLocked: %v", err)
	}
	if _, err := svc.JoinByInvite(context.Background(), token, joinerID); !errors.Is(err, game.ErrLobbyLocked) {
		t.Fatalf("err = %v, want ErrLobbyLocked", err)
	}
}

func TestJoinByInvite_GameFull_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	input := gauntletInput()
	input.TeamSize = 1
	g, _, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := svc.JoinByInvite(context.Background(), token, joinerID); !errors.Is(err, game.ErrGameFull) {
		t.Fatalf("err = %v, want ErrGameFull", err)
	}
}

func TestInviteStatus_UnknownToken_ReturnsExpired(t *testing.T) {
	svc, _ := newTestGameService(t)
	if got := svc.InviteStatus(context.Background(), "does-not-exist"); got != enums.InviteExpired {
		t.Fatalf("InviteStatus = %v, want EXPIRED", got)
	}
}

func TestInviteStatus_ValidToken_ReturnsValid(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if got := svc.InviteStatus(context.Background(), token); got != enums.InviteValid {
		t.Fatalf("InviteStatus = %v, want VALID", got)
	}
}

// TestInviteStatus_RotatedCode_ReturnsExpired is the public-endpoint half
// of the revocation guarantee - an unauthenticated visitor must see the
// link as dead, not merely be refused when they try to join.
func TestInviteStatus_RotatedCode_ReturnsExpired(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := svc.RegenerateGameCode(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("RegenerateGameCode: %v", err)
	}
	if got := svc.InviteStatus(context.Background(), token); got != enums.InviteExpired {
		t.Fatalf("InviteStatus after code rotation = %v, want EXPIRED", got)
	}
}

// TestInviteStatus_GameStartedOrFull_StillReturnsValid documents the
// deliberate non-leak: the public status route must not fold lobby state
// or fullness into its answer, since that would let an unauthenticated
// caller (including an unfurl bot) learn facts about the lobby from a
// token alone. Those causes only ever surface at redeem time, behind auth.
func TestInviteStatus_GameStarted_StillReturnsValid(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := svc.StartGame(context.Background(), g.ID(), hostOf(t, g)); err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	if got := svc.InviteStatus(context.Background(), token); got != enums.InviteValid {
		t.Fatalf("InviteStatus after game start = %v, want VALID (state is not a public fact)", got)
	}
}

func TestInvitePreview_Success(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	listing, err := svc.InvitePreview(context.Background(), token)
	if err != nil {
		t.Fatalf("InvitePreview: %v", err)
	}
	if listing.GameID != g.ID() {
		t.Fatalf("InvitePreview GameID = %v, want %v", listing.GameID, g.ID())
	}
}

func TestInvitePreview_RotatedCode_ReturnsRevoked(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	token, _, err := svc.CreateInvite(context.Background(), g.ID(), hostID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := svc.RegenerateGameCode(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("RegenerateGameCode: %v", err)
	}
	if _, err := svc.InvitePreview(context.Background(), token); !errors.Is(err, services.ErrInviteRevoked) {
		t.Fatalf("err = %v, want ErrInviteRevoked", err)
	}
}

func TestInvitePreview_UnknownToken_ReturnsInviteInvalid(t *testing.T) {
	svc, _ := newTestGameService(t)
	if _, err := svc.InvitePreview(context.Background(), "does-not-exist"); !errors.Is(err, ports.ErrInviteInvalid) {
		t.Fatalf("err = %v, want ErrInviteInvalid", err)
	}
}
