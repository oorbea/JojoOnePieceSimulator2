package services

import (
	"context"
	"errors"
	"io"
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

// PicturePolicy bounds what SetStandPicture accepts, kept independent from
// object-storage concerns.
type PicturePolicy struct {
	MaxBytes     int64
	AllowedTypes []string
	// MaxPixels bounds width*height*pages, checked against the cheap header
	// probe before any decode - the guard against a small file declaring a
	// huge decoded size ("decompression bomb"). 0 disables the check.
	MaxPixels int64
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
	processor ports.IImageProcessor
	enqueuer  ports.IPictureEnqueuer
	picPolicy PicturePolicy
}

func NewStandService(
	standRepo ports.IStandRepository,
	idGen ports.IIdGenerator[powers.PowerID],
	pictures ports.IPictureStorage,
	processor ports.IImageProcessor,
	enqueuer ports.IPictureEnqueuer,
	picPolicy PicturePolicy,
) *StandService {
	return &StandService{
		standRepo: standRepo, idGen: idGen, pictures: pictures,
		processor: processor, enqueuer: enqueuer, picPolicy: picPolicy,
	}
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

// SetStandPicture validates an uploaded picture and hands it to the
// background compression worker, moving the Stand's picture pipeline to
// PENDING without touching its currently-served renditions. The returned
// Stand still carries the previous picture/thumbnail keys (or none, on a
// first upload) - the worker publishes the new ones once it finishes.
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

	// Captured before touching the repo or the worker: once Enqueue returns,
	// the worker may already have run (a synchronous or very fast
	// implementation) and mutated the persisted renditions, so stand's own
	// getters can no longer be trusted to still reflect the pre-upload
	// state.
	previousMain, previousThumb, previousStatus := stand.Picture(), stand.PictureThumb(), stand.PictureStatus()

	if err := s.standRepo.UpdatePicture(ctx, id, nil, nil, enums.PicturePending); err != nil {
		return nil, err
	}

	if err := s.enqueuer.Enqueue(ports.PictureJob{PowerID: id, Kind: enums.StandKind, Content: buf, ContentType: pic.ContentType}); err != nil {
		if revertErr := s.standRepo.UpdatePicture(ctx, id, nil, nil, previousStatus); revertErr != nil {
			log.Printf("reverting picture status for stand %s after enqueue failure: %v", id, revertErr)
		}
		return nil, err
	}

	stand.SetPictureRenditions(previousMain, previousThumb, enums.PicturePending)
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
