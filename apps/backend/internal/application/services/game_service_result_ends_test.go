package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// TestCastVote_ResultEndsAt_TracksThenClearsOnCompleteRound mirrors
// RevealEndsAt/VotingEndsAt's own happy-path tests (2026-08-28, added
// alongside the round-result display feature): the deadline is recorded
// the instant a clear winner parks the Game in RESOLVING, and cleared once
// the result display's own timer elapses and CompleteRound advances it.
func TestCastVote_ResultEndsAt_TracksThenClearsOnCompleteRound(t *testing.T) {
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
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	if _, ok := svc.ResultEndsAt(g.ID()); ok {
		t.Fatalf("ResultEndsAt before any round resolves: want not-ok")
	}

	// SURVIVE just advances the fixture's first of two stages - round 1
	// resolves cleanly (no tie), parking the Game in RESOLVING.
	g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), game.OptionID("SURVIVE"))
	if err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if g.State() != enums.Resolving {
		t.Fatalf("state after CastVote = %v, want RESOLVING", g.State())
	}
	want := deps.clock.Now().Add(game.ResultDuration)
	got, ok := svc.ResultEndsAt(g.ID())
	if !ok {
		t.Fatalf("ResultEndsAt during RESOLVING: want ok")
	}
	if !got.Equal(want) {
		t.Fatalf("ResultEndsAt = %v, want %v", got, want)
	}

	advanceResult(deps)
	if _, ok := svc.ResultEndsAt(g.ID()); ok {
		t.Fatalf("ResultEndsAt after CompleteRound: want not-ok")
	}
	if g.State() != enums.Voting {
		t.Fatalf("state after CompleteRound = %v, want VOTING (round 2)", g.State())
	}
}

// TestAbortGame_DuringResolving_ClearsResultEndsAt mirrors
// TestAbortGame_DuringReveal_NeverOpensVoting/
// TestAbortGame_DuringVoting_ClearsVotingEndsAt: aborting mid-result-display
// must cancel that timer outright, not let it fire into a Game that no
// longer exists.
func TestAbortGame_DuringResolving_ClearsResultEndsAt(t *testing.T) {
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
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), game.OptionID("SURVIVE"))
	if err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if g.State() != enums.Resolving {
		t.Fatalf("state after CastVote = %v, want RESOLVING", g.State())
	}

	if _, err := svc.AbortGame(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("AbortGame: %v", err)
	}
	if _, ok := svc.ResultEndsAt(g.ID()); ok {
		t.Fatal("ResultEndsAt after abort: want not-ok, the result timer should be cancelled")
	}

	advanceResult(deps)

	if _, err := svc.GetGame(context.Background(), g.ID()); !errors.Is(err, ports.ErrGameNotFound) {
		t.Fatalf("GetGame after abort+advance: err = %v, want ErrGameNotFound", err)
	}
}
