package services

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// PicturePublisher is the minimal slice of a subject repository the
// PictureWorker needs: read the currently-served renditions (to know which
// object keys become orphaned) and publish new ones. Keeping this separate
// from ports.IStandRepository/IDevilFruitRepository/IUserRepository means
// the worker only depends on what it actually uses, and adding a new subject
// never requires touching those repository interfaces or their existing
// fakes. id is the subject's id formatted as a string; each adapter parses
// it back into its own concrete id type.
type PicturePublisher interface {
	// PictureKeys returns the main and thumbnail object-storage keys
	// currently stored for id.
	PictureKeys(ctx context.Context, id string) (main, thumb string, err error)
	// UpdatePicture updates only the picture renditions and pipeline status
	// for id. A nil main or thumb leaves that column untouched.
	UpdatePicture(ctx context.Context, id string, main, thumb *string, status enums.PictureStatus) error
}

// PictureTarget pairs a PicturePublisher with the object-storage key prefix
// its pictures are stored under (e.g. "stands", "devil-fruits", "users").
type PictureTarget struct {
	Publisher PicturePublisher
	KeyPrefix string
}

// standPicturePublisher adapts a ports.IStandRepository to PicturePublisher.
type standPicturePublisher struct {
	repo ports.IStandRepository
}

// NewStandPicturePublisher wraps repo so the picture worker can publish
// transcoded renditions onto Stands without depending on the full
// ports.IStandRepository surface.
func NewStandPicturePublisher(repo ports.IStandRepository) PicturePublisher {
	return &standPicturePublisher{repo: repo}
}

func (p *standPicturePublisher) PictureKeys(ctx context.Context, id string) (string, string, error) {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return "", "", err
	}
	stand, err := p.repo.FindByID(ctx, powerID, enums.EnGB)
	if err != nil {
		return "", "", err
	}
	return stand.Picture(), stand.PictureThumb(), nil
}

func (p *standPicturePublisher) UpdatePicture(ctx context.Context, id string, main, thumb *string, status enums.PictureStatus) error {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return err
	}
	return p.repo.UpdatePicture(ctx, powerID, main, thumb, status)
}

// devilFruitPicturePublisher adapts a ports.IDevilFruitRepository to
// PicturePublisher.
type devilFruitPicturePublisher struct {
	repo ports.IDevilFruitRepository
}

// NewDevilFruitPicturePublisher wraps repo so the picture worker can publish
// transcoded renditions onto DevilFruits without depending on the full
// ports.IDevilFruitRepository surface.
func NewDevilFruitPicturePublisher(repo ports.IDevilFruitRepository) PicturePublisher {
	return &devilFruitPicturePublisher{repo: repo}
}

func (p *devilFruitPicturePublisher) PictureKeys(ctx context.Context, id string) (string, string, error) {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return "", "", err
	}
	fruit, err := p.repo.FindByID(ctx, powerID, enums.EnGB)
	if err != nil {
		return "", "", err
	}
	return fruit.Picture(), fruit.PictureThumb(), nil
}

func (p *devilFruitPicturePublisher) UpdatePicture(ctx context.Context, id string, main, thumb *string, status enums.PictureStatus) error {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return err
	}
	return p.repo.UpdatePicture(ctx, powerID, main, thumb, status)
}

// userPicturePublisher adapts a ports.IUserRepository to PicturePublisher, so
// the picture worker can publish transcoded avatar renditions onto Users.
type userPicturePublisher struct {
	repo ports.IUserRepository
}

// NewUserPicturePublisher wraps repo so the picture worker can publish
// transcoded avatar renditions onto Users without depending on the full
// ports.IUserRepository surface.
func NewUserPicturePublisher(repo ports.IUserRepository) PicturePublisher {
	return &userPicturePublisher{repo: repo}
}

func (p *userPicturePublisher) PictureKeys(ctx context.Context, id string) (string, string, error) {
	userID, err := user.ParseUserID(id)
	if err != nil {
		return "", "", err
	}
	return p.repo.AvatarKeys(ctx, userID)
}

func (p *userPicturePublisher) UpdatePicture(ctx context.Context, id string, main, thumb *string, status enums.PictureStatus) error {
	userID, err := user.ParseUserID(id)
	if err != nil {
		return err
	}
	return p.repo.UpdateAvatar(ctx, userID, main, thumb, status)
}
