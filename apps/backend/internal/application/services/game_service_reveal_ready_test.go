package services_test

import (
	"context"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// TestMarkRevealReady_OpensVotingOnceEveryConnectedHumanIsReady is
// MarkRevealReady's own end-to-end happy path, mirroring
// TestVotingEndsAt_TracksThenClearsOnEarlyClose: the reveal timer set by
// StartGame must still be pending after only some connected humans call
// REVEAL_READY, and cut short the instant the last one does - without
// ever advancing the fake clock past the reveal duration.
func TestMarkRevealReady_OpensVotingOnceEveryConnectedHumanIsReady(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	guestID := mustTestUser(t, deps, "guest")

	g, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, guestID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	guestParticipant := otherParticipant(t, g, g.HostID())

	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	if g.State() != enums.Assigning {
		t.Fatalf("state after StartGame = %v, want ASSIGNING", g.State())
	}
	if _, ok := svc.RevealEndsAt(g.ID()); !ok {
		t.Fatal("RevealEndsAt: want ok while the reveal is pending")
	}

	g, err = svc.MarkRevealReady(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("MarkRevealReady (host): %v", err)
	}
	if g.State() != enums.Assigning {
		t.Fatalf("state with one of two humans ready = %v, want still ASSIGNING", g.State())
	}
	if _, ok := svc.RevealEndsAt(g.ID()); !ok {
		t.Fatal("RevealEndsAt: want still ok with one holdout left")
	}

	g, err = svc.MarkRevealReady(context.Background(), g.ID(), guestParticipant)
	if err != nil {
		t.Fatalf("MarkRevealReady (guest): %v", err)
	}
	if g.State() != enums.Voting {
		t.Fatalf("state once every human is ready = %v, want VOTING", g.State())
	}
	if _, ok := svc.RevealEndsAt(g.ID()); ok {
		t.Fatal("RevealEndsAt: want not-ok once voting has opened")
	}
}

// TestMarkRevealReady_RejectedOnceVotingHasOpened guards against a stale
// client REVEAL_READY (e.g. a message that was already in flight when the
// reveal ended on its own) mutating a Game that has moved on.
func TestMarkRevealReady_RejectedOnceVotingHasOpened(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, gauntletInput().PowerMangas)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}
	if g.State() != enums.Voting {
		t.Fatalf("state after reveal = %v, want VOTING", g.State())
	}

	if _, err := svc.MarkRevealReady(context.Background(), g.ID(), g.HostID()); err != game.ErrInvalidStateTransition {
		t.Fatalf("MarkRevealReady once VOTING = %v, want ErrInvalidStateTransition", err)
	}
}

// otherParticipant returns the ParticipantID in g that isn't exclude - used
// to find the guest a test just joined without threading its ID through
// every call.
func otherParticipant(t *testing.T, g *game.Game, exclude game.ParticipantID) game.ParticipantID {
	t.Helper()
	for _, p := range g.Participants() {
		if p.ID() != exclude {
			return p.ID()
		}
	}
	t.Fatalf("no participant other than %v found", exclude)
	return game.ParticipantID{}
}
