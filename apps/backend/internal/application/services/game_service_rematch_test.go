package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// TestRematch_CarriesRosterOntoAFreshLobby is the happy path: same config,
// same seats, same teams, new ids, new code, LOBBY.
func TestRematch_CarriesRosterOntoAFreshLobby(t *testing.T) {
	svc, deps := newTestGameService(t)
	ctx := context.Background()
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	old, oldCode, err := svc.CreateGame(ctx, hostID, versusInput(3))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.JoinByCode(ctx, oldCode, joinerID); err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	old, err = svc.GetGame(ctx, old.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	host := hostOf(t, old)
	if _, err := svc.AddBot(ctx, old.ID(), host, old.Teams()[1].ID()); err != nil {
		t.Fatalf("AddBot: %v", err)
	}

	old, err = svc.AbortGame(ctx, old.ID(), host)
	if err != nil {
		t.Fatalf("AbortGame: %v", err)
	}
	if old.State() != enums.Aborted {
		t.Fatalf("state = %v, want ABORTED", old.State())
	}

	oldParticipants := old.Participants()
	oldSnapshot := old.Snapshot()

	fresh, newCode, err := svc.Rematch(ctx, old.ID(), host)
	if err != nil {
		t.Fatalf("Rematch: %v", err)
	}

	if fresh.ID() == old.ID() {
		t.Fatal("rematch reused the source game's id, want a brand new one")
	}
	if newCode == oldCode {
		t.Fatalf("rematch reused the source join code %q, want a new one", newCode)
	}
	if fresh.State() != enums.Lobby {
		t.Fatalf("new game state = %v, want LOBBY", fresh.State())
	}
	if fresh.Config().Mode() != old.Config().Mode() ||
		fresh.Config().TeamSize() != old.Config().TeamSize() ||
		fresh.Config().VotingWindowSeconds() != old.Config().VotingWindowSeconds() {
		t.Fatalf("config not carried over: %+v vs %+v", fresh.Config(), old.Config())
	}

	// Same roster, same size, same team positions.
	if got, want := len(fresh.Participants()), len(oldParticipants); got != want {
		t.Fatalf("new roster has %d seats, want %d", got, want)
	}
	teamIndex := func(g *game.Game, id game.TeamID) int {
		for i, tm := range g.Teams() {
			if tm.ID() == id {
				return i
			}
		}
		return -1
	}
	for i, oldP := range oldParticipants {
		newP := fresh.Participants()[i]
		if newP.DisplayName() != oldP.DisplayName() {
			t.Errorf("seat %d name = %q, want %q", i, newP.DisplayName(), oldP.DisplayName())
		}
		if newP.IsBot() != oldP.IsBot() {
			t.Errorf("seat %d bot = %v, want %v", i, newP.IsBot(), oldP.IsBot())
		}
		if got, want := teamIndex(fresh, newP.TeamID()), teamIndex(old, oldP.TeamID()); got != want {
			t.Errorf("seat %d landed on team index %d, want %d", i, got, want)
		}
		if newP.ID() == oldP.ID() {
			t.Errorf("seat %d reused the old participant id", i)
		}
	}

	// The source game is untouched: still terminal, still readable, byte-for
	// -byte the same aggregate it was before the rematch.
	after, err := svc.GetGame(ctx, old.ID())
	if err != nil {
		t.Fatalf("the source game stopped being readable after a rematch: %v", err)
	}
	if after.State() != enums.Aborted {
		t.Fatalf("source game state = %v, want it left ABORTED", after.State())
	}
	if got := after.Snapshot(); !snapshotsEquivalent(oldSnapshot, got) {
		t.Fatal("the source game was mutated by the rematch, want it left completely untouched")
	}
}

// snapshotsEquivalent compares the fields a rematch could plausibly have
// disturbed on the source aggregate.
func snapshotsEquivalent(a, b game.Snapshot) bool {
	if a.ID != b.ID || a.State != b.State || a.HostID != b.HostID || a.Locked != b.Locked {
		return false
	}
	if len(a.Participants) != len(b.Participants) || len(a.Rounds) != len(b.Rounds) || len(a.Teams) != len(b.Teams) {
		return false
	}
	for i := range a.Participants {
		if a.Participants[i].ID != b.Participants[i].ID || a.Participants[i].TeamID != b.Participants[i].TeamID {
			return false
		}
	}
	return true
}

func TestRematch_NonHost_Rejected(t *testing.T) {
	svc, deps := newTestGameService(t)
	ctx := context.Background()
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(ctx, hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.JoinByCode(ctx, code, joinerID); err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	g, err = svc.GetGame(ctx, g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	host := hostOf(t, g)
	if _, err := svc.AbortGame(ctx, g.ID(), host); err != nil {
		t.Fatalf("AbortGame: %v", err)
	}

	var nonHost game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != host {
			nonHost = p.ID()
		}
	}
	if nonHost == (game.ParticipantID{}) {
		t.Fatal("test setup: no non-host participant")
	}

	if _, _, err := svc.Rematch(ctx, g.ID(), nonHost); !errors.Is(err, game.ErrNotHost) {
		t.Fatalf("err = %v, want ErrNotHost", err)
	}
}

func TestRematch_GameStillInProgress_Rejected(t *testing.T) {
	svc, deps := newTestGameService(t)
	ctx := context.Background()
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(ctx, hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	// Still LOBBY.
	if _, _, err := svc.Rematch(ctx, g.ID(), hostOf(t, g)); !errors.Is(err, services.ErrGameNotOver) {
		t.Fatalf("err = %v, want ErrGameNotOver for a lobby that never started", err)
	}

	// Still mid-match.
	if _, err := svc.StartGame(ctx, g.ID(), hostOf(t, g)); err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	if _, _, err := svc.Rematch(ctx, g.ID(), hostOf(t, g)); !errors.Is(err, services.ErrGameNotOver) {
		t.Fatalf("err = %v, want ErrGameNotOver for a game in progress", err)
	}
}

// TestRematch_PublishesToEveryOldSubscriber pins that REMATCH_READY reaches
// a second connection subscribed to the OLD game, not just the requester's.
func TestRematch_PublishesToEveryOldSubscriber(t *testing.T) {
	svc, deps := newTestGameService(t)
	ctx := context.Background()
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(ctx, hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	host := hostOf(t, g)
	if _, err := svc.AbortGame(ctx, g.ID(), host); err != nil {
		t.Fatalf("AbortGame: %v", err)
	}

	// Two independent subscribers on the old game, as two browser tabs would
	// be.
	first, unsubFirst := deps.hub.Subscribe(g.ID())
	defer unsubFirst()
	second, unsubSecond := deps.hub.Subscribe(g.ID())
	defer unsubSecond()

	fresh, _, err := svc.Rematch(ctx, g.ID(), host)
	if err != nil {
		t.Fatalf("Rematch: %v", err)
	}

	for i, ch := range []<-chan services.GameEvent{first, second} {
		var got *game.RematchReady
		for len(ch) > 0 {
			evt := <-ch
			if rr, ok := evt.Event.(game.RematchReady); ok {
				if evt.GameID != g.ID() {
					t.Fatalf("subscriber %d: event published on game %s, want the OLD game %s", i, evt.GameID, g.ID())
				}
				got = &rr
			}
		}
		if got == nil {
			t.Fatalf("subscriber %d never received REMATCH_READY", i)
		}
		if got.GameID != fresh.ID() {
			t.Fatalf("subscriber %d: REMATCH_READY carries %s, want the NEW game id %s", i, got.GameID, fresh.ID())
		}
	}
}
