//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

func newTestOnePieceCharacterRepo(t *testing.T) *repositories.OnePieceCharacterRepository {
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
	return repositories.NewOnePieceCharacterRepository(pool)
}

func newTestOnePieceCharacter(t *testing.T, name string) *characters.OnePieceCharacter {
	t.Helper()
	base, err := characters.NewCharacter(characterIDGen.NewID(), enums.OnePiece, name, enums.Epic, name+" description", "pic.png")
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}
	op, err := characters.NewOnePieceCharacter(base, enums.PhysicalFormMarineCaptain, enums.HakiViceAdmiral, enums.HakiPrivate, enums.HakiNone, enums.FruitMasteryAdvanced)
	if err != nil {
		t.Fatalf("NewOnePieceCharacter: %v", err)
	}
	return op
}

// saveOnePieceCharacter saves c and registers a cleanup that deletes it -
// same reasoning as saveJojoCharacter.
func saveOnePieceCharacter(t *testing.T, repo *repositories.OnePieceCharacterRepository, ctx context.Context, c *characters.OnePieceCharacter) {
	t.Helper()
	if err := repo.Save(ctx, c, ports.CharacterTranslations{enums.EnGB: c.Description()}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(context.Background(), c.ID()); err != nil && !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
			t.Errorf("cleanup Delete(%s): %v", c.Name(), err)
		}
	})
}

func TestOnePieceCharacterRepository_SaveAndFindByName(t *testing.T) {
	repo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Roronoa Zoro")
	c := newTestOnePieceCharacter(t, name)
	saveOnePieceCharacter(t, repo, ctx, c)

	got, err := repo.FindByName(ctx, name, enums.EnGB)
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if got.PhysicalForm() != enums.PhysicalFormMarineCaptain {
		t.Errorf("PhysicalForm() = %v, want MARINE_CAPTAIN", got.PhysicalForm())
	}
	if got.ArmamentHaki() != enums.HakiViceAdmiral {
		t.Errorf("ArmamentHaki() = %v, want VICE_ADMIRAL", got.ArmamentHaki())
	}
	if got.ObservationHaki() != enums.HakiPrivate {
		t.Errorf("ObservationHaki() = %v, want PRIVATE", got.ObservationHaki())
	}
	if got.ConquerorHaki() != enums.HakiNone {
		t.Errorf("ConquerorHaki() = %v, want NONE", got.ConquerorHaki())
	}
	if got.FruitMastery() != enums.FruitMasteryAdvanced {
		t.Errorf("FruitMastery() = %v, want ADVANCED", got.FruitMastery())
	}
}

func TestOnePieceCharacterRepository_SaveIsIdempotentByID(t *testing.T) {
	repo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Nami")
	c := newTestOnePieceCharacter(t, name)
	saveOnePieceCharacter(t, repo, ctx, c)

	if err := repo.Save(ctx, c, ports.CharacterTranslations{enums.EnGB: c.Description()}); err != nil {
		t.Fatalf("Save (2nd): %v", err)
	}

	all, err := repo.GetAll(ctx, enums.EnGB)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	count := 0
	for _, got := range all {
		if got.Name() == name {
			count++
		}
	}
	if count != 1 {
		t.Errorf("found %d rows for %q, want 1", count, name)
	}
}

func TestOnePieceCharacterRepository_FindByName_NotFound(t *testing.T) {
	repo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	_, err := repo.FindByName(ctx, uniqueName(t, "Nonexistent Character"), enums.EnGB)
	if !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
		t.Errorf("err = %v, want ports.ErrOnePieceCharacterNotFound", err)
	}
}

func TestOnePieceCharacterRepository_Filter(t *testing.T) {
	repo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Sanji")
	c := newTestOnePieceCharacter(t, name)
	saveOnePieceCharacter(t, repo, ctx, c)

	fruitMastery := enums.FruitMasteryAdvanced
	results, err := repo.Filter(ctx, ports.OnePieceCharacterFilters{FruitMastery: &fruitMastery}, enums.EnGB)
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	found := false
	for _, got := range results {
		if got.Name() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("Filter(FruitMastery=ADVANCED) did not include %q", name)
	}

	awakened := enums.FruitMasteryAwakened
	results, err = repo.Filter(ctx, ports.OnePieceCharacterFilters{FruitMastery: &awakened}, enums.EnGB)
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	for _, got := range results {
		if got.Name() == name {
			t.Errorf("Filter(FruitMastery=AWAKENED) unexpectedly included %q (its mastery is ADVANCED)", name)
		}
	}
}

func TestOnePieceCharacterRepository_FindByID(t *testing.T) {
	repo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Usopp")
	c := newTestOnePieceCharacter(t, name)
	saveOnePieceCharacter(t, repo, ctx, c)

	got, err := repo.FindByID(ctx, c.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID() != c.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), c.ID())
	}
}

func TestOnePieceCharacterRepository_FindByID_NotFound(t *testing.T) {
	repo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, characterIDGen.NewID(), enums.EnGB)
	if !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
		t.Errorf("err = %v, want ports.ErrOnePieceCharacterNotFound", err)
	}
}

func TestOnePieceCharacterRepository_Delete(t *testing.T) {
	repo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Franky")
	c := newTestOnePieceCharacter(t, name)
	if err := repo.Save(ctx, c, ports.CharacterTranslations{enums.EnGB: c.Description()}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := repo.Delete(ctx, c.ID()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, c.ID(), enums.EnGB); !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
		t.Errorf("FindByID after delete: err = %v, want ports.ErrOnePieceCharacterNotFound", err)
	}
}

func TestOnePieceCharacterRepository_Delete_NotFound(t *testing.T) {
	repo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	err := repo.Delete(ctx, characterIDGen.NewID())
	if !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
		t.Errorf("err = %v, want ports.ErrOnePieceCharacterNotFound", err)
	}
}

// TestOnePieceCharacterRepository_Delete_DoesNotDeleteJojoCharacter is the
// mirror image of the JoJo repo's cross-kind delete test - see
// jojo_character_repository_test.go's
// TestJojoCharacterRepository_Delete_DoesNotDeleteOnePieceCharacter.
func TestOnePieceCharacterRepository_Delete_DoesNotDeleteJojoCharacter(t *testing.T) {
	jojoRepo := newTestJojoCharacterRepo(t)
	opRepo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Dio Brando")
	jojo := newTestJojoCharacter(t, name, 200)
	saveJojoCharacter(t, jojoRepo, ctx, jojo)

	if err := opRepo.Delete(ctx, jojo.ID()); !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
		t.Fatalf("opRepo.Delete(jojoID) err = %v, want ports.ErrOnePieceCharacterNotFound", err)
	}
	if _, err := jojoRepo.FindByID(ctx, jojo.ID(), enums.EnGB); err != nil {
		t.Fatalf("FindByID(jojo character) after cross-kind delete attempt: %v", err)
	}
}
