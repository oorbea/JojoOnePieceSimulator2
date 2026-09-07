package dto

import (
	"context"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func newTestUser(t *testing.T, googlePicture string) *user.User {
	t.Helper()
	id, err := user.ParseUserID("00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("ParseUserID: %v", err)
	}
	u, err := user.NewUser(id, "google-sub", "user@example.com", "someone", "Some One", googlePicture, enums.Regular)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	return u
}

func noopResolve(_ context.Context, key string) (string, error) {
	return "https://r2.test/" + key, nil
}

// A user who never uploaded their own avatar has an empty AvatarKey, so
// resolveAvatar falls back to the Google-synced picture. AvatarThumb must
// carry that same URL - not "" - or every card bound to avatarThumb renders
// nothing for every user who never uploaded an avatar.
func TestNewUserResponse_NoOwnAvatar_UsesGooglePictureForBothMainAndThumb(t *testing.T) {
	u := newTestUser(t, "https://google.test/photo.jpg")

	resp, err := NewUserResponse(context.Background(), u, noopResolve)
	if err != nil {
		t.Fatalf("NewUserResponse: %v", err)
	}
	if resp.Avatar != "https://google.test/photo.jpg" {
		t.Errorf("Avatar = %q, want the Google picture", resp.Avatar)
	}
	if resp.AvatarThumb != "https://google.test/photo.jpg" {
		t.Errorf("AvatarThumb = %q, want the Google picture, not empty", resp.AvatarThumb)
	}
}

func TestNewUserResponse_OwnAvatar_ResolvesBothRenditionsThroughStorage(t *testing.T) {
	u := newTestUser(t, "https://google.test/photo.jpg")
	u.SetAvatarRenditions("users/x/main.webp", "users/x/thumb.webp", "users/x/card.webp", "", enums.PictureReady)

	resp, err := NewUserResponse(context.Background(), u, noopResolve)
	if err != nil {
		t.Fatalf("NewUserResponse: %v", err)
	}
	if resp.Avatar != "https://r2.test/users/x/main.webp" {
		t.Errorf("Avatar = %q, want the resolved main key", resp.Avatar)
	}
	if resp.AvatarThumb != "https://r2.test/users/x/thumb.webp" {
		t.Errorf("AvatarThumb = %q, want the resolved thumb key", resp.AvatarThumb)
	}
}

func TestNewPublicUserResponse_NoOwnAvatar_UsesGooglePictureForBothMainAndThumb(t *testing.T) {
	u := newTestUser(t, "https://google.test/photo.jpg")

	resp, err := NewPublicUserResponse(context.Background(), u, noopResolve)
	if err != nil {
		t.Fatalf("NewPublicUserResponse: %v", err)
	}
	if resp.Avatar != "https://google.test/photo.jpg" || resp.AvatarThumb != "https://google.test/photo.jpg" {
		t.Errorf("Avatar/AvatarThumb = %q/%q, want the Google picture on both", resp.Avatar, resp.AvatarThumb)
	}
}
