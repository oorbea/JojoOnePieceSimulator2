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
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

var testIDGen = idgen.UUIDGenerator[powers.PowerID]{}

func newTestRepo(t *testing.T) *repositories.StandRepository {
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
	return repositories.NewStandRepository(pool)
}

func newTestStand(t *testing.T, name string, evolvesFrom *powers.Stand) *powers.Stand {
	t.Helper()
	power, err := powers.NewPower(testIDGen.NewID(), name, name+" description", enums.Rare, &[]string{"punch", "dash"}, "pic.png")
	if err != nil {
		t.Fatalf("NewPower: %v", err)
	}
	stand, err := powers.NewStand(*power, enums.A, enums.B, enums.C, enums.D, enums.E, enums.Infinite, evolvesFrom)
	if err != nil {
		t.Fatalf("NewStand: %v", err)
	}
	return stand
}

// saveStand saves stand and registers a cleanup that deletes it, so reruns
// of the same test don't collide with a leftover row - now that ids (and
// hence upserts) are owned by the application, re-saving under the same
// name but a fresh id would otherwise hit the unique name constraint.
func saveStand(t *testing.T, repo *repositories.StandRepository, ctx context.Context, stand *powers.Stand) {
	t.Helper()
	if err := repo.Save(ctx, stand); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(context.Background(), stand.ID()); err != nil && !errors.Is(err, ports.ErrStandNotFound) {
			t.Errorf("cleanup Delete(%s): %v", stand.Name(), err)
		}
	})
}

func TestStandRepository_SaveAndFindByName(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Silver Chariot")
	stand := newTestStand(t, name, nil)
	saveStand(t, repo, ctx, stand)

	got, err := repo.FindByName(ctx, name)
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if got.Name() != name {
		t.Errorf("Name() = %q, want %q", got.Name(), name)
	}
	if got.EvolvesFrom() != nil {
		t.Errorf("EvolvesFrom() = %v, want nil", got.EvolvesFrom())
	}
	if len(got.Skills()) != 2 || got.Skills()[0] != "punch" || got.Skills()[1] != "dash" {
		t.Errorf("Skills() = %v, want [punch dash] in order", got.Skills())
	}
	if got.AttackPower() != enums.A {
		t.Errorf("AttackPower() = %v, want A", got.AttackPower())
	}
}

func TestStandRepository_EvolutionChain(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	parentName := uniqueName(t, "Silver Chariot")
	parent := newTestStand(t, parentName, nil)
	saveStand(t, repo, ctx, parent)

	childName := uniqueName(t, "Silver Chariot Requiem")
	child := newTestStand(t, childName, parent)
	saveStand(t, repo, ctx, child)

	got, err := repo.FindByName(ctx, childName)
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if got.EvolvesFrom() == nil {
		t.Fatalf("EvolvesFrom() = nil, want %q", parentName)
	}
	if got.EvolvesFrom().Name() != parentName {
		t.Errorf("EvolvesFrom().Name() = %q, want %q", got.EvolvesFrom().Name(), parentName)
	}
}

func TestStandRepository_SaveIsIdempotentByName(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Star Platinum")
	stand := newTestStand(t, name, nil)

	if err := repo.Save(ctx, stand); err != nil {
		t.Fatalf("Save (1st): %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(context.Background(), stand.ID()); err != nil && !errors.Is(err, ports.ErrStandNotFound) {
			t.Errorf("cleanup Delete(%s): %v", stand.Name(), err)
		}
	})
	if err := repo.Save(ctx, stand); err != nil {
		t.Fatalf("Save (2nd): %v", err)
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	count := 0
	for _, s := range all {
		if s.Name() == name {
			count++
		}
	}
	if count != 1 {
		t.Errorf("found %d rows for %q, want 1", count, name)
	}
}

func TestStandRepository_FindByName_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.FindByName(ctx, uniqueName(t, "Nonexistent Stand"))
	if !errors.Is(err, ports.ErrStandNotFound) {
		t.Errorf("err = %v, want ports.ErrStandNotFound", err)
	}
}

func TestStandRepository_Filter(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "The World")
	stand := newTestStand(t, name, nil)
	saveStand(t, repo, ctx, stand)

	attackPower := enums.A
	results, err := repo.Filter(ctx, ports.StandFilters{AttackPower: &attackPower})
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	found := false
	for _, s := range results {
		if s.Name() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("Filter(AttackPower=A) did not include %q", name)
	}

	speed := enums.Null
	results, err = repo.Filter(ctx, ports.StandFilters{Speed: &speed})
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	for _, s := range results {
		if s.Name() == name {
			t.Errorf("Filter(Speed=NULL) unexpectedly included %q (its speed is B)", name)
		}
	}
}

func TestStandRepository_FindByID(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Silver Chariot")
	stand := newTestStand(t, name, nil)
	saveStand(t, repo, ctx, stand)

	got, err := repo.FindByID(ctx, stand.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Name() != name {
		t.Errorf("Name() = %q, want %q", got.Name(), name)
	}
	if got.ID() != stand.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), stand.ID())
	}
}

func TestStandRepository_FindByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, testIDGen.NewID())
	if !errors.Is(err, ports.ErrStandNotFound) {
		t.Errorf("err = %v, want ports.ErrStandNotFound", err)
	}
}

func TestStandRepository_Delete(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	parentName := uniqueName(t, "Silver Chariot")
	parent := newTestStand(t, parentName, nil)
	if err := repo.Save(ctx, parent); err != nil {
		t.Fatalf("Save parent: %v", err)
	}
	// Deleted explicitly below rather than via saveStand's cleanup - no
	// double-delete cleanup needed for the parent.

	childName := uniqueName(t, "Silver Chariot Requiem")
	child := newTestStand(t, childName, parent)
	saveStand(t, repo, ctx, child)

	if err := repo.Delete(ctx, parent.ID()); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := repo.FindByID(ctx, parent.ID()); !errors.Is(err, ports.ErrStandNotFound) {
		t.Errorf("FindByID(parent) after delete: err = %v, want ports.ErrStandNotFound", err)
	}

	got, err := repo.FindByName(ctx, childName)
	if err != nil {
		t.Fatalf("FindByName(child): %v", err)
	}
	if got.EvolvesFrom() != nil {
		t.Errorf("EvolvesFrom() = %v, want nil after parent deletion", got.EvolvesFrom())
	}
}

// TestStandRepository_Save_RejectsSelfEvolution proves the
// stands_no_self_evolution CHECK constraint is the actual enforcement point
// (not merely a service-layer guard): it builds a Stand whose evolvesFrom
// shares its own id via two distinct powers.Stand values (the domain type
// can't literally self-reference before construction completes), bypassing
// StandService.ErrSelfEvolution entirely by calling the repository direct.
func TestStandRepository_Save_RejectsSelfEvolution(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	id := testIDGen.NewID()
	name := uniqueName(t, "Ouroboros")

	selfPower, err := powers.NewPower(id, name, name+" description", enums.Rare, &[]string{"loop"}, "pic.png")
	if err != nil {
		t.Fatalf("NewPower (self ref): %v", err)
	}
	selfRef, err := powers.NewStand(*selfPower, enums.A, enums.B, enums.C, enums.D, enums.E, enums.Infinite, nil)
	if err != nil {
		t.Fatalf("NewStand (self ref): %v", err)
	}

	power, err := powers.NewPower(id, name, name+" description", enums.Rare, &[]string{"loop"}, "pic.png")
	if err != nil {
		t.Fatalf("NewPower: %v", err)
	}
	stand, err := powers.NewStand(*power, enums.A, enums.B, enums.C, enums.D, enums.E, enums.Infinite, selfRef)
	if err != nil {
		t.Fatalf("NewStand: %v", err)
	}

	err = repo.Save(ctx, stand)
	if !errors.Is(err, ports.ErrConstraintViolation) {
		t.Fatalf("err = %v, want ports.ErrConstraintViolation", err)
	}
	// Save rolled back inside its own transaction - the powers row from
	// UpsertPower never commits, so there is nothing to clean up.
}

func TestStandRepository_Delete_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.Delete(ctx, testIDGen.NewID())
	if !errors.Is(err, ports.ErrStandNotFound) {
		t.Errorf("err = %v, want ports.ErrStandNotFound", err)
	}
}

// uniqueName keeps repeated test runs from colliding on the `name` UNIQUE
// constraint by scoping the name to the calling test.
func uniqueName(t *testing.T, base string) string {
	t.Helper()
	return base + " #" + t.Name()
}
