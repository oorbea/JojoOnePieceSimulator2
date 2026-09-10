package services

import (
	"context"
	"io"
	"log"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// JojoCharacterInput carries every field needed to create or update a
// JojoCharacter, so CreateJojoCharacter/UpdateJojoCharacter take one
// argument instead of a long positional list - same shape as
// DevilFruitInput. Translations must always include enums.EnGB - callers
// validate this before it reaches the service.
type JojoCharacterInput struct {
	Name          string
	Translations  ports.CharacterTranslations
	Picture       string
	PictureThumb  string
	PictureCard   string
	Rarity        enums.PowerRarity
	PictureStatus enums.PictureStatus
	PictureLqip   string
	Hamon         enums.HamonLevel
	Spin          enums.SpinLevel
	BattleIQ      byte
}

// JojoCharacterService coordinates JojoCharacter use cases against the
// injected repository - same shape as DevilFruitService, reusing its
// PicturePolicy/sentinels (both catalogues share the same picture pipeline
// rules).
type JojoCharacterService struct {
	repo      ports.IJojoCharacterRepository
	idGen     ports.IIdGenerator[characters.CharacterID]
	pictures  ports.IPictureStorage
	processor ports.IImageProcessor
	enqueuer  ports.IPictureEnqueuer
	picPolicy PicturePolicy
}

func NewJojoCharacterService(
	repo ports.IJojoCharacterRepository,
	idGen ports.IIdGenerator[characters.CharacterID],
	pictures ports.IPictureStorage,
	processor ports.IImageProcessor,
	enqueuer ports.IPictureEnqueuer,
	picPolicy PicturePolicy,
) *JojoCharacterService {
	return &JojoCharacterService{
		repo: repo, idGen: idGen, pictures: pictures,
		processor: processor, enqueuer: enqueuer, picPolicy: picPolicy,
	}
}

// CreateJojoCharacter builds a new JojoCharacter with a freshly generated id
// and persists it.
func (s *JojoCharacterService) CreateJojoCharacter(ctx context.Context, input JojoCharacterInput) (*characters.JojoCharacter, error) {
	return s.saveJojoCharacter(ctx, s.idGen.NewID(), input)
}

// UpdateJojoCharacter rebuilds the JojoCharacter identified by id with the
// given fields and persists it, keeping its original id and its picture
// (set separately via SetJojoCharacterPicture, not through this JSON body).
func (s *JojoCharacterService) UpdateJojoCharacter(ctx context.Context, id characters.CharacterID, input JojoCharacterInput) (*characters.JojoCharacter, error) {
	existing, err := s.repo.FindByID(ctx, id, enums.EnGB)
	if err != nil {
		return nil, err
	}
	input.Picture = existing.Picture()
	input.PictureThumb = existing.PictureThumb()
	input.PictureCard = existing.PictureCard()
	input.PictureStatus = existing.PictureStatus()
	input.PictureLqip = existing.PictureLqip()
	return s.saveJojoCharacter(ctx, id, input)
}

func (s *JojoCharacterService) saveJojoCharacter(ctx context.Context, id characters.CharacterID, input JojoCharacterInput) (*characters.JojoCharacter, error) {
	description := input.Translations[enums.EnGB]
	character, err := characters.NewCharacter(id, enums.Jojo, input.Name, input.Rarity, description, input.Picture)
	if err != nil {
		return nil, err
	}
	character.SetPictureRenditions(input.Picture, input.PictureThumb, input.PictureCard, input.PictureLqip, input.PictureStatus)

	c, err := characters.NewJojoCharacter(character, input.Hamon, input.Spin, input.BattleIQ)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, c, input.Translations); err != nil {
		return nil, err
	}
	return c, nil
}

// GetJojoCharacter returns the character identified by id, description
// resolved for locale.
func (s *JojoCharacterService) GetJojoCharacter(ctx context.Context, id characters.CharacterID, locale enums.Locale) (*characters.JojoCharacter, error) {
	return s.repo.FindByID(ctx, id, locale)
}

// ListJojoCharacters returns every JoJo character, description resolved for
// locale.
func (s *JojoCharacterService) ListJojoCharacters(ctx context.Context, locale enums.Locale) ([]*characters.JojoCharacter, error) {
	return s.repo.GetAll(ctx, locale)
}

// FilterJojoCharacters returns every JoJo character matching the given
// filters, description resolved for locale.
func (s *JojoCharacterService) FilterJojoCharacters(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale) ([]*characters.JojoCharacter, error) {
	return s.repo.Filter(ctx, filters, locale)
}

// PageJojoCharacters returns up to limit+1 JoJo characters matching
// filters, ordered by name after afterName - see
// ports.IJojoCharacterRepository.Page's doc.
func (s *JojoCharacterService) PageJojoCharacters(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.JojoCharacter, bool, error) {
	return s.repo.Page(ctx, filters, locale, afterName, limit)
}

// CountJojoCharacters returns the total number of JoJo characters matching
// filters, ignoring pagination.
func (s *JojoCharacterService) CountJojoCharacters(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale) (int, error) {
	return s.repo.Count(ctx, filters, locale)
}

// JojoCharacterTranslations returns every locale's content for id, for the
// admin edit form.
func (s *JojoCharacterService) JojoCharacterTranslations(ctx context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	return s.repo.Translations(ctx, id)
}

// DeleteJojoCharacter removes the character identified by id, then
// best-effort deletes its picture renditions from object storage - see
// StandService.DeleteStand for why.
func (s *JojoCharacterService) DeleteJojoCharacter(ctx context.Context, id characters.CharacterID) error {
	c, err := s.repo.FindByID(ctx, id, enums.EnGB)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if key := c.Picture(); key != "" {
		if err := s.pictures.Delete(ctx, key); err != nil {
			log.Printf("deleting picture %q for jojo character %s: %v", key, id, err)
		}
	}
	if key := c.PictureThumb(); key != "" {
		if err := s.pictures.Delete(ctx, key); err != nil {
			log.Printf("deleting picture thumbnail %q for jojo character %s: %v", key, id, err)
		}
	}
	return nil
}

// SetJojoCharacterPicture validates an uploaded picture and hands it to the
// background compression worker - same shape as
// DevilFruitService.SetDevilFruitPicture.
func (s *JojoCharacterService) SetJojoCharacterPicture(ctx context.Context, id characters.CharacterID, pic ports.Picture) (*characters.JojoCharacter, error) {
	c, err := s.repo.FindByID(ctx, id, enums.EnGB)
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

	previousMain, previousThumb, previousCard, previousLqip, previousStatus := c.Picture(), c.PictureThumb(), c.PictureCard(), c.PictureLqip(), c.PictureStatus()

	if err := s.repo.UpdatePicture(ctx, id, nil, nil, nil, nil, enums.PicturePending); err != nil {
		return nil, err
	}

	if err := s.enqueuer.Enqueue(ports.PictureJob{SubjectID: id.String(), Kind: enums.JojoCharacterSubject, Content: buf, ContentType: pic.ContentType}); err != nil {
		if revertErr := s.repo.UpdatePicture(ctx, id, nil, nil, nil, nil, previousStatus); revertErr != nil {
			log.Printf("reverting picture status for jojo character %s after enqueue failure: %v", id, revertErr)
		}
		return nil, err
	}

	c.SetPictureRenditions(previousMain, previousThumb, previousCard, previousLqip, enums.PicturePending)
	return c, nil
}

// PictureURL resolves a JojoCharacter's stored picture key into a URL a
// client can GET, or "" if it has no picture.
func (s *JojoCharacterService) PictureURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return s.pictures.PresignGetURL(ctx, key)
}

// MaxPictureBytes exposes the configured picture size limit so the HTTP
// handler can size its request-body guard without importing config.
func (s *JojoCharacterService) MaxPictureBytes() int64 {
	return s.picPolicy.MaxBytes
}
