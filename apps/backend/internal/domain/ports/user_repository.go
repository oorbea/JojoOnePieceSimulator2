package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type IUserRepository interface {
	Save(ctx context.Context, u *user.User) error
	FindByID(ctx context.Context, id user.UserID) (*user.User, error)
	FindByGoogleSub(ctx context.Context, sub string) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByUsername(ctx context.Context, username string) (*user.User, error)
	// UpdateUsername changes only id's username.
	UpdateUsername(ctx context.Context, id user.UserID, username string) error
	// UpdateAvatar updates only id's avatar renditions and pipeline status. A
	// nil main or thumb leaves that column untouched, mirroring
	// IStandRepository.UpdatePicture.
	UpdateAvatar(ctx context.Context, id user.UserID, main, thumb *string, status enums.PictureStatus) error
	// AvatarKeys returns the main and thumbnail object-storage keys currently
	// stored for id's avatar.
	AvatarKeys(ctx context.Context, id user.UserID) (main, thumb string, err error)
	// UpdateRole changes only id's role.
	UpdateRole(ctx context.Context, id user.UserID, role enums.UserRole) error
	// Delete removes the user with the given id.
	Delete(ctx context.Context, id user.UserID) error
	// List returns up to limit users, ordered by creation, skipping the
	// first offset.
	List(ctx context.Context, limit, offset int32) ([]*user.User, error)
	// CountAdmins returns how many users currently hold the ADMIN role, used
	// to guard against demoting/deleting the last admin.
	CountAdmins(ctx context.Context) (int64, error)
}
