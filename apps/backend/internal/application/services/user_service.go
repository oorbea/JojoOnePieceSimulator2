package services

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// ErrLastAdmin is returned when an action (deleting a user, changing a
// user's role away from ADMIN) would leave the system with zero ADMINs.
var ErrLastAdmin = errors.New("cannot remove the last admin")

// UserService coordinates User profile/avatar/admin use cases. It reuses the
// PicturePolicy and picture-related sentinels declared in stand_service.go -
// avatars share the same upload rules as Stand/DevilFruit pictures.
type UserService struct {
	users     ports.IUserRepository
	pictures  ports.IPictureStorage
	processor ports.IImageProcessor
	enqueuer  ports.IPictureEnqueuer
	picPolicy PicturePolicy
}

func NewUserService(
	users ports.IUserRepository,
	pictures ports.IPictureStorage,
	processor ports.IImageProcessor,
	enqueuer ports.IPictureEnqueuer,
	picPolicy PicturePolicy,
) *UserService {
	return &UserService{users: users, pictures: pictures, processor: processor, enqueuer: enqueuer, picPolicy: picPolicy}
}

// GetByID returns the user identified by id.
func (s *UserService) GetByID(ctx context.Context, id user.UserID) (*user.User, error) {
	return s.users.FindByID(ctx, id)
}

// ChangeUsername validates and persists a new username for id.
func (s *UserService) ChangeUsername(ctx context.Context, id user.UserID, username string) (*user.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.ChangeUsername(username); err != nil {
		return nil, err
	}
	if err := s.users.UpdateUsername(ctx, id, username); err != nil {
		return nil, err
	}
	return u, nil
}

// SetAvatar validates an uploaded picture and hands it to the background
// compression worker, moving id's avatar pipeline to PENDING without
// touching the currently-served renditions. Mirrors
// StandService.SetStandPicture.
func (s *UserService) SetAvatar(ctx context.Context, id user.UserID, pic ports.Picture) (*user.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.picPolicy.MaxBytes > 0 && pic.Size > s.picPolicy.MaxBytes {
		return nil, ErrPictureTooLarge
	}
	if !s.picPolicy.allows(pic.ContentType) {
		return nil, ErrUnsupportedPictureType
	}

	buf, err := io.ReadAll(pic.Content)
	if err != nil {
		return nil, err
	}

	meta, err := s.processor.Probe(buf)
	if err != nil {
		return nil, err
	}
	if s.picPolicy.MaxPixels > 0 && int64(meta.Width)*int64(meta.Height)*int64(max(meta.Pages, 1)) > s.picPolicy.MaxPixels {
		return nil, ports.ErrInvalidImage
	}

	// Captured before touching the repo or the worker, same reasoning as
	// StandService.SetStandPicture: once Enqueue returns, the worker may
	// already have run and mutated the persisted renditions.
	previousMain, previousThumb, previousStatus := u.AvatarKey(), u.AvatarThumbKey(), u.AvatarStatus()

	if err := s.users.UpdateAvatar(ctx, id, nil, nil, enums.PicturePending); err != nil {
		return nil, err
	}

	if err := s.enqueuer.Enqueue(ports.PictureJob{SubjectID: id.String(), Kind: enums.UserSubject, Content: buf, ContentType: pic.ContentType}); err != nil {
		if revertErr := s.users.UpdateAvatar(ctx, id, nil, nil, previousStatus); revertErr != nil {
			log.Printf("reverting avatar status for user %s after enqueue failure: %v", id, revertErr)
		}
		return nil, err
	}

	u.SetAvatarRenditions(previousMain, previousThumb, enums.PicturePending)
	return u, nil
}

// DeleteAvatar clears id's self-uploaded avatar, reverting display back to
// the Google-synced picture, and best-effort deletes the now-orphaned
// object-storage keys.
func (s *UserService) DeleteAvatar(ctx context.Context, id user.UserID) (*user.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	mainKey, thumbKey := u.AvatarKey(), u.AvatarThumbKey()
	empty := ""
	if err := s.users.UpdateAvatar(ctx, id, &empty, &empty, enums.PictureNone); err != nil {
		return nil, err
	}

	if mainKey != "" {
		if err := s.pictures.Delete(ctx, mainKey); err != nil {
			log.Printf("deleting avatar %q for user %s: %v", mainKey, id, err)
		}
	}
	if thumbKey != "" {
		if err := s.pictures.Delete(ctx, thumbKey); err != nil {
			log.Printf("deleting avatar thumbnail %q for user %s: %v", thumbKey, id, err)
		}
	}

	u.SetAvatarRenditions("", "", enums.PictureNone)
	return u, nil
}

// AvatarURL resolves a stored object-storage key into a URL a client can
// GET, or "" if key is empty. Mirrors StandService.PictureURL.
func (s *UserService) AvatarURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return s.pictures.PresignGetURL(ctx, key)
}

// MaxPictureBytes exposes the configured avatar size limit so the HTTP
// handler can size its request-body guard without importing config.
func (s *UserService) MaxPictureBytes() int64 {
	return s.picPolicy.MaxBytes
}

// List returns up to limit users, ordered by creation, skipping the first
// offset.
func (s *UserService) List(ctx context.Context, limit, offset int32) ([]*user.User, error) {
	return s.users.List(ctx, limit, offset)
}

// ChangeRole changes id's role, refusing to demote the last remaining admin.
func (s *UserService) ChangeRole(ctx context.Context, id user.UserID, role enums.UserRole) (*user.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.IsAdmin() && role != enums.Admin {
		if err := s.guardLastAdmin(ctx); err != nil {
			return nil, err
		}
	}
	if err := u.ChangeRole(role); err != nil {
		return nil, err
	}
	if err := s.users.UpdateRole(ctx, id, role); err != nil {
		return nil, err
	}
	return u, nil
}

// Delete removes the user identified by id, refusing to remove the last
// remaining admin.
func (s *UserService) Delete(ctx context.Context, id user.UserID) error {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if u.IsAdmin() {
		if err := s.guardLastAdmin(ctx); err != nil {
			return err
		}
	}
	return s.users.Delete(ctx, id)
}

// guardLastAdmin returns ErrLastAdmin if there is only one admin left in the
// system.
func (s *UserService) guardLastAdmin(ctx context.Context) error {
	count, err := s.users.CountAdmins(ctx)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrLastAdmin
	}
	return nil
}
