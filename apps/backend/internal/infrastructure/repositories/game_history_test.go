//go:build integration

package repositories_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

func newTestGameHistory(t *testing.T) (*repositories.GameHistory, *pgxpool.Pool) {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("connecting to database: %v", err)
	}
	t.Cleanup(pool.Close)
	return repositories.NewGameHistory(pool), pool
}

// insertTestUser creates a throwaway user row for a FK-satisfying
// user_id, cleaned up (and cascaded to game_result_participants.user_id via
// ON DELETE SET NULL) at test end.
func insertTestUser(t *testing.T, pool *pgxpool.Pool) user.UserID {
	t.Helper()
	raw := uuid.New()
	var id user.UserID
	copy(id[:], raw[:])
	ctx := context.Background()
	suffix := raw.String()[:8]
	_, err := pool.Exec(ctx,
		`INSERT INTO users (id, google_sub, email, username) VALUES ($1, $2, $3, $4)`,
		id.String(), "google-sub-"+suffix, "user-"+suffix+"@example.com", "user"+suffix,
	)
	if err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", id.String()); err != nil {
			t.Errorf("cleanup delete users %s: %v", id, err)
		}
	})
	return id
}

func cleanupGameResult(t *testing.T, pool *pgxpool.Pool, gameID game.GameID) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), "DELETE FROM game_results WHERE game_id = $1", gameID.String()); err != nil {
			t.Errorf("cleanup delete game_results %s: %v", gameID, err)
		}
	})
}

func TestGameHistory_Record_WithParticipants(t *testing.T) {
	history, pool := newTestGameHistory(t)
	ctx := context.Background()

	userID := insertTestUser(t, pool)

	var gID, botPID, humanPID, teamID [16]byte
	gID[0], gID[1] = 0x10, 0x01
	botPID[0], botPID[1] = 0x10, 0x02
	humanPID[0], humanPID[1] = 0x10, 0x03
	teamID[0], teamID[1] = 0x10, 0x04
	gameID := game.GameID(gID)
	cleanupGameResult(t, pool, gameID)

	result := game.GameResult{
		GameID:       gameID,
		Mode:         enums.Versus,
		Winner:       game.OptionID(game.TeamID(teamID).String()),
		RoundsPlayed: 3,
		Aborted:      false,
		Participants: []game.ParticipantOutcome{
			{ParticipantID: game.ParticipantID(humanPID), UserID: &userID, DisplayName: "Human", TeamID: game.TeamID(teamID), Bot: false},
			{ParticipantID: game.ParticipantID(botPID), UserID: nil, DisplayName: "Bot", TeamID: game.TeamID(teamID), Bot: true},
		},
	}

	if err := history.Record(ctx, result); err != nil {
		t.Fatalf("Record: %v", err)
	}

	// Idempotent: recording the same GameID again must not error.
	if err := history.Record(ctx, result); err != nil {
		t.Fatalf("Record (again): %v", err)
	}

	var mode, winner string
	var roundsPlayed int
	var aborted bool
	err := pool.QueryRow(ctx, "SELECT mode, winner, rounds_played, aborted FROM game_results WHERE game_id = $1", gameID.String()).
		Scan(&mode, &winner, &roundsPlayed, &aborted)
	if err != nil {
		t.Fatalf("querying game_results: %v", err)
	}
	if mode != "VERSUS" || roundsPlayed != 3 || aborted {
		t.Errorf("game_results row = mode=%q rounds=%d aborted=%v", mode, roundsPlayed, aborted)
	}

	var botUserID *string
	err = pool.QueryRow(ctx, "SELECT user_id FROM game_result_participants WHERE game_id = $1 AND participant_id = $2",
		gameID.String(), game.ParticipantID(botPID).String()).Scan(&botUserID)
	if err != nil {
		t.Fatalf("querying bot participant row: %v", err)
	}
	if botUserID != nil {
		t.Errorf("bot participant user_id = %v, want NULL", *botUserID)
	}

	// Deleting the user must SET NULL on the participant row, not remove it.
	if _, err := pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID.String()); err != nil {
		t.Fatalf("deleting user: %v", err)
	}
	var displayName string
	var humanUserID *string
	err = pool.QueryRow(ctx, "SELECT display_name, user_id FROM game_result_participants WHERE game_id = $1 AND participant_id = $2",
		gameID.String(), game.ParticipantID(humanPID).String()).Scan(&displayName, &humanUserID)
	if err != nil {
		t.Fatalf("querying human participant row after user delete: %v", err)
	}
	if displayName != "Human" {
		t.Errorf("display_name = %q, want %q (row must survive user deletion)", displayName, "Human")
	}
	if humanUserID != nil {
		t.Errorf("user_id after deleting the user = %v, want NULL", *humanUserID)
	}
}
