package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

func TestGame_RevealReadyComplete_FalseUntilAllHumansReady(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 3)
	assignOnly(t, g)

	if g.RevealReadyComplete() {
		t.Fatalf("expected RevealReadyComplete to be false before anyone is ready")
	}
	if err := g.MarkRevealReady(players[0].ID()); err != nil {
		t.Fatalf("MarkRevealReady: %v", err)
	}
	if g.RevealReadyComplete() {
		t.Fatalf("expected RevealReadyComplete to be false with two humans left unready")
	}
	if err := g.MarkRevealReady(players[1].ID()); err != nil {
		t.Fatalf("MarkRevealReady: %v", err)
	}
	if err := g.MarkRevealReady(players[2].ID()); err != nil {
		t.Fatalf("MarkRevealReady: %v", err)
	}
	if !g.RevealReadyComplete() {
		t.Fatalf("expected RevealReadyComplete to be true once every human is ready")
	}
}

func TestGame_MarkRevealReady_Idempotent(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignOnly(t, g)

	if err := g.MarkRevealReady(players[0].ID()); err != nil {
		t.Fatalf("MarkRevealReady: %v", err)
	}
	if err := g.MarkRevealReady(players[0].ID()); err != nil {
		t.Fatalf("MarkRevealReady (repeat): %v", err)
	}
	ready, total := g.RevealReadyProgress()
	if ready != 1 || total != 2 {
		t.Fatalf("RevealReadyProgress = (%d, %d), want (1, 2)", ready, total)
	}
}

func TestGame_MarkRevealReady_RejectedOutsideAssigning(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 2)
	assignAndOpenVoting(t, g)

	if err := g.MarkRevealReady(players[0].ID()); err != game.ErrInvalidStateTransition {
		t.Fatalf("MarkRevealReady during VOTING = %v, want ErrInvalidStateTransition", err)
	}
}

func TestGame_MarkRevealReady_UnknownParticipant(t *testing.T) {
	g, _ := newGauntletGame(t, oneStage(t), 1)
	assignOnly(t, g)

	if err := g.MarkRevealReady(game.ParticipantID{99}); err != game.ErrParticipantNotFound {
		t.Fatalf("MarkRevealReady(unknown) = %v, want ErrParticipantNotFound", err)
	}
}

// TestGame_RevealReadyComplete_ResetsEachAssigningWindow guards against a
// stale ready set leaking from one round's sorteo into the next: a
// Gauntlet's later rounds move back through ASSIGNING (see
// Game.CompleteRound) even though they never call AssignLoadouts again, so
// CompleteRound itself must also clear revealReady.
func TestGame_RevealReadyComplete_ResetsEachAssigningWindow(t *testing.T) {
	g, players := newGauntletGame(t, oneStage(t), 1)
	assignOnly(t, g)

	if err := g.MarkRevealReady(players[0].ID()); err != nil {
		t.Fatalf("MarkRevealReady: %v", err)
	}
	if !g.RevealReadyComplete() {
		t.Fatalf("expected RevealReadyComplete to be true with the lone human ready")
	}

	if err := g.OpenVoting(&fakeRandom{}); err != nil {
		t.Fatalf("OpenVoting: %v", err)
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
	if g.RevealReadyComplete() {
		t.Fatalf("expected a fresh ASSIGNING window to start with an empty ready set")
	}
}
