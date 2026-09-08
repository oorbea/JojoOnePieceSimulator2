package dto

import (
	"context"
	"time"

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
	AvatarCard   string `json:"avatarCard"`
	AvatarStatus string `json:"avatarStatus" ts:"PictureStatus"`
	AvatarLqip   string `json:"avatarLqip"`
	Role         string `json:"role" ts:"UserRole"`
	Language     string `json:"language" ts:"Locale"`
}

// resolveAvatar picks the avatar to show: the user's own uploaded avatar
// (presigned through resolve, an R2 object key) if one exists, else the
// Google-synced picture (already a full external URL - never passed through
// resolve, which only knows how to presign this app's own object-storage
// keys). The Google picture is used for main/thumb/card alike: it is already
// a small, externally-hosted image, and leaving a rendition empty made every
// card bound to it render nothing for a user who never uploaded one. lqip is
// always empty for the Google fallback - there is no local pipeline output
// to embed for an external URL.
func resolveAvatar(ctx context.Context, u *user.User, resolve PictureURLResolver, media MediaURLBuilder) (main, thumb, card, lqip string, err error) {
	if u.AvatarKey() == "" {
		return u.GooglePicture(), u.GooglePicture(), u.GooglePicture(), "", nil
	}
	if mediaID := u.AvatarMediaID(); mediaID != "" {
		now := time.Now()
		return media.Private(mediaID, "main", now), media.Private(mediaID, "thumb", now),
			media.Private(mediaID, "card", now), u.AvatarLqip(), nil
	}
	main, err = resolve(ctx, u.AvatarKey())
	if err != nil {
		return "", "", "", "", err
	}
	if u.AvatarThumbKey() != "" {
		thumb, err = resolve(ctx, u.AvatarThumbKey())
		if err != nil {
			return "", "", "", "", err
		}
	}
	if u.AvatarCardKey() != "" {
		card, err = resolve(ctx, u.AvatarCardKey())
		if err != nil {
			return "", "", "", "", err
		}
	}
	return main, thumb, card, u.AvatarLqip(), nil
}

// NewUserResponse builds a UserResponse from a domain User, resolving its
// avatar (own upload, or the Google-synced picture as a fallback) through
// resolve, or through media's signed private URLs once AvatarMediaID is
// backfilled.
func NewUserResponse(ctx context.Context, u *user.User, resolve PictureURLResolver, media MediaURLBuilder) (UserResponse, error) {
	avatar, avatarThumb, avatarCard, avatarLqip, err := resolveAvatar(ctx, u, resolve, media)
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
		AvatarCard:   avatarCard,
		AvatarStatus: u.AvatarStatus().String(),
		AvatarLqip:   avatarLqip,
		Role:         u.Role().String(),
		Language:     u.Language().String(),
	}, nil
}
