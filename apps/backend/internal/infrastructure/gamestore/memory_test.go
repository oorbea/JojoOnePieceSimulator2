package gamestore

import (
	"context"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func newTestGame(t *testing.T, seed byte) *game.Game {
	t.Helper()
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo},
		enums.Random, game.MaxGauntletPlayers, false, enums.Private, 30, game.PoolFilter{},
		enums.Normal, game.DefaultSummaryDurationSeconds)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	host, err := game.NewHumanParticipant(game.ParticipantID{seed, 1}, user.UserID{seed, 2}, "host", game.TeamID{seed, 10})
	if err != nil {
		t.Fatalf("NewHumanParticipant: %v", err)
	}
	team, err := game.NewTeam(game.TeamID{seed, 10}, "Squad", 0)
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}
	stage, err := game.NewStage(game.StageID{seed, 20}, enums.Jojo, 0, "Phantom Blood", "a test stage", "")
	if err != nil {
		t.Fatalf("NewStage: %v", err)
	}
	g, err := game.NewGame(game.GameID{seed, 0xAA}, cfg, host, []*game.Team{team}, []game.Stage{stage})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	return g
}

// TestMemoryStore_SaveWithTTL_ExpiresOnItsOwnShorterClock covers the
// terminal-game save. This store has no per-key expiry of its own - the
// Reaper drives it via DeleteExpired(lobbyTTL) - so the override has to be
// honored there, otherwise a finished game would linger for the full lobby
// TTL like any other entry.
func TestMemoryStore_SaveWithTTL_ExpiresOnItsOwnShorterClock(t *testing.T) {
	s := NewMemoryGameStore()
	ctx := context.Background()

	now := time.Unix(1_000_000, 0)
	s.nowFunc = func() time.Time { return now }

	shortLived := newTestGame(t, 1)
	normal := newTestGame(t, 2)

	if err := s.Create(ctx, "SHORT1", shortLived); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Create(ctx, "NORMAL", normal); err != nil {
		t.Fatalf("Create: %v", err)
	}

	const shortTTL = 15 * time.Minute
	const lobbyTTL = 2 * time.Hour

	if err := s.SaveWithTTL(ctx, shortLived, shortTTL); err != nil {
		t.Fatalf("SaveWithTTL: %v", err)
	}
	if err := s.Save(ctx, normal); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Immediately readable - the point of keeping a terminal game around.
	if _, err := s.Get(ctx, shortLived.ID()); err != nil {
		t.Fatalf("Get right after SaveWithTTL: %v", err)
	}

	// Well past the short TTL but nowhere near the lobby TTL: only the
	// short-lived entry goes.
	now = now.Add(30 * time.Minute)
	if removed := s.DeleteExpired(ctx, lobbyTTL); removed != 1 {
		t.Fatalf("DeleteExpired removed %d entries, want exactly the short-TTL one", removed)
	}
	if _, err := s.Get(ctx, shortLived.ID()); err == nil {
		t.Fatal("the short-TTL game is still present past its own TTL")
	}
	if _, err := s.Get(ctx, normal.ID()); err != nil {
		t.Fatalf("the normal game expired early: %v", err)
	}

	// And the normal one still goes when the lobby TTL does elapse.
	now = now.Add(lobbyTTL)
	if removed := s.DeleteExpired(ctx, lobbyTTL); removed != 1 {
		t.Fatalf("DeleteExpired removed %d entries, want the remaining one", removed)
	}
}

// TestMemoryStore_SaveWithTTL_ZeroMeansDefault pins that a zero override
// behaves exactly like a plain Save.
func TestMemoryStore_SaveWithTTL_ZeroMeansDefault(t *testing.T) {
	s := NewMemoryGameStore()
	ctx := context.Background()
	now := time.Unix(1_000_000, 0)
	s.nowFunc = func() time.Time { return now }

	g := newTestGame(t, 3)
	if err := s.Create(ctx, "ZERO01", g); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.SaveWithTTL(ctx, g, 0); err != nil {
		t.Fatalf("SaveWithTTL(0): %v", err)
	}

	now = now.Add(30 * time.Minute)
	if removed := s.DeleteExpired(ctx, 2*time.Hour); removed != 0 {
		t.Fatalf("DeleteExpired removed %d entries, want 0 - a zero TTL must mean the default", removed)
	}
}

// TestMemoryStore_GamesForUser covers the /games/me resume path's store
// query: a user seated in two Games at once (see joinLocked's own doc - the
// only duplicate-seat guard is within a single Game) gets both back, and a
// user with none gets an empty result rather than an error.
func TestMemoryStore_GamesForUser(t *testing.T) {
	s := NewMemoryGameStore()
	ctx := context.Background()

	g1 := newTestGame(t, 1) // host userID {1, 2}
	if err := s.Create(ctx, "GAME01", g1); err != nil {
		t.Fatalf("Create g1: %v", err)
	}

	g2 := newTestGame(t, 2) // host userID {2, 2}
	shared, err := game.NewHumanParticipant(game.ParticipantID{2, 99}, user.UserID{1, 2}, "shared", g2.Teams()[0].ID())
	if err != nil {
		t.Fatalf("NewHumanParticipant(shared): %v", err)
	}
	if err := g2.Join(shared); err != nil {
		t.Fatalf("Join(shared): %v", err)
	}
	if err := s.Create(ctx, "GAME02", g2); err != nil {
		t.Fatalf("Create g2: %v", err)
	}

	got, err := s.GamesForUser(ctx, user.UserID{1, 2})
	if err != nil {
		t.Fatalf("GamesForUser: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("GamesForUser returned %d games, want 2 (user seated in both)", len(got))
	}

	none, err := s.GamesForUser(ctx, user.UserID{9, 9})
	if err != nil {
		t.Fatalf("GamesForUser (unknown user): %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("GamesForUser for an unseated user = %d games, want 0", len(none))
	}
}
