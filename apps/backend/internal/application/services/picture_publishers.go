package services

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
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
	// PictureKeys returns the main, thumbnail and card object-storage keys
	// currently stored for id.
	PictureKeys(ctx context.Context, id string) (main, thumb, card string, err error)
	// UpdatePicture updates only the picture renditions and pipeline status
	// for id. A nil main/thumb/card/lqip leaves that column untouched.
	UpdatePicture(ctx context.Context, id string, main, thumb, card, lqip *string, status enums.PictureStatus) error
	// SetMediaID updates only the content-addressed media group id, once the
	// worker has persisted the corresponding media_objects rows.
	SetMediaID(ctx context.Context, id string, mediaID string) error
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

func (p *standPicturePublisher) PictureKeys(ctx context.Context, id string) (string, string, string, error) {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return "", "", "", err
	}
	stand, err := p.repo.FindByID(ctx, powerID, enums.EnGB)
	if err != nil {
		return "", "", "", err
	}
	return stand.Picture(), stand.PictureThumb(), stand.PictureCard(), nil
}

func (p *standPicturePublisher) UpdatePicture(ctx context.Context, id string, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return err
	}
	return p.repo.UpdatePicture(ctx, powerID, main, thumb, card, lqip, status)
}

func (p *standPicturePublisher) SetMediaID(ctx context.Context, id string, mediaID string) error {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return err
	}
	return p.repo.SetMediaID(ctx, powerID, mediaID)
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

func (p *devilFruitPicturePublisher) PictureKeys(ctx context.Context, id string) (string, string, string, error) {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return "", "", "", err
	}
	fruit, err := p.repo.FindByID(ctx, powerID, enums.EnGB)
	if err != nil {
		return "", "", "", err
	}
	return fruit.Picture(), fruit.PictureThumb(), fruit.PictureCard(), nil
}

func (p *devilFruitPicturePublisher) UpdatePicture(ctx context.Context, id string, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return err
	}
	return p.repo.UpdatePicture(ctx, powerID, main, thumb, card, lqip, status)
}

func (p *devilFruitPicturePublisher) SetMediaID(ctx context.Context, id string, mediaID string) error {
	powerID, err := powers.ParsePowerID(id)
	if err != nil {
		return err
	}
	return p.repo.SetMediaID(ctx, powerID, mediaID)
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

func (p *userPicturePublisher) PictureKeys(ctx context.Context, id string) (string, string, string, error) {
	userID, err := user.ParseUserID(id)
	if err != nil {
		return "", "", "", err
	}
	return p.repo.AvatarKeys(ctx, userID)
}

func (p *userPicturePublisher) UpdatePicture(ctx context.Context, id string, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	userID, err := user.ParseUserID(id)
	if err != nil {
		return err
	}
	return p.repo.UpdateAvatar(ctx, userID, main, thumb, card, lqip, status)
}

func (p *userPicturePublisher) SetMediaID(ctx context.Context, id string, mediaID string) error {
	userID, err := user.ParseUserID(id)
	if err != nil {
		return err
	}
	return p.repo.SetAvatarMediaID(ctx, userID, mediaID)
}

// stagePicturePublisher adapts a ports.IStageRepository to PicturePublisher.
type stagePicturePublisher struct {
	repo ports.IStageRepository
}

// NewStagePicturePublisher wraps repo so the picture worker can publish
// transcoded renditions onto Stages without depending on the full
// ports.IStageRepository surface.
func NewStagePicturePublisher(repo ports.IStageRepository) PicturePublisher {
	return &stagePicturePublisher{repo: repo}
}

func (p *stagePicturePublisher) PictureKeys(ctx context.Context, id string) (string, string, string, error) {
	stageID, err := game.ParseStageID(id)
	if err != nil {
		return "", "", "", err
	}
	st, err := p.repo.FindByID(ctx, stageID, enums.EnGB)
	if err != nil {
		return "", "", "", err
	}
	return st.Picture(), st.PictureThumb(), st.PictureCard(), nil
}

func (p *stagePicturePublisher) UpdatePicture(ctx context.Context, id string, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	stageID, err := game.ParseStageID(id)
	if err != nil {
		return err
	}
	return p.repo.UpdatePicture(ctx, stageID, main, thumb, card, lqip, status)
}

func (p *stagePicturePublisher) SetMediaID(ctx context.Context, id string, mediaID string) error {
	stageID, err := game.ParseStageID(id)
	if err != nil {
		return err
	}
	return p.repo.SetMediaID(ctx, stageID, mediaID)
}

// characterEntity is the slice of a Character subtype's public surface
// characterPicturePublisher needs. Both characters.JojoCharacter and
// characters.OnePieceCharacter embed characters.Character, which declares
// all three methods, so both satisfy this.
type characterEntity interface {
	Picture() string
	PictureThumb() string
	PictureCard() string
}

// characterRepo is the slice of a Character repository
// characterPicturePublisher needs, satisfied structurally by both
// ports.IJojoCharacterRepository and ports.IOnePieceCharacterRepository -
// unlike Stand/DevilFruit (routed by enums.PowerKind through two entirely
// separate publisher types), the two Character kinds share one CharacterID
// space, so one generic publisher covers both.
type characterRepo[T characterEntity] interface {
	FindByID(ctx context.Context, id characters.CharacterID, locale enums.Locale) (T, error)
	UpdatePicture(ctx context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error
	SetMediaID(ctx context.Context, id characters.CharacterID, mediaID string) error
}

// characterPicturePublisher adapts a characterRepo[T] to PicturePublisher.
type characterPicturePublisher[T characterEntity] struct {
	repo characterRepo[T]
}

// NewJojoCharacterPicturePublisher wraps repo so the picture worker can
// publish transcoded renditions onto JojoCharacters.
func NewJojoCharacterPicturePublisher(repo ports.IJojoCharacterRepository) PicturePublisher {
	return &characterPicturePublisher[*characters.JojoCharacter]{repo: repo}
}

// NewOnePieceCharacterPicturePublisher wraps repo so the picture worker can
// publish transcoded renditions onto OnePieceCharacters.
func NewOnePieceCharacterPicturePublisher(repo ports.IOnePieceCharacterRepository) PicturePublisher {
	return &characterPicturePublisher[*characters.OnePieceCharacter]{repo: repo}
}

func (p *characterPicturePublisher[T]) PictureKeys(ctx context.Context, id string) (string, string, string, error) {
	characterID, err := characters.ParseCharacterID(id)
	if err != nil {
		return "", "", "", err
	}
	c, err := p.repo.FindByID(ctx, characterID, enums.EnGB)
	if err != nil {
		return "", "", "", err
	}
	return c.Picture(), c.PictureThumb(), c.PictureCard(), nil
}

func (p *characterPicturePublisher[T]) UpdatePicture(ctx context.Context, id string, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	characterID, err := characters.ParseCharacterID(id)
	if err != nil {
		return err
	}
	return p.repo.UpdatePicture(ctx, characterID, main, thumb, card, lqip, status)
}

func (p *characterPicturePublisher[T]) SetMediaID(ctx context.Context, id string, mediaID string) error {
	characterID, err := characters.ParseCharacterID(id)
	if err != nil {
		return err
	}
	return p.repo.SetMediaID(ctx, characterID, mediaID)
}
