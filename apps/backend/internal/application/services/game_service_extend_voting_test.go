package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestExtendVoting_HostPushesDeadlineAndTimerBack(t *testing.T) {
	svc, deps, g := startedVotingGame(t)
	ctx := context.Background()
	opened := deps.clock.Now()

	if _, err := svc.ExtendVoting(ctx, g.ID(), g.HostID()); err != nil {
		t.Fatalf("ExtendVoting: %v", err)
	}
	endsAt, ok := svc.VotingEndsAt(g.ID())
	if !ok || !endsAt.Equal(opened.Add(40*time.Second)) {
		t.Fatalf("VotingEndsAt = %v (ok=%v), want %v", endsAt, ok, opened.Add(40*time.Second))
	}

	// The original 30s timer must be gone: still voting at +35s...
	deps.clock.Advance(35 * time.Second)
	got, err := svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if got.State() != enums.Voting {
		t.Fatalf("state at +35s = %v, want VOTING", got.State())
	}

	// ...and the extended one closes the window at +40s.
	deps.clock.Advance(6 * time.Second)
	got, err = svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if got.State() == enums.Voting {
		t.Fatal("voting should have closed once the extended deadline passed")
	}
}

func TestExtendVoting_StacksWhenPressedTwice(t *testing.T) {
	svc, deps, g := startedVotingGame(t)
	ctx := context.Background()
	opened := deps.clock.Now()

	for i := 0; i < 2; i++ {
		if _, err := svc.ExtendVoting(ctx, g.ID(), g.HostID()); err != nil {
			t.Fatalf("ExtendVoting #%d: %v", i+1, err)
		}
	}
	endsAt, _ := svc.VotingEndsAt(g.ID())
	if !endsAt.Equal(opened.Add(50 * time.Second)) {
		t.Fatalf("VotingEndsAt = %v, want %v", endsAt, opened.Add(50*time.Second))
	}
}

func TestExtendVoting_RejectsNonHost(t *testing.T) {
	svc, _, g := startedVotingGame(t)
	_, err := svc.ExtendVoting(context.Background(), g.ID(), game.ParticipantID{0xEE})
	if !errors.Is(err, game.ErrNotHost) {
		t.Fatalf("err = %v, want ErrNotHost", err)
	}
}

func TestExtendVoting_RejectsOutsideVoting(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.ExtendVoting(context.Background(), g.ID(), g.HostID()); !errors.Is(err, game.ErrVotingClosed) {
		t.Fatalf("err in LOBBY = %v, want ErrVotingClosed", err)
	}
}
