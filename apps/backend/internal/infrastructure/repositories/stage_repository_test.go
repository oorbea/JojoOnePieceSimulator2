//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

func newTestStageRepository(t *testing.T) (*repositories.StageRepository, *pgxpool.Pool) {
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
	return repositories.NewStageRepository(pool), pool
}

func cleanupStage(t *testing.T, pool *pgxpool.Pool, id game.StageID) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), "DELETE FROM stages WHERE id = $1", id.String()); err != nil {
			t.Errorf("cleanup delete stages %s: %v", id, err)
		}
	})
}

func TestStageRepository_SeedIsPresentAndOrdered(t *testing.T) {
	repo, _ := newTestStageRepository(t)
	ctx := context.Background()

	jojo, err := repo.Stages(ctx, enums.Jojo)
	if err != nil {
		t.Fatalf("Stages(JOJO): %v", err)
	}
	if len(jojo) != 8 {
		t.Fatalf("len(jojo stages) = %d, want 8", len(jojo))
	}
	for i, st := range jojo {
		if st.Order() != i {
			t.Errorf("jojo[%d].Order() = %d, want %d", i, st.Order(), i)
		}
	}
	if jojo[0].Name() != "Phantom Blood" || jojo[7].Name() != "JoJolion" {
		t.Errorf("unexpected jojo seed order: first=%q last=%q", jojo[0].Name(), jojo[7].Name())
	}

	onePiece, err := repo.Stages(ctx, enums.OnePiece)
	if err != nil {
		t.Fatalf("Stages(ONE_PIECE): %v", err)
	}
	if len(onePiece) != 11 {
		t.Fatalf("len(one piece stages) = %d, want 11", len(onePiece))
	}
}

func TestStageRepository_CreateUpdateDeleteRoundTrip(t *testing.T) {
	repo, pool := newTestStageRepository(t)
	ctx := context.Background()

	var id [16]byte
	id[0] = 0xAB
	id[1] = 0xCD
	stageID := game.StageID(id)
	cleanupStage(t, pool, stageID)

	st, err := game.NewStage(stageID, enums.Jojo, 99, "Integration Test Stage")
	if err != nil {
		t.Fatalf("NewStage: %v", err)
	}
	if err := repo.Save(ctx, st); err != nil {
		t.Fatalf("Save (create): %v", err)
	}

	got, err := repo.FindByID(ctx, stageID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Name() != "Integration Test Stage" || got.Order() != 99 {
		t.Errorf("FindByID after create = %+v", got)
	}

	updated, err := game.NewStage(stageID, enums.OnePiece, 5, "Updated Stage")
	if err != nil {
		t.Fatalf("NewStage (update): %v", err)
	}
	if err := repo.Save(ctx, updated); err != nil {
		t.Fatalf("Save (update): %v", err)
	}
	got, err = repo.FindByID(ctx, stageID)
	if err != nil {
		t.Fatalf("FindByID after update: %v", err)
	}
	if got.Name() != "Updated Stage" || got.Manga() != enums.OnePiece || got.Order() != 5 {
		t.Errorf("FindByID after update = %+v", got)
	}

	if err := repo.Delete(ctx, stageID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, stageID); err != ports.ErrStageNotFound {
		t.Errorf("FindByID after delete = %v, want ErrStageNotFound", err)
	}
	if err := repo.Delete(ctx, stageID); err != ports.ErrStageNotFound {
		t.Errorf("Delete (again) = %v, want ErrStageNotFound", err)
	}
}

func TestStageRepository_DuplicateNameConflicts(t *testing.T) {
	repo, pool := newTestStageRepository(t)
	ctx := context.Background()

	var idA, idB [16]byte
	idA[0], idA[1] = 0xEE, 0x01
	idB[0], idB[1] = 0xEE, 0x02
	stageA, stageB := game.StageID(idA), game.StageID(idB)
	cleanupStage(t, pool, stageA)
	cleanupStage(t, pool, stageB)

	a, err := game.NewStage(stageA, enums.Jojo, 200, "Duplicate Name Stage")
	if err != nil {
		t.Fatalf("NewStage(a): %v", err)
	}
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("Save(a): %v", err)
	}

	b, err := game.NewStage(stageB, enums.Jojo, 201, "Duplicate Name Stage")
	if err != nil {
		t.Fatalf("NewStage(b): %v", err)
	}
	if err := repo.Save(ctx, b); err == nil || !errors.Is(err, ports.ErrStageAlreadyExists) {
		t.Errorf("Save(b) = %v, want wrapping ErrStageAlreadyExists", err)
	}
}

func TestStageRepository_UnknownID_NotFound(t *testing.T) {
	repo, _ := newTestStageRepository(t)
	ctx := context.Background()
	var id [16]byte
	id[0] = 0xFF
	if _, err := repo.FindByID(ctx, game.StageID(id)); err != ports.ErrStageNotFound {
		t.Errorf("FindByID(unknown) = %v, want ErrStageNotFound", err)
	}
}
