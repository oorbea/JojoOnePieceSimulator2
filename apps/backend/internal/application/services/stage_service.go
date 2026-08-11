package services

import (
	"context"
	"io"
	"log"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// StageInput is the mutable content of a Stage - everything NewStage needs
// besides its ID and picture (set separately via SetStagePicture, not
// through this input). Translations must always include all three
// enums.Locales() - callers validate this before it reaches the service
// (see dto.StageRequest.Validate).
type StageInput struct {
	Manga        enums.Manga
	Order        int
	Name         string
	Translations ports.StageTranslations
}

// StageService is the admin-facing CRUD service over ports.IStageRepository,
// the counterpart to StandService/DevilFruitService for the game feature's
// round content - same picture pipeline (background WebP transcoding via
// the shared PictureWorker), same translated-description pattern as
// power_translations (minus Skills, which a Stage has none of).
type StageService struct {
	repo      ports.IStageRepository
	ids       ports.IIdGenerator[game.StageID]
	pictures  ports.IPictureStorage
	processor ports.IImageProcessor
	enqueuer  ports.IPictureEnqueuer
	picPolicy PicturePolicy
}

// NewStageService builds a StageService. Every read/write goes through
// ports.IStageRepository - the admin CRUD API has no need for the
// gameplay-facing ports.IStageCatalog (see that port's doc on why it's
// locale-free and EnGB-fixed).
func NewStageService(
	repo ports.IStageRepository,
	ids ports.IIdGenerator[game.StageID],
	pictures ports.IPictureStorage,
	processor ports.IImageProcessor,
	enqueuer ports.IPictureEnqueuer,
	picPolicy PicturePolicy,
) *StageService {
	return &StageService{
		repo: repo, ids: ids, pictures: pictures,
		processor: processor, enqueuer: enqueuer, picPolicy: picPolicy,
	}
}

// ListStages returns every Stage, ordered by manga then position then name,
// description resolved for locale.
func (s *StageService) ListStages(ctx context.Context, locale enums.Locale) ([]game.Stage, error) {
	return s.repo.List(ctx, locale)
}

// StagesByManga returns every Stage for manga, ordered by position,
// description resolved for locale. Admin-facing (locale-aware) - unrelated
// to IStageCatalog.Stages, which the gameplay engine uses and which always
// resolves at a fixed enums.EnGB (see that port's doc).
func (s *StageService) StagesByManga(ctx context.Context, manga enums.Manga, locale enums.Locale) ([]game.Stage, error) {
	return s.repo.ListByManga(ctx, manga, locale)
}

// GetStage returns the Stage matching id, description resolved for locale,
// or ports.ErrStageNotFound.
func (s *StageService) GetStage(ctx context.Context, id game.StageID, locale enums.Locale) (game.Stage, error) {
	return s.repo.FindByID(ctx, id, locale)
}

// CreateStage builds and persists a new Stage with a freshly generated ID.
func (s *StageService) CreateStage(ctx context.Context, input StageInput) (game.Stage, error) {
	return s.saveStage(ctx, s.ids.NewID(), "", "", enums.PictureNone, input)
}

// UpdateStage replaces the content of the Stage matching id, keeping its ID
// and its picture (set separately via SetStagePicture, not through this
// JSON body).
func (s *StageService) UpdateStage(ctx context.Context, id game.StageID, input StageInput) (game.Stage, error) {
	existing, err := s.repo.FindByID(ctx, id, enums.EnGB)
	if err != nil {
		return game.Stage{}, err
	}
	return s.saveStage(ctx, id, existing.Picture(), existing.PictureThumb(), existing.PictureStatus(), input)
}

func (s *StageService) saveStage(ctx context.Context, id game.StageID, picture, pictureThumb string, pictureStatus enums.PictureStatus, input StageInput) (game.Stage, error) {
	description := input.Translations[enums.EnGB]
	st, err := game.NewStage(id, input.Manga, input.Order, input.Name, description, picture)
	if err != nil {
		return game.Stage{}, err
	}
	st.SetPictureRenditions(picture, pictureThumb, pictureStatus)

	if err := s.repo.Save(ctx, st, input.Translations); err != nil {
		return game.Stage{}, err
	}
	return st, nil
}

// DeleteStage removes the Stage matching id, then best-effort deletes its
// picture renditions from object storage - mirrors StandService.DeleteStand.
func (s *StageService) DeleteStage(ctx context.Context, id game.StageID) error {
	st, err := s.repo.FindByID(ctx, id, enums.EnGB)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if key := st.Picture(); key != "" {
		if err := s.pictures.Delete(ctx, key); err != nil {
			log.Printf("deleting picture %q for stage %s: %v", key, id, err)
		}
	}
	if key := st.PictureThumb(); key != "" {
		if err := s.pictures.Delete(ctx, key); err != nil {
			log.Printf("deleting picture thumbnail %q for stage %s: %v", key, id, err)
		}
	}
	return nil
}

// StageTranslations returns every locale's content for id, for the admin
// edit form.
func (s *StageService) StageTranslations(ctx context.Context, id game.StageID) (ports.StageTranslations, error) {
	return s.repo.Translations(ctx, id)
}

// SetStagePicture validates an uploaded picture and hands it to the
// background compression worker, moving the Stage's picture pipeline to
// PENDING without touching its currently-served renditions - identical
// contract to StandService.SetStandPicture.
func (s *StageService) SetStagePicture(ctx context.Context, id game.StageID, pic ports.Picture) (game.Stage, error) {
	st, err := s.repo.FindByID(ctx, id, enums.EnGB)
	if err != nil {
		return game.Stage{}, err
	}

	if s.picPolicy.MaxBytes > 0 && pic.Size > s.picPolicy.MaxBytes {
		return game.Stage{}, ErrPictureTooLarge
	}
	if !s.picPolicy.allows(pic.ContentType) {
		return game.Stage{}, ErrUnsupportedPictureType
	}

	buf, err := io.ReadAll(pic.Content)
	if err != nil {
		return game.Stage{}, err
	}

	meta, err := s.processor.Probe(buf)
	if err != nil {
		return game.Stage{}, err
	}
	if s.picPolicy.MaxPixels > 0 && int64(meta.Width)*int64(meta.Height)*int64(max(meta.Pages, 1)) > s.picPolicy.MaxPixels {
		return game.Stage{}, ports.ErrInvalidImage
	}

	previousMain, previousThumb, previousStatus := st.Picture(), st.PictureThumb(), st.PictureStatus()

	if err := s.repo.UpdatePicture(ctx, id, nil, nil, enums.PicturePending); err != nil {
		return game.Stage{}, err
	}

	if err := s.enqueuer.Enqueue(ports.PictureJob{SubjectID: id.String(), Kind: enums.StageSubject, Content: buf, ContentType: pic.ContentType}); err != nil {
		if revertErr := s.repo.UpdatePicture(ctx, id, nil, nil, previousStatus); revertErr != nil {
			log.Printf("reverting picture status for stage %s after enqueue failure: %v", id, revertErr)
		}
		return game.Stage{}, err
	}

	st.SetPictureRenditions(previousMain, previousThumb, enums.PicturePending)
	return st, nil
}

// PictureURL resolves a Stage's stored picture key into a URL a client can
// GET, or "" if the Stage has no picture.
func (s *StageService) PictureURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return s.pictures.PresignGetURL(ctx, key)
}

// MaxPictureBytes exposes the configured picture size limit so the HTTP
// handler can size its request-body guard without importing config.
func (s *StageService) MaxPictureBytes() int64 {
	return s.picPolicy.MaxBytes
}
