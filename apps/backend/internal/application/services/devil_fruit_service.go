package services

import (
	"context"
	"io"
	"log"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// DevilFruitInput carries every field needed to create or update a
// DevilFruit, so CreateDevilFruit/UpdateDevilFruit take one argument instead
// of a long positional list.
type DevilFruitInput struct {
	Name          string
	Description   string
	Rarity        enums.PowerRarity
	Skills        *[]string
	Picture       string
	PictureThumb  string
	PictureStatus enums.PictureStatus
	FruitType     enums.FruitType
}

// DevilFruitService coordinates DevilFruit use cases against the injected
// repository. It reuses the sentinels and PicturePolicy declared in
// stand_service.go - both catalogues share the same picture pipeline rules.
type DevilFruitService struct {
	repo      ports.IDevilFruitRepository
	idGen     ports.IIdGenerator[powers.PowerID]
	pictures  ports.IPictureStorage
	processor ports.IImageProcessor
	enqueuer  ports.IPictureEnqueuer
	picPolicy PicturePolicy
}

func NewDevilFruitService(
	repo ports.IDevilFruitRepository,
	idGen ports.IIdGenerator[powers.PowerID],
	pictures ports.IPictureStorage,
	processor ports.IImageProcessor,
	enqueuer ports.IPictureEnqueuer,
	picPolicy PicturePolicy,
) *DevilFruitService {
	return &DevilFruitService{
		repo: repo, idGen: idGen, pictures: pictures,
		processor: processor, enqueuer: enqueuer, picPolicy: picPolicy,
	}
}

// CreateDevilFruit builds a new DevilFruit with a freshly generated id and
// persists it.
func (s *DevilFruitService) CreateDevilFruit(ctx context.Context, input DevilFruitInput) (*powers.DevilFruit, error) {
	return s.saveDevilFruit(ctx, s.idGen.NewID(), input)
}

// UpdateDevilFruit rebuilds the DevilFruit identified by id with the given
// fields and persists it, keeping its original id and its picture (set
// separately via SetDevilFruitPicture, not through this JSON body).
func (s *DevilFruitService) UpdateDevilFruit(ctx context.Context, id powers.PowerID, input DevilFruitInput) (*powers.DevilFruit, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	input.Picture = existing.Picture()
	input.PictureThumb = existing.PictureThumb()
	input.PictureStatus = existing.PictureStatus()
	return s.saveDevilFruit(ctx, id, input)
}

func (s *DevilFruitService) saveDevilFruit(ctx context.Context, id powers.PowerID, input DevilFruitInput) (*powers.DevilFruit, error) {
	power, err := powers.NewPower(id, input.Name, input.Description, input.Rarity, input.Skills, input.Picture)
	if err != nil {
		return nil, err
	}
	power.SetPictureRenditions(input.Picture, input.PictureThumb, input.PictureStatus)

	fruit, err := powers.NewDevilFruit(*power, input.FruitType)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, fruit); err != nil {
		return nil, err
	}
	return fruit, nil
}

// GetDevilFruit returns the devil fruit identified by id.
func (s *DevilFruitService) GetDevilFruit(ctx context.Context, id powers.PowerID) (*powers.DevilFruit, error) {
	return s.repo.FindByID(ctx, id)
}

// ListDevilFruits returns every devil fruit.
func (s *DevilFruitService) ListDevilFruits(ctx context.Context) ([]*powers.DevilFruit, error) {
	return s.repo.GetAll(ctx)
}

// FilterDevilFruits returns every devil fruit matching the given filters.
func (s *DevilFruitService) FilterDevilFruits(ctx context.Context, filters ports.DevilFruitFilters) ([]*powers.DevilFruit, error) {
	return s.repo.Filter(ctx, filters)
}

// DeleteDevilFruit removes the devil fruit identified by id.
func (s *DevilFruitService) DeleteDevilFruit(ctx context.Context, id powers.PowerID) error {
	return s.repo.Delete(ctx, id)
}

// SetDevilFruitPicture validates an uploaded picture and hands it to the
// background compression worker, moving the DevilFruit's picture pipeline to
// PENDING without touching its currently-served renditions. The returned
// DevilFruit still carries the previous picture/thumbnail keys (or none, on
// a first upload) - the worker publishes the new ones once it finishes.
func (s *DevilFruitService) SetDevilFruitPicture(ctx context.Context, id powers.PowerID, pic ports.Picture) (*powers.DevilFruit, error) {
	fruit, err := s.repo.FindByID(ctx, id)
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
	previousMain, previousThumb, previousStatus := fruit.Picture(), fruit.PictureThumb(), fruit.PictureStatus()

	if err := s.repo.UpdatePicture(ctx, id, nil, nil, enums.PicturePending); err != nil {
		return nil, err
	}

	if err := s.enqueuer.Enqueue(ports.PictureJob{SubjectID: id.String(), Kind: enums.DevilFruitSubject, Content: buf, ContentType: pic.ContentType}); err != nil {
		if revertErr := s.repo.UpdatePicture(ctx, id, nil, nil, previousStatus); revertErr != nil {
			log.Printf("reverting picture status for devil fruit %s after enqueue failure: %v", id, revertErr)
		}
		return nil, err
	}

	fruit.SetPictureRenditions(previousMain, previousThumb, enums.PicturePending)
	return fruit, nil
}

// PictureURL resolves a DevilFruit's stored picture key into a URL a client
// can GET, or "" if the DevilFruit has no picture.
func (s *DevilFruitService) PictureURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return s.pictures.PresignGetURL(ctx, key)
}

// MaxPictureBytes exposes the configured picture size limit so the HTTP
// handler can size its request-body guard without importing config.
func (s *DevilFruitService) MaxPictureBytes() int64 {
	return s.picPolicy.MaxBytes
}
