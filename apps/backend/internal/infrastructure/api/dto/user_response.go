package dto

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
)

// UserResponse is the JSON representation of a User. GoogleSub never
// appears here - it is an internal identity detail, not user-facing.
type UserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	CompleteName string `json:"completeName"`
	Avatar       string `json:"avatar"`
	AvatarThumb  string `json:"avatarThumb"`
	AvatarStatus string `json:"avatarStatus" ts:"PictureStatus"`
	Role         string `json:"role" ts:"UserRole"`
	Language     string `json:"language" ts:"Locale"`
}

// resolveAvatar picks the avatar to show: the user's own uploaded avatar
// (presigned through resolve, an R2 object key) if one exists, else the
// Google-synced picture (already a full external URL - never passed through
// resolve, which only knows how to presign this app's own object-storage
// keys).
func resolveAvatar(ctx context.Context, u *user.User, resolve PictureURLResolver) (main, thumb string, err error) {
	if u.AvatarKey() == "" {
		return u.GooglePicture(), "", nil
	}
	main, err = resolve(ctx, u.AvatarKey())
	if err != nil {
		return "", "", err
	}
	if u.AvatarThumbKey() != "" {
		thumb, err = resolve(ctx, u.AvatarThumbKey())
		if err != nil {
			return "", "", err
		}
	}
	return main, thumb, nil
}

// NewUserResponse builds a UserResponse from a domain User, resolving its
// avatar (own upload, or the Google-synced picture as a fallback) through
// resolve.
func NewUserResponse(ctx context.Context, u *user.User, resolve PictureURLResolver) (UserResponse, error) {
	avatar, avatarThumb, err := resolveAvatar(ctx, u, resolve)
	if err != nil {
		return UserResponse{}, err
	}
	return UserResponse{
		ID:           u.ID().String(),
		Email:        u.Email(),
		Username:     u.Username(),
		CompleteName: u.CompleteName(),
		Avatar:       avatar,
		AvatarThumb:  avatarThumb,
		AvatarStatus: u.AvatarStatus().String(),
		Role:         u.Role().String(),
		Language:     u.Language().String(),
	}, nil
}
