package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// GameHistory is the Postgres-backed ports.IGameHistory adapter, backed by
// game_results/game_result_participants. Record is called once by
// GameService.finalizeLocked right before a finished/aborted Game is
// deleted from its live store - best-effort on that caller's side (it only
// logs a failure), so game_id is the table's PK specifically to make Record
// idempotent under a retry (ON CONFLICT DO NOTHING, see db/query/game_history.sql).
type GameHistory struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

var _ ports.IGameHistory = (*GameHistory)(nil)

// NewGameHistory builds a GameHistory over pool.
func NewGameHistory(pool *pgxpool.Pool) *GameHistory {
	return &GameHistory{pool: pool, queries: db.New(pool)}
}

// Record implements ports.IGameHistory, writing the result row and every
// participant row in one transaction.
func (r *GameHistory) Record(ctx context.Context, result game.GameResult) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			fmt.Printf("rolling back transaction recording game result %s: %v\n", result.GameID, rollbackErr)
		}
	}()

	q := r.queries.WithTx(tx)

	if err := q.RecordGameResult(ctx, db.RecordGameResultParams{
		GameID:       pgtype.UUID{Bytes: result.GameID, Valid: true},
		Mode:         result.Mode.String(),
		Winner:       string(result.Winner),
		RoundsPlayed: int32(result.RoundsPlayed),
		Aborted:      result.Aborted,
	}); err != nil {
		return fmt.Errorf("recording game result %s: %w", result.GameID, wrapPgError(err, ports.ErrConstraintViolation))
	}

	for _, p := range result.Participants {
		userID := pgtype.UUID{}
		if p.UserID != nil {
			userID = pgtype.UUID{Bytes: *p.UserID, Valid: true}
		}
		if err := q.RecordGameResultParticipant(ctx, db.RecordGameResultParticipantParams{
			GameID:        pgtype.UUID{Bytes: result.GameID, Valid: true},
			ParticipantID: pgtype.UUID{Bytes: p.ParticipantID, Valid: true},
			UserID:        userID,
			DisplayName:   p.DisplayName,
			TeamID:        pgtype.UUID{Bytes: p.TeamID, Valid: true},
			IsBot:         p.Bot,
		}); err != nil {
			return fmt.Errorf("recording game result participant %s for game %s: %w", p.ParticipantID, result.GameID, wrapPgError(err, ports.ErrConstraintViolation))
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing game result %s: %w", result.GameID, err)
	}
	return nil
}
