package services

import (
	"context"
	"io"
	"log"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// OnePieceCharacterInput carries every field needed to create or update an
// OnePieceCharacter - same shape as JojoCharacterInput.
type OnePieceCharacterInput struct {
	Name            string
	Translations    ports.CharacterTranslations
	Picture         string
	PictureThumb    string
	PictureCard     string
	Rarity          enums.PowerRarity
	PictureStatus   enums.PictureStatus
	PictureLqip     string
	PhysicalForm    enums.PhysicalForm
	ArmamentHaki    enums.HakiLevel
	ObservationHaki enums.HakiLevel
	ConquerorHaki   enums.HakiLevel
	FruitMastery    enums.FruitMastery
}

// OnePieceCharacterService coordinates OnePieceCharacter use cases against
// the injected repository - same shape as JojoCharacterService.
type OnePieceCharacterService struct {
	repo      ports.IOnePieceCharacterRepository
	idGen     ports.IIdGenerator[characters.CharacterID]
	pictures  ports.IPictureStorage
	processor ports.IImageProcessor
	enqueuer  ports.IPictureEnqueuer
	picPolicy PicturePolicy
}

func NewOnePieceCharacterService(
	repo ports.IOnePieceCharacterRepository,
	idGen ports.IIdGenerator[characters.CharacterID],
	pictures ports.IPictureStorage,
	processor ports.IImageProcessor,
	enqueuer ports.IPictureEnqueuer,
	picPolicy PicturePolicy,
) *OnePieceCharacterService {
	return &OnePieceCharacterService{
		repo: repo, idGen: idGen, pictures: pictures,
		processor: processor, enqueuer: enqueuer, picPolicy: picPolicy,
	}
}

func (s *OnePieceCharacterService) CreateOnePieceCharacter(ctx context.Context, input OnePieceCharacterInput) (*characters.OnePieceCharacter, error) {
	return s.saveOnePieceCharacter(ctx, s.idGen.NewID(), input)
}

func (s *OnePieceCharacterService) UpdateOnePieceCharacter(ctx context.Context, id characters.CharacterID, input OnePieceCharacterInput) (*characters.OnePieceCharacter, error) {
	existing, err := s.repo.FindByID(ctx, id, enums.EnGB)
	if err != nil {
		return nil, err
	}
	input.Picture = existing.Picture()
	input.PictureThumb = existing.PictureThumb()
	input.PictureCard = existing.PictureCard()
	input.PictureStatus = existing.PictureStatus()
	input.PictureLqip = existing.PictureLqip()
	return s.saveOnePieceCharacter(ctx, id, input)
}

func (s *OnePieceCharacterService) saveOnePieceCharacter(ctx context.Context, id characters.CharacterID, input OnePieceCharacterInput) (*characters.OnePieceCharacter, error) {
	description := input.Translations[enums.EnGB]
	character, err := characters.NewCharacter(id, enums.OnePiece, input.Name, input.Rarity, description, input.Picture)
	if err != nil {
		return nil, err
	}
	character.SetPictureRenditions(input.Picture, input.PictureThumb, input.PictureCard, input.PictureLqip, input.PictureStatus)

	c, err := characters.NewOnePieceCharacter(character, input.PhysicalForm, input.ArmamentHaki, input.ObservationHaki, input.ConquerorHaki, input.FruitMastery)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, c, input.Translations); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *OnePieceCharacterService) GetOnePieceCharacter(ctx context.Context, id characters.CharacterID, locale enums.Locale) (*characters.OnePieceCharacter, error) {
	return s.repo.FindByID(ctx, id, locale)
}

func (s *OnePieceCharacterService) ListOnePieceCharacters(ctx context.Context, locale enums.Locale) ([]*characters.OnePieceCharacter, error) {
	return s.repo.GetAll(ctx, locale)
}

func (s *OnePieceCharacterService) FilterOnePieceCharacters(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale) ([]*characters.OnePieceCharacter, error) {
	return s.repo.Filter(ctx, filters, locale)
}

func (s *OnePieceCharacterService) PageOnePieceCharacters(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.OnePieceCharacter, bool, error) {
	return s.repo.Page(ctx, filters, locale, afterName, limit)
}

func (s *OnePieceCharacterService) CountOnePieceCharacters(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale) (int, error) {
	return s.repo.Count(ctx, filters, locale)
}

func (s *OnePieceCharacterService) OnePieceCharacterTranslations(ctx context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	return s.repo.Translations(ctx, id)
}

func (s *OnePieceCharacterService) DeleteOnePieceCharacter(ctx context.Context, id characters.CharacterID) error {
	c, err := s.repo.FindByID(ctx, id, enums.EnGB)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if key := c.Picture(); key != "" {
		if err := s.pictures.Delete(ctx, key); err != nil {
			log.Printf("deleting picture %q for one piece character %s: %v", key, id, err)
		}
	}
	if key := c.PictureThumb(); key != "" {
		if err := s.pictures.Delete(ctx, key); err != nil {
			log.Printf("deleting picture thumbnail %q for one piece character %s: %v", key, id, err)
		}
	}
	return nil
}

func (s *OnePieceCharacterService) SetOnePieceCharacterPicture(ctx context.Context, id characters.CharacterID, pic ports.Picture) (*characters.OnePieceCharacter, error) {
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

	if err := s.enqueuer.Enqueue(ports.PictureJob{SubjectID: id.String(), Kind: enums.OnePieceCharacterSubject, Content: buf, ContentType: pic.ContentType}); err != nil {
		if revertErr := s.repo.UpdatePicture(ctx, id, nil, nil, nil, nil, previousStatus); revertErr != nil {
			log.Printf("reverting picture status for one piece character %s after enqueue failure: %v", id, revertErr)
		}
		return nil, err
	}

	c.SetPictureRenditions(previousMain, previousThumb, previousCard, previousLqip, enums.PicturePending)
	return c, nil
}

func (s *OnePieceCharacterService) PictureURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return s.pictures.PresignGetURL(ctx, key)
}

func (s *OnePieceCharacterService) MaxPictureBytes() int64 {
	return s.picPolicy.MaxBytes
}
