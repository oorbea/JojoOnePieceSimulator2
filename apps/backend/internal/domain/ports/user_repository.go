package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
)

type IUserRepository interface {
	Save(ctx context.Context, u *user.User) error
	FindByID(ctx context.Context, id user.UserID) (*user.User, error)
	FindByGoogleSub(ctx context.Context, sub string) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByUsername(ctx context.Context, username string) (*user.User, error)
}
