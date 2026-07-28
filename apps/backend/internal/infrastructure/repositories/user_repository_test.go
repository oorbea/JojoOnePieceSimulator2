//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

var testUserIDGen = idgen.UUIDGenerator[user.UserID]{}

// newTestUserRepo returns a UserRepository plus the underlying pool, so
// tests can clean up rows directly - IUserRepository intentionally has no
// Delete method, since nothing in the domain ever deletes a User.
func newTestUserRepo(t *testing.T) (*repositories.UserRepository, *pgxpool.Pool) {
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
	return repositories.NewUserRepository(pool), pool
}

// saveUser saves u and registers a cleanup that deletes its row.
func saveUser(t *testing.T, repo *repositories.UserRepository, pool *pgxpool.Pool, ctx context.Context, u *user.User) {
	t.Helper()
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", pgtype.UUID{Bytes: u.ID(), Valid: true})
		if err != nil {
			t.Errorf("cleanup delete user %q: %v", u.Username(), err)
		}
	})
}

func newTestUser(t *testing.T, suffix string) *user.User {
	t.Helper()
	id := testUserIDGen.NewID()
	u, err := user.NewUser(id, "google-sub-"+suffix, "jotaro-"+suffix+"@example.com", "jotaro_"+suffix, "Jotaro Kujo", "pic.png", enums.Regular)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	return u
}

func TestUserRepository_SaveAndFindByID(t *testing.T) {
	repo, pool := newTestUserRepo(t)
	ctx := context.Background()

	u := newTestUser(t, t.Name())
	saveUser(t, repo, pool, ctx, u)

	got, err := repo.FindByID(ctx, u.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Email() != u.Email() {
		t.Errorf("Email() = %q, want %q", got.Email(), u.Email())
	}
	if got.GoogleSub() != u.GoogleSub() {
		t.Errorf("GoogleSub() = %q, want %q", got.GoogleSub(), u.GoogleSub())
	}
	if got.Role() != enums.Regular {
		t.Errorf("Role() = %v, want Regular", got.Role())
	}
}

func TestUserRepository_FindByGoogleSub(t *testing.T) {
	repo, pool := newTestUserRepo(t)
	ctx := context.Background()

	u := newTestUser(t, t.Name())
	saveUser(t, repo, pool, ctx, u)

	got, err := repo.FindByGoogleSub(ctx, u.GoogleSub())
	if err != nil {
		t.Fatalf("FindByGoogleSub: %v", err)
	}
	if got.ID() != u.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), u.ID())
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	repo, pool := newTestUserRepo(t)
	ctx := context.Background()

	u := newTestUser(t, t.Name())
	saveUser(t, repo, pool, ctx, u)

	got, err := repo.FindByEmail(ctx, u.Email())
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if got.ID() != u.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), u.ID())
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	repo, pool := newTestUserRepo(t)
	ctx := context.Background()

	u := newTestUser(t, t.Name())
	saveUser(t, repo, pool, ctx, u)

	got, err := repo.FindByUsername(ctx, u.Username())
	if err != nil {
		t.Fatalf("FindByUsername: %v", err)
	}
	if got.ID() != u.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), u.ID())
	}
}

func TestUserRepository_SaveIsIdempotentByID(t *testing.T) {
	repo, pool := newTestUserRepo(t)
	ctx := context.Background()

	u := newTestUser(t, t.Name())
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save (1st): %v", err)
	}
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", pgtype.UUID{Bytes: u.ID(), Valid: true})
		if err != nil {
			t.Errorf("cleanup delete user %q: %v", u.Username(), err)
		}
	})

	promoted, err := user.NewUser(u.ID(), u.GoogleSub(), u.Email(), u.Username(), u.CompleteName(), u.Picture(), enums.Admin)
	if err != nil {
		t.Fatalf("NewUser (promoted): %v", err)
	}
	if err := repo.Save(ctx, promoted); err != nil {
		t.Fatalf("Save (2nd): %v", err)
	}

	got, err := repo.FindByID(ctx, u.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Role() != enums.Admin {
		t.Errorf("Role() = %v, want Admin after re-saving with a new role", got.Role())
	}
}

func TestUserRepository_Save_DuplicateEmailConflicts(t *testing.T) {
	repo, pool := newTestUserRepo(t)
	ctx := context.Background()

	first := newTestUser(t, t.Name())
	saveUser(t, repo, pool, ctx, first)

	second, err := user.NewUser(testUserIDGen.NewID(), "another-google-sub", first.Email(), "another_username", "", "", enums.Regular)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	if err := repo.Save(ctx, second); !errors.Is(err, ports.ErrUserAlreadyExists) {
		t.Errorf("err = %v, want ports.ErrUserAlreadyExists", err)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	repo, _ := newTestUserRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, testUserIDGen.NewID())
	if !errors.Is(err, ports.ErrUserNotFound) {
		t.Errorf("err = %v, want ports.ErrUserNotFound", err)
	}
}

func TestUserRepository_FindByGoogleSub_NotFound(t *testing.T) {
	repo, _ := newTestUserRepo(t)
	ctx := context.Background()

	_, err := repo.FindByGoogleSub(ctx, "nonexistent-sub-"+t.Name())
	if !errors.Is(err, ports.ErrUserNotFound) {
		t.Errorf("err = %v, want ports.ErrUserNotFound", err)
	}
}
