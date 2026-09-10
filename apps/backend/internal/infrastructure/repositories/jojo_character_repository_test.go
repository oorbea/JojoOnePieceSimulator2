//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

// characterIDGen is shared with one_piece_character_repository_test.go
// (same package).
var characterIDGen = idgen.UUIDGenerator[characters.CharacterID]{}

func newTestJojoCharacterRepo(t *testing.T) *repositories.JojoCharacterRepository {
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
	return repositories.NewJojoCharacterRepository(pool)
}

func newTestJojoCharacter(t *testing.T, name string, battleIQ byte) *characters.JojoCharacter {
	t.Helper()
	base, err := characters.NewCharacter(characterIDGen.NewID(), enums.Jojo, name, enums.Rare, name+" description", "pic.png")
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}
	jojo, err := characters.NewJojoCharacter(base, enums.HamonAdvanced, enums.SpinGolden, battleIQ)
	if err != nil {
		t.Fatalf("NewJojoCharacter: %v", err)
	}
	return jojo
}

// saveJojoCharacter saves c and registers a cleanup that deletes it, so
// reruns of the same test don't collide with a leftover row - same
// reasoning as saveDevilFruit.
func saveJojoCharacter(t *testing.T, repo *repositories.JojoCharacterRepository, ctx context.Context, c *characters.JojoCharacter) {
	t.Helper()
	if err := repo.Save(ctx, c, ports.CharacterTranslations{enums.EnGB: c.Description()}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(context.Background(), c.ID()); err != nil && !errors.Is(err, ports.ErrJojoCharacterNotFound) {
			t.Errorf("cleanup Delete(%s): %v", c.Name(), err)
		}
	})
}

func TestJojoCharacterRepository_SaveAndFindByName(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Jotaro Kujo")
	c := newTestJojoCharacter(t, name, 130)
	saveJojoCharacter(t, repo, ctx, c)

	got, err := repo.FindByName(ctx, name, enums.EnGB)
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if got.Name() != name {
		t.Errorf("Name() = %q, want %q", got.Name(), name)
	}
	if got.Hamon() != enums.HamonAdvanced {
		t.Errorf("Hamon() = %v, want ADVANCED", got.Hamon())
	}
	if got.Spin() != enums.SpinGolden {
		t.Errorf("Spin() = %v, want GOLDEN", got.Spin())
	}
	if got.BattleIQ() != 130 {
		t.Errorf("BattleIQ() = %d, want 130", got.BattleIQ())
	}
}

func TestJojoCharacterRepository_BattleIQZeroRoundTrips(t *testing.T) {
	// 0 is a legitimate (if absurd) admin-authored score - it must survive
	// the smallint column and not be confused with NULL/absent, the same
	// present-0 concern game.BattleIQ's wire encoding guards against.
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Forgettable Extra")
	c := newTestJojoCharacter(t, name, 0)
	saveJojoCharacter(t, repo, ctx, c)

	got, err := repo.FindByID(ctx, c.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.BattleIQ() != 0 {
		t.Errorf("BattleIQ() = %d, want 0", got.BattleIQ())
	}
}

func TestJojoCharacterRepository_SaveIsIdempotentByID(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Joseph Joestar")
	c := newTestJojoCharacter(t, name, 120)
	saveJojoCharacter(t, repo, ctx, c)

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

func TestJojoCharacterRepository_FindByName_NotFound(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	_, err := repo.FindByName(ctx, uniqueName(t, "Nonexistent Character"), enums.EnGB)
	if !errors.Is(err, ports.ErrJojoCharacterNotFound) {
		t.Errorf("err = %v, want ports.ErrJojoCharacterNotFound", err)
	}
}

func TestJojoCharacterRepository_Filter(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Giorno Giovanna")
	c := newTestJojoCharacter(t, name, 140)
	saveJojoCharacter(t, repo, ctx, c)

	hamon := enums.HamonAdvanced
	results, err := repo.Filter(ctx, ports.JojoCharacterFilters{Hamon: &hamon}, enums.EnGB)
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
		t.Errorf("Filter(Hamon=ADVANCED) did not include %q", name)
	}

	basic := enums.HamonBasic
	results, err = repo.Filter(ctx, ports.JojoCharacterFilters{Hamon: &basic}, enums.EnGB)
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	for _, got := range results {
		if got.Name() == name {
			t.Errorf("Filter(Hamon=BASIC) unexpectedly included %q (its hamon is ADVANCED)", name)
		}
	}

	iq := byte(140)
	results, err = repo.Filter(ctx, ports.JojoCharacterFilters{BattleIQ: &iq}, enums.EnGB)
	if err != nil {
		t.Fatalf("Filter(BattleIQ): %v", err)
	}
	found = false
	for _, got := range results {
		if got.Name() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("Filter(BattleIQ=140) did not include %q", name)
	}
}

func TestJojoCharacterRepository_Filter_SearchMatchesCaseInsensitivelyAndLocale(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Josuke Higashikata")
	c := newTestJojoCharacter(t, name, 100)
	if err := repo.Save(ctx, c, ports.CharacterTranslations{
		enums.EnGB: "an English-only clue",
		enums.EsES: "una pista solo en español",
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(context.Background(), c.ID()); err != nil && !errors.Is(err, ports.ErrJojoCharacterNotFound) {
			t.Errorf("cleanup Delete(%s): %v", c.Name(), err)
		}
	})

	upperName := strings.ToUpper(name[:6])
	results, err := repo.Filter(ctx, ports.JojoCharacterFilters{Search: &upperName}, enums.EnGB)
	if err != nil {
		t.Fatalf("Filter(search name): %v", err)
	}
	found := false
	for _, got := range results {
		if got.Name() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("Filter(search=%q) did not include %q", upperName, name)
	}

	needle := "español"
	esResults, err := repo.Filter(ctx, ports.JojoCharacterFilters{Search: &needle}, enums.EsES)
	if err != nil {
		t.Fatalf("Filter(search, es-ES): %v", err)
	}
	found = false
	for _, got := range esResults {
		if got.Name() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("Filter(search=%q, es-ES) did not include %q", needle, name)
	}
	enResults, err := repo.Filter(ctx, ports.JojoCharacterFilters{Search: &needle}, enums.EnGB)
	if err != nil {
		t.Fatalf("Filter(search, en-GB): %v", err)
	}
	for _, got := range enResults {
		if got.Name() == name {
			t.Errorf("Filter(search=%q, en-GB) unexpectedly included %q - its en-GB description has no Spanish text", needle, name)
		}
	}
}

func TestJojoCharacterRepository_FindByID(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Jolyne Cujoh")
	c := newTestJojoCharacter(t, name, 115)
	saveJojoCharacter(t, repo, ctx, c)

	got, err := repo.FindByID(ctx, c.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID() != c.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), c.ID())
	}
}

func TestJojoCharacterRepository_FindByID_NotFound(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, characterIDGen.NewID(), enums.EnGB)
	if !errors.Is(err, ports.ErrJojoCharacterNotFound) {
		t.Errorf("err = %v, want ports.ErrJojoCharacterNotFound", err)
	}
}

func TestJojoCharacterRepository_Delete(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Josuke Higashikata (Part 4)")
	c := newTestJojoCharacter(t, name, 110)
	if err := repo.Save(ctx, c, ports.CharacterTranslations{enums.EnGB: c.Description()}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := repo.Delete(ctx, c.ID()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, c.ID(), enums.EnGB); !errors.Is(err, ports.ErrJojoCharacterNotFound) {
		t.Errorf("FindByID after delete: err = %v, want ports.ErrJojoCharacterNotFound", err)
	}
}

func TestJojoCharacterRepository_Delete_NotFound(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	err := repo.Delete(ctx, characterIDGen.NewID())
	if !errors.Is(err, ports.ErrJojoCharacterNotFound) {
		t.Errorf("err = %v, want ports.ErrJojoCharacterNotFound", err)
	}
}

// TestJojoCharacterRepository_Delete_DoesNotDeleteOnePieceCharacter proves
// the manga guard added to DeleteCharacterByID (db/query/characters.sql):
// a JoJo character's DELETE must never remove a base characters row
// belonging to a One Piece character, even though both subtypes share the
// same base table and the same generated id space.
func TestJojoCharacterRepository_Delete_DoesNotDeleteOnePieceCharacter(t *testing.T) {
	jojoRepo := newTestJojoCharacterRepo(t)
	opRepo := newTestOnePieceCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Nico Robin")
	op := newTestOnePieceCharacter(t, name)
	saveOnePieceCharacter(t, opRepo, ctx, op)

	if err := jojoRepo.Delete(ctx, op.ID()); !errors.Is(err, ports.ErrJojoCharacterNotFound) {
		t.Fatalf("jojoRepo.Delete(onePieceID) err = %v, want ports.ErrJojoCharacterNotFound", err)
	}
	if _, err := opRepo.FindByID(ctx, op.ID(), enums.EnGB); err != nil {
		t.Fatalf("FindByID(one piece character) after cross-kind delete attempt: %v", err)
	}
}

func TestJojoCharacterRepository_Translations_ResolvesPartialTranslations(t *testing.T) {
	repo := newTestJojoCharacterRepo(t)
	ctx := context.Background()

	name := uniqueName(t, "Rohan Kishibe")
	c := newTestJojoCharacter(t, name, 128)
	if err := repo.Save(ctx, c, ports.CharacterTranslations{enums.EnGB: "an emerald splendor character"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(context.Background(), c.ID()); err != nil && !errors.Is(err, ports.ErrJojoCharacterNotFound) {
			t.Errorf("cleanup Delete(%s): %v", c.Name(), err)
		}
	})

	translations, err := repo.Translations(ctx, c.ID())
	if err != nil {
		t.Fatalf("Translations: %v", err)
	}
	if len(translations) != 1 {
		t.Fatalf("Translations() = %v, want exactly one locale (en-GB)", translations)
	}
	if translations[enums.EnGB] != "an emerald splendor character" {
		t.Errorf("Translations()[en-GB] = %q, want the en-GB translation", translations[enums.EnGB])
	}

	got, err := repo.FindByID(ctx, c.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID en-GB: %v", err)
	}
	if got.Description() != "an emerald splendor character" {
		t.Errorf("en-GB Description() = %q, want the en-GB translation", got.Description())
	}
}
