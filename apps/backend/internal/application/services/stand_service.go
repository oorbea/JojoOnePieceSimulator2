package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// ErrPictureQueueFull is returned when the background compression worker's
// queue has no room for another job.
var ErrPictureQueueFull = errors.New("picture processing queue is full")

// ErrSelfEvolution is returned when a Stand is asked to evolve from itself.
var ErrSelfEvolution = errors.New("a stand cannot evolve from itself")

// ErrPictureRequired is returned when PATCH /stands/{id}/picture is called
// without a "picture" form file.
var ErrPictureRequired = errors.New("picture file is required")

// ErrUnsupportedPictureType is returned when an uploaded picture's sniffed
// content type is not in the configured allow-list.
var ErrUnsupportedPictureType = errors.New("unsupported picture content type")

// ErrPictureTooLarge is returned when an uploaded picture exceeds the
// configured max size.
var ErrPictureTooLarge = errors.New("picture exceeds the maximum allowed size")

// pictureExtensions maps a sniffed content type to the file extension used
// in the object key, so keys stay human-readable in the bucket.
var pictureExtensions = map[string]string{
	"image/webp": ".webp",
	"image/avif": ".avif",
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
}

// PicturePolicy bounds what SetStandPicture accepts, kept independent from
// object-storage concerns.
type PicturePolicy struct {
	MaxBytes     int64
	AllowedTypes []string
}

func (p PicturePolicy) allows(contentType string) bool {
	for _, allowed := range p.AllowedTypes {
		if allowed == contentType {
			return true
		}
	}
	return false
}

// StandInput carries every field needed to create or update a Stand, so
// CreateStand/UpdateStand take one argument instead of a long positional
// list.
type StandInput struct {
	Name          string
	Description   string
	Rarity        enums.PowerRarity
	Skills        *[]string
	Picture       string
	PictureThumb  string
	PictureStatus enums.PictureStatus
	AttackPower   enums.StandStat
	Speed         enums.StandStat
	AttackRange   enums.StandStat
	Endurance     enums.StandStat
	Precision     enums.StandStat
	Potential     enums.StandStat
	EvolvesFrom   *powers.PowerID
}

// StandService coordinates Stand use cases against the injected repository.
type StandService struct {
	standRepo ports.IStandRepository
	idGen     ports.IIdGenerator[powers.PowerID]
	pictures  ports.IPictureStorage
	picPolicy PicturePolicy
}

func NewStandService(
	standRepo ports.IStandRepository,
	idGen ports.IIdGenerator[powers.PowerID],
	pictures ports.IPictureStorage,
	picPolicy PicturePolicy,
) *StandService {
	return &StandService{standRepo: standRepo, idGen: idGen, pictures: pictures, picPolicy: picPolicy}
}

// CreateStand builds a new Stand with a freshly generated id and persists it.
func (s *StandService) CreateStand(ctx context.Context, input StandInput) (*powers.Stand, error) {
	return s.saveStand(ctx, s.idGen.NewID(), input)
}

// UpdateStand rebuilds the Stand identified by id with the given fields and
// persists it, keeping its original id and its picture (set separately via
// SetStandPicture, not through this JSON body).
func (s *StandService) UpdateStand(ctx context.Context, id powers.PowerID, input StandInput) (*powers.Stand, error) {
	existing, err := s.standRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.EvolvesFrom != nil && *input.EvolvesFrom == id {
		return nil, ErrSelfEvolution
	}
	input.Picture = existing.Picture()
	input.PictureThumb = existing.PictureThumb()
	input.PictureStatus = existing.PictureStatus()
	return s.saveStand(ctx, id, input)
}

func (s *StandService) saveStand(ctx context.Context, id powers.PowerID, input StandInput) (*powers.Stand, error) {
	power, err := powers.NewPower(id, input.Name, input.Description, input.Rarity, input.Skills, input.Picture)
	if err != nil {
		return nil, err
	}
	power.SetPictureRenditions(input.Picture, input.PictureThumb, input.PictureStatus)

	var evolvesFromStand *powers.Stand
	if input.EvolvesFrom != nil {
		evolvesFromStand, err = s.standRepo.FindByID(ctx, *input.EvolvesFrom)
		if err != nil {
			return nil, err
		}
	}

	stand, err := powers.NewStand(*power, input.AttackPower, input.Speed, input.AttackRange, input.Endurance, input.Precision, input.Potential, evolvesFromStand)
	if err != nil {
		return nil, err
	}

	if err := s.standRepo.Save(ctx, stand); err != nil {
		return nil, err
	}
	return stand, nil
}

// GetStand returns the stand identified by id.
func (s *StandService) GetStand(ctx context.Context, id powers.PowerID) (*powers.Stand, error) {
	return s.standRepo.FindByID(ctx, id)
}

// ListStands returns every stand.
func (s *StandService) ListStands(ctx context.Context) ([]*powers.Stand, error) {
	return s.standRepo.GetAll(ctx)
}

// FilterStands returns every stand matching the given filters.
func (s *StandService) FilterStands(ctx context.Context, filters ports.StandFilters) ([]*powers.Stand, error) {
	return s.standRepo.Filter(ctx, filters)
}

// DeleteStand removes the stand identified by id.
func (s *StandService) DeleteStand(ctx context.Context, id powers.PowerID) error {
	return s.standRepo.Delete(ctx, id)
}

// SetStandPicture validates and uploads pic to object storage, then persists
// its resulting key on the Stand identified by id. A previously-stored
// picture, if any, is deleted best-effort - a failure there is logged but
// does not fail the request, since the new picture is already saved.
func (s *StandService) SetStandPicture(ctx context.Context, id powers.PowerID, pic ports.Picture) (*powers.Stand, error) {
	stand, err := s.standRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.picPolicy.MaxBytes > 0 && pic.Size > s.picPolicy.MaxBytes {
		return nil, ErrPictureTooLarge
	}
	if !s.picPolicy.allows(pic.ContentType) {
		return nil, ErrUnsupportedPictureType
	}

	oldKey := stand.Picture()
	key := fmt.Sprintf("stands/%s/%s%s", id, s.idGen.NewID(), pictureExtensions[pic.ContentType])

	if err := s.pictures.Upload(ctx, key, pic); err != nil {
		return nil, err
	}

	stand.SetPicture(key)
	if err := s.standRepo.Save(ctx, stand); err != nil {
		return nil, err
	}

	if oldKey != "" && oldKey != key {
		if err := s.pictures.Delete(ctx, oldKey); err != nil {
			log.Printf("deleting old picture %q: %v", oldKey, err)
		}
	}

	return stand, nil
}

// PictureURL resolves a Stand's stored picture key into a URL a client can
// GET, or "" if the Stand has no picture.
func (s *StandService) PictureURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return s.pictures.PresignGetURL(ctx, key)
}

// MaxPictureBytes exposes the configured picture size limit so the HTTP
// handler can size its request-body guard without importing config.
func (s *StandService) MaxPictureBytes() int64 {
	return s.picPolicy.MaxBytes
}
