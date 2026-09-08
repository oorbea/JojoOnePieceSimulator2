package dto

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
)

// PublicUserResponse is the JSON representation of another user's public
// profile (GET /users/{id}). It deliberately excludes Email and Role - a
// caller must never be able to learn either about a user other than
// themselves through this route.
type PublicUserResponse struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	CompleteName string `json:"completeName"`
	Avatar       string `json:"avatar"`
	AvatarThumb  string `json:"avatarThumb"`
	AvatarCard   string `json:"avatarCard"`
	AvatarLqip   string `json:"avatarLqip"`
}

// NewPublicUserResponse builds a PublicUserResponse from a domain User,
// resolving its avatar the same way NewUserResponse does.
func NewPublicUserResponse(ctx context.Context, u *user.User, resolve PictureURLResolver, media MediaURLBuilder) (PublicUserResponse, error) {
	avatar, avatarThumb, avatarCard, avatarLqip, err := resolveAvatar(ctx, u, resolve, media)
	if err != nil {
		return PublicUserResponse{}, err
	}
	return PublicUserResponse{
		ID:           u.ID().String(),
		Username:     u.Username(),
		CompleteName: u.CompleteName(),
		Avatar:       avatar,
		AvatarThumb:  avatarThumb,
		AvatarCard:   avatarCard,
		AvatarLqip:   avatarLqip,
	}, nil
}
