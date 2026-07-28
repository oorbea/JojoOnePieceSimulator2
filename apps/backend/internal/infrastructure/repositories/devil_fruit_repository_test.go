//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

// testIDGen and uniqueName are shared with stand_repository_test.go (same
// package).

func newTestDevilFruitRepo(t *testing.T) *repositories.DevilFruitRepository {
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
	return repositories.NewDevilFruitRepository(pool)
}

func newTestDevilFruit(t *testing.T, name string, fruitType enums.FruitType) *powers.DevilFruit {
	t.Helper()
	power, err := powers.NewPower(testIDGen.NewID(), name, name+" description", enums.Rare, &[]string{"punch", "dash"}, "pic.png")
	if err != nil {
		t.Fatalf("NewPower: %v", err)
	}
	fruit, err := powers.NewDevilFruit(*power, fruitType)
	if err != nil {
		t.Fatalf("NewDevilFruit: %v", err)
	}
	return fruit
}

// saveDevilFruit saves fruit and registers a cleanup that deletes it, so
// reruns of the same test don't collide with a leftover row - same reasoning
// as saveStand.
func saveDevilFruit(t *testing.T, repo *repositories.DevilFruitRepository, ctx context.Context, fruit *powers.DevilFruit) {
	t.Helper()
	if err := repo.Save(ctx, fruit); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(context.Background(), fruit.ID()); err != nil && !errors.Is(err, ports.ErrDevilFruitNotFound) {
			t.Errorf("cleanup Delete(%s): %v", fruit.Name(), err)
		}
	})
}

func TestDevilFruitRepository_SaveAndFindByName(t *testing.T) {
	repo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Gomu Gomu no Mi")
	fruit := newTestDevilFruit(t, name, enums.MythicalZoan)
	saveDevilFruit(t, repo, ctx, fruit)

	got, err := repo.FindByName(ctx, name)
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if got.Name() != name {
		t.Errorf("Name() = %q, want %q", got.Name(), name)
	}
	if got.FruitType() != enums.MythicalZoan {
		t.Errorf("FruitType() = %v, want MYTHICAL_ZOAN", got.FruitType())
	}
	if len(got.Skills()) != 2 || got.Skills()[0] != "punch" || got.Skills()[1] != "dash" {
		t.Errorf("Skills() = %v, want [punch dash] in order", got.Skills())
	}
}

func TestDevilFruitRepository_SaveIsIdempotentByName(t *testing.T) {
	repo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Mera Mera no Mi")
	fruit := newTestDevilFruit(t, name, enums.Logia)

	if err := repo.Save(ctx, fruit); err != nil {
		t.Fatalf("Save (1st): %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(context.Background(), fruit.ID()); err != nil && !errors.Is(err, ports.ErrDevilFruitNotFound) {
			t.Errorf("cleanup Delete(%s): %v", fruit.Name(), err)
		}
	})
	if err := repo.Save(ctx, fruit); err != nil {
		t.Fatalf("Save (2nd): %v", err)
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	count := 0
	for _, f := range all {
		if f.Name() == name {
			count++
		}
	}
	if count != 1 {
		t.Errorf("found %d rows for %q, want 1", count, name)
	}
}

func TestDevilFruitRepository_FindByName_NotFound(t *testing.T) {
	repo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	_, err := repo.FindByName(ctx, uniqueName(t, "Nonexistent Fruit"))
	if !errors.Is(err, ports.ErrDevilFruitNotFound) {
		t.Errorf("err = %v, want ports.ErrDevilFruitNotFound", err)
	}
}

func TestDevilFruitRepository_Filter(t *testing.T) {
	repo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Suna Suna no Mi")
	fruit := newTestDevilFruit(t, name, enums.Paramecia)
	saveDevilFruit(t, repo, ctx, fruit)

	rarity := enums.Rare
	results, err := repo.Filter(ctx, ports.DevilFruitFilters{Rarity: &rarity})
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	found := false
	for _, f := range results {
		if f.Name() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("Filter(Rarity=RARE) did not include %q", name)
	}

	fruitType := enums.Logia
	results, err = repo.Filter(ctx, ports.DevilFruitFilters{FruitType: &fruitType})
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	for _, f := range results {
		if f.Name() == name {
			t.Errorf("Filter(FruitType=LOGIA) unexpectedly included %q (its type is PARAMECIA)", name)
		}
	}
}

func TestDevilFruitRepository_FindByID(t *testing.T) {
	repo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Ope Ope no Mi")
	fruit := newTestDevilFruit(t, name, enums.Paramecia)
	saveDevilFruit(t, repo, ctx, fruit)

	got, err := repo.FindByID(ctx, fruit.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Name() != name {
		t.Errorf("Name() = %q, want %q", got.Name(), name)
	}
	if got.ID() != fruit.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), fruit.ID())
	}
}

func TestDevilFruitRepository_FindByID_NotFound(t *testing.T) {
	repo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, testIDGen.NewID())
	if !errors.Is(err, ports.ErrDevilFruitNotFound) {
		t.Errorf("err = %v, want ports.ErrDevilFruitNotFound", err)
	}
}

func TestDevilFruitRepository_Delete(t *testing.T) {
	repo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Yami Yami no Mi")
	fruit := newTestDevilFruit(t, name, enums.Logia)
	if err := repo.Save(ctx, fruit); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := repo.Delete(ctx, fruit.ID()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, fruit.ID()); !errors.Is(err, ports.ErrDevilFruitNotFound) {
		t.Errorf("FindByID after delete: err = %v, want ports.ErrDevilFruitNotFound", err)
	}
}

func TestDevilFruitRepository_Delete_NotFound(t *testing.T) {
	repo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	err := repo.Delete(ctx, testIDGen.NewID())
	if !errors.Is(err, ports.ErrDevilFruitNotFound) {
		t.Errorf("err = %v, want ports.ErrDevilFruitNotFound", err)
	}
}

// TestDeleteStandByID_DoesNotDeleteDevilFruit proves the kind guard added to
// DeleteStandByID (db/query/stands.sql): a Stand's DELETE must never remove
// a power row of kind DEVIL_FRUIT, even though both subtypes share the same
// base `powers` table and the same generated id space.
func TestDeleteStandByID_DoesNotDeleteDevilFruit(t *testing.T) {
	standRepo := newTestRepo(t)
	fruitRepo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	fruitName := uniqueName(t, "Kilo Kilo no Mi")
	fruit := newTestDevilFruit(t, fruitName, enums.Paramecia)
	saveDevilFruit(t, fruitRepo, ctx, fruit)

	// A Stand delete for the fruit's id must report not-found, not silently
	// succeed against the wrong subtype.
	if err := standRepo.Delete(ctx, fruit.ID()); !errors.Is(err, ports.ErrStandNotFound) {
		t.Fatalf("standRepo.Delete(devilFruitID) err = %v, want ports.ErrStandNotFound", err)
	}

	// The fruit itself must still be there.
	if _, err := fruitRepo.FindByID(ctx, fruit.ID()); err != nil {
		t.Fatalf("FindByID(fruit) after cross-kind delete attempt: %v", err)
	}
}

// TestDeleteDevilFruitByID_DoesNotDeleteStand is the mirror image: a
// DevilFruit delete must never remove a power row of kind STAND.
func TestDeleteDevilFruitByID_DoesNotDeleteStand(t *testing.T) {
	standRepo := newTestRepo(t)
	fruitRepo := newTestDevilFruitRepo(t)
	ctx := context.Background()

	standName := uniqueName(t, "Star Platinum")
	stand := newTestStand(t, standName, nil)
	saveStand(t, standRepo, ctx, stand)

	if err := fruitRepo.Delete(ctx, stand.ID()); !errors.Is(err, ports.ErrDevilFruitNotFound) {
		t.Fatalf("fruitRepo.Delete(standID) err = %v, want ports.ErrDevilFruitNotFound", err)
	}

	if _, err := standRepo.FindByID(ctx, stand.ID()); err != nil {
		t.Fatalf("FindByID(stand) after cross-kind delete attempt: %v", err)
	}
}
