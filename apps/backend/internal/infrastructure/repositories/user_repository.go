package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// UserRepository is the pgx/sqlc adapter for ports.IUserRepository.
type UserRepository struct {
	queries *db.Queries
}

var _ ports.IUserRepository = (*UserRepository)(nil)

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: db.New(pool)}
}

// Save upserts u by id, replacing every mutable field.
func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	_, err := r.queries.UpsertUser(ctx, db.UpsertUserParams{
		ID:           pgtype.UUID{Bytes: u.ID(), Valid: true},
		GoogleSub:    u.GoogleSub(),
		Email:        u.Email(),
		Username:     u.Username(),
		CompleteName: u.CompleteName(),
		Picture:      u.Picture(),
		Role:         u.Role().String(),
	})
	if err != nil {
		return fmt.Errorf("upserting user %q: %w", u.Username(), wrapPgError(err, ports.ErrUserAlreadyExists))
	}
	return nil
}

// FindByID loads the user with the given id.
func (r *UserRepository) FindByID(ctx context.Context, id user.UserID) (*user.User, error) {
	row, err := r.queries.GetUserByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", ports.ErrUserNotFound, id)
		}
		return nil, fmt.Errorf("querying user %s: %w", id, err)
	}
	return buildUser(userRowFromGetByID(row))
}

// FindByGoogleSub loads the user identified by the given Google 'sub' claim.
func (r *UserRepository) FindByGoogleSub(ctx context.Context, sub string) (*user.User, error) {
	row, err := r.queries.GetUserByGoogleSub(ctx, sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: google sub %q", ports.ErrUserNotFound, sub)
		}
		return nil, fmt.Errorf("querying user by google sub: %w", err)
	}
	return buildUser(userRowFromGetByGoogleSub(row))
}

// FindByEmail loads the user with the given email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %q", ports.ErrUserNotFound, email)
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	return buildUser(userRowFromGetByEmail(row))
}

// FindByUsername loads the user with the given username.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	row, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %q", ports.ErrUserNotFound, username)
		}
		return nil, fmt.Errorf("querying user by username: %w", err)
	}
	return buildUser(userRowFromGetByUsername(row))
}
