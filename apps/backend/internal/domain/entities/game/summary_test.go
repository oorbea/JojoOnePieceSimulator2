package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// TestGame_OpenSummary_MovesFromAssigningToSummary is OpenSummary's own
// happy path - the loadout-summary screen the owner added 2026-08-30
// between the sorteo (ASSIGNING) and the actual vote (VOTING).
func TestGame_OpenSummary_MovesFromAssigningToSummary(t *testing.T) {
	g, _ := newGauntletGame(t, oneStage(t), 1)
	assignOnly(t, g)

	if err := g.OpenSummary(); err != nil {
		t.Fatalf("OpenSummary: %v", err)
	}
	if g.State().String() != "SUMMARY" {
		t.Fatalf("state after OpenSummary = %v, want SUMMARY", g.State())
	}
}

// TestGame_OpenSummary_RejectedOutsideAssigning guards OpenSummary against
// being called on a Game that isn't mid-sorteo.
func TestGame_OpenSummary_RejectedOutsideAssigning(t *testing.T) {
	g, _ := newGauntletGame(t, oneStage(t), 1)
	assignAndOpenVoting(t, g)

	if err := g.OpenSummary(); err != game.ErrInvalidStateTransition {
		t.Fatalf("OpenSummary during VOTING = %v, want ErrInvalidStateTransition", err)
	}
}

// TestGame_OpenVoting_AcceptsBothAssigningAndSummary is the core of the
// loadout-summary feature's design: a round that never reassigns Loadouts
// (a Gauntlet round after the first) skips SUMMARY entirely and opens
// voting straight from ASSIGNING, exactly as before; a round that did
// reassign goes through SUMMARY first. OpenVoting must accept either origin.
func TestGame_OpenVoting_AcceptsBothAssigningAndSummary(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 1)
	assignOnly(t, g)

	// Straight from ASSIGNING, no summary screen at all - the pre-existing
	// path, still valid.
	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting from ASSIGNING: %v", err)
	}
	if err := g.CastVote(players[0].ID(), "SURVIVE"); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if _, err := g.CloseVoting(); err != nil {
		t.Fatalf("CloseVoting: %v", err)
	}
	if err := g.CompleteRound(); err != nil {
		t.Fatalf("CompleteRound: %v", err)
	}

	// Now via SUMMARY.
	g2, _ := newGauntletGame(t, oneStage(t), 1)
	assignOnly(t, g2)
	if err := g2.OpenSummary(); err != nil {
		t.Fatalf("OpenSummary: %v", err)
	}
	if err := g2.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting from SUMMARY: %v", err)
	}
	if g2.State().String() != "VOTING" {
		t.Fatalf("state after OpenVoting from SUMMARY = %v, want VOTING", g2.State())
	}
}

// TestGame_SummaryReadyComplete_FalseUntilAllHumansReady mirrors
// TestGame_RevealReadyComplete_FalseUntilAllHumansReady one phase later.
func TestGame_SummaryReadyComplete_FalseUntilAllHumansReady(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 3)
	assignOnly(t, g)
	if err := g.OpenSummary(); err != nil {
		t.Fatalf("OpenSummary: %v", err)
	}

	if g.SummaryReadyComplete() {
		t.Fatalf("expected SummaryReadyComplete to be false before anyone is ready")
	}
	if err := g.MarkSummaryReady(players[0].ID()); err != nil {
		t.Fatalf("MarkSummaryReady: %v", err)
	}
	if g.SummaryReadyComplete() {
		t.Fatalf("expected SummaryReadyComplete to be false with two humans left unready")
	}
	if err := g.MarkSummaryReady(players[1].ID()); err != nil {
		t.Fatalf("MarkSummaryReady: %v", err)
	}
	if err := g.MarkSummaryReady(players[2].ID()); err != nil {
		t.Fatalf("MarkSummaryReady: %v", err)
	}
	if !g.SummaryReadyComplete() {
		t.Fatalf("expected SummaryReadyComplete to be true once every human is ready")
	}
}

// TestGame_MarkSummaryReady_RejectedOutsideSummary mirrors
// TestGame_MarkRevealReady_RejectedOutsideAssigning one phase later.
func TestGame_MarkSummaryReady_RejectedOutsideSummary(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignAndOpenVoting(t, g)

	if err := g.MarkSummaryReady(players[0].ID()); err != game.ErrInvalidStateTransition {
		t.Fatalf("MarkSummaryReady during VOTING = %v, want ErrInvalidStateTransition", err)
	}
}

// TestGame_MarkSummaryReady_UnknownParticipant mirrors
// TestGame_MarkRevealReady_UnknownParticipant one phase later.
func TestGame_MarkSummaryReady_UnknownParticipant(t *testing.T) {
	g, _ := newGauntletGame(t, oneStage(t), 1)
	assignOnly(t, g)
	if err := g.OpenSummary(); err != nil {
		t.Fatalf("OpenSummary: %v", err)
	}

	if err := g.MarkSummaryReady(game.ParticipantID{99}); err != game.ErrParticipantNotFound {
		t.Fatalf("MarkSummaryReady(unknown) = %v, want ErrParticipantNotFound", err)
	}
}

// TestConfig_SummaryDurationSeconds_Bounds guards NewConfig's validation of
// the new SummaryDurationSeconds field, mirroring how VotingWindowSeconds/
// RevealSpeed are already bounds-checked.
func TestConfig_SummaryDurationSeconds_Bounds(t *testing.T) {
	build := func(seconds int) error {
		_, err := game.NewConfig(
			enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random,
			1, false, enums.Private, 30, game.PoolFilter{}, enums.Normal, seconds,
		)
		return err
	}

	if err := build(game.MinSummaryDurationSeconds - 1); err != game.ErrInvalidSummaryDuration {
		t.Fatalf("build(min-1) = %v, want ErrInvalidSummaryDuration", err)
	}
	if err := build(game.MaxSummaryDurationSeconds + 1); err != game.ErrInvalidSummaryDuration {
		t.Fatalf("build(max+1) = %v, want ErrInvalidSummaryDuration", err)
	}
	if err := build(game.DefaultSummaryDurationSeconds); err != nil {
		t.Fatalf("build(default) = %v, want nil", err)
	}
}
