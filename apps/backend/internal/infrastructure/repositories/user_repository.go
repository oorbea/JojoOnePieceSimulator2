package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
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

// Save upserts u by id, replacing every mutable field. Never touches u's
// avatar - that pipeline is only ever mutated through UpdateAvatar.
func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	_, err := r.queries.UpsertUser(ctx, db.UpsertUserParams{
		ID:            pgtype.UUID{Bytes: u.ID(), Valid: true},
		GoogleSub:     u.GoogleSub(),
		Email:         u.Email(),
		Username:      u.Username(),
		CompleteName:  u.CompleteName(),
		GooglePicture: u.GooglePicture(),
		Role:          u.Role().String(),
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

// UpdateUsername changes only id's username.
func (r *UserRepository) UpdateUsername(ctx context.Context, id user.UserID, username string) error {
	err := r.queries.UpdateUsername(ctx, db.UpdateUsernameParams{
		Username: username,
		ID:       pgtype.UUID{Bytes: id, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("updating username for user %s: %w", id, wrapPgError(err, ports.ErrUserAlreadyExists))
	}
	return nil
}

// UpdateLanguage changes only id's preferred locale.
func (r *UserRepository) UpdateLanguage(ctx context.Context, id user.UserID, language enums.Locale) error {
	err := r.queries.UpdateUserLanguage(ctx, db.UpdateUserLanguageParams{
		Language: language.String(),
		ID:       pgtype.UUID{Bytes: id, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("updating language for user %s: %w", id, err)
	}
	return nil
}

// UpdateAvatar updates only id's avatar renditions and pipeline status.
func (r *UserRepository) UpdateAvatar(ctx context.Context, id user.UserID, main, thumb *string, status enums.PictureStatus) error {
	err := r.queries.UpdateUserAvatar(ctx, db.UpdateUserAvatarParams{
		ID:             pgtype.UUID{Bytes: id, Valid: true},
		AvatarKey:      main,
		AvatarThumbKey: thumb,
		AvatarStatus:   status.String(),
	})
	if err != nil {
		return fmt.Errorf("updating avatar for user %s: %w", id, err)
	}
	return nil
}

// AvatarKeys returns the main and thumbnail object-storage keys currently
// stored for id's avatar.
func (r *UserRepository) AvatarKeys(ctx context.Context, id user.UserID) (string, string, error) {
	row, err := r.queries.GetUserAvatarKeys(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", fmt.Errorf("%w: %s", ports.ErrUserNotFound, id)
		}
		return "", "", fmt.Errorf("querying avatar keys for user %s: %w", id, err)
	}
	return row.AvatarKey, row.AvatarThumbKey, nil
}

// UpdateRole changes only id's role.
func (r *UserRepository) UpdateRole(ctx context.Context, id user.UserID, role enums.UserRole) error {
	err := r.queries.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		Role: role.String(),
		ID:   pgtype.UUID{Bytes: id, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("updating role for user %s: %w", id, err)
	}
	return nil
}

// Delete removes the user with the given id.
func (r *UserRepository) Delete(ctx context.Context, id user.UserID) error {
	if err := r.queries.DeleteUser(ctx, pgtype.UUID{Bytes: id, Valid: true}); err != nil {
		return fmt.Errorf("deleting user %s: %w", id, err)
	}
	return nil
}

// List returns up to limit users, ordered by creation, skipping the first
// offset.
func (r *UserRepository) List(ctx context.Context, limit, offset int32) ([]*user.User, error) {
	rows, err := r.queries.ListUsers(ctx, db.ListUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	users := make([]*user.User, 0, len(rows))
	for _, row := range rows {
		u, err := buildUser(userRowFromListUsers(row))
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// CountAdmins returns how many users currently hold the ADMIN role.
func (r *UserRepository) CountAdmins(ctx context.Context) (int64, error) {
	count, err := r.queries.CountAdmins(ctx)
	if err != nil {
		return 0, fmt.Errorf("counting admins: %w", err)
	}
	return count, nil
}
