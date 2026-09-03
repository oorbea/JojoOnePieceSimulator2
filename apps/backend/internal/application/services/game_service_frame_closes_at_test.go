package services_test

import (
	"context"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// TestPublish_LoadoutsAssigned_CarriesRevealDeadline and
// TestPublish_RoundResolved_CarriesResultDeadline are the service-level
// regression guard for the two frames that gained a stamped closesAt
// alongside VOTING_OPENED/TIEBREAK_OPENED/SUMMARY_OPENED: they prove the
// GameEvent handed to publish carries the *same* instant the accessor
// serves a reconnecting client, which is the property that keeps a live
// client and one that just RESYNCed from ever disagreeing. The transport
// test (TestBuildEventFrame_TimedFrames_UseStampedClosesAt in
// endpoints/game_ws_endpoints_test.go) can only check the rendering step;
// it can't catch a future emit path that publishes before its schedule*
// call arms the timer - only a service-level test that reads the
// accessor can.

func TestPublish_LoadoutsAssigned_CarriesRevealDeadline(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	sub, unsub := deps.hub.Subscribe(g.ID())
	var mu sync.Mutex
	var assigned []services.GameEvent
	done := make(chan struct{})
	go func() {
		for e := range sub {
			if _, ok := e.Event.(game.LoadoutsAssigned); ok {
				mu.Lock()
				assigned = append(assigned, e)
				mu.Unlock()
			}
		}
		close(done)
	}()

	if _, err := svc.StartGame(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("StartGame: %v", err)
	}

	unsub()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(assigned) != 1 {
		t.Fatalf("published LoadoutsAssigned events = %d, want 1", len(assigned))
	}
	if assigned[0].ClosesAt.IsZero() {
		t.Fatal("LoadoutsAssigned.ClosesAt: want non-zero, StartGame's AssignLoadouts+scheduleRevealDelay ran in the same withGame closure")
	}
	want, ok := svc.RevealEndsAt(g.ID())
	if !ok {
		t.Fatal("RevealEndsAt: want ok right after StartGame")
	}
	if !assigned[0].ClosesAt.Equal(want) {
		t.Fatalf("LoadoutsAssigned.ClosesAt = %v, want %v (the same instant RevealEndsAt serves a reconnecting client)", assigned[0].ClosesAt, want)
	}
}

func TestPublish_RoundResolved_CarriesResultDeadline(t *testing.T) {
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

	sub, unsub := deps.hub.Subscribe(g.ID())
	var mu sync.Mutex
	var resolved []services.GameEvent
	done := make(chan struct{})
	go func() {
		for e := range sub {
			if _, ok := e.Event.(game.RoundResolved); ok {
				mu.Lock()
				resolved = append(resolved, e)
				mu.Unlock()
			}
		}
		close(done)
	}()

	// SURVIVE resolves round 1 cleanly (no tie), parking the Game in
	// RESOLVING - see TestCastVote_ResultEndsAt_TracksThenClearsOnCompleteRound.
	g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), game.OptionID("SURVIVE"))
	if err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if g.State() != enums.Resolving {
		t.Fatalf("state after CastVote = %v, want RESOLVING", g.State())
	}

	unsub()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(resolved) != 1 {
		t.Fatalf("published RoundResolved events = %d, want 1", len(resolved))
	}
	if resolved[0].ClosesAt.IsZero() {
		t.Fatal("RoundResolved.ClosesAt: want non-zero, closeVoting's resolve tail scheduled the result timer in the same withGame closure")
	}
	want, ok := svc.ResultEndsAt(g.ID())
	if !ok {
		t.Fatal("ResultEndsAt: want ok during RESOLVING")
	}
	if !resolved[0].ClosesAt.Equal(want) {
		t.Fatalf("RoundResolved.ClosesAt = %v, want %v (the same instant ResultEndsAt serves a reconnecting client)", resolved[0].ClosesAt, want)
	}
}
