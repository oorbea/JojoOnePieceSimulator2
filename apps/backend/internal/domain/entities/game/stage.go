package game

import (
	"errors"
	"sort"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// Stage is a single unit of round content - a JoJo part or a One Piece
// saga. Name/Manga/Order are what game.Interleave/IGameMode actually use;
// Description/Picture are display metadata, resolved the same way
// powers.Power resolves its own (a single already-resolved value per
// fetch - see ports.IStageRepository/IStageCatalog and, for how a live
// match re-resolves it per viewer's own configured language instead of
// whatever was baked in at round-assignment time,
// api/dto.NewGameStateResponse).
type Stage struct {
	name          string
	description   string
	picture       string
	pictureThumb  string
	order         int
	id            StageID
	manga         enums.Manga
	pictureStatus enums.PictureStatus
}

// NewStage validates and builds a Stage. picture is the only picture field
// accepted here - thumb/status start empty/NONE, same convention as
// powers.NewPower; set them together via SetPictureRenditions.
func NewStage(id StageID, manga enums.Manga, order int, name string, description string, picture string) (Stage, error) {
	if id.IsNil() {
		return Stage{}, errors.New("id is required")
	}
	if !manga.IsValid() {
		return Stage{}, enums.ErrInvalidManga
	}
	if order < 0 {
		return Stage{}, errors.New("order must be non-negative")
	}
	if name == "" {
		return Stage{}, errors.New("name is required")
	}
	if description == "" {
		return Stage{}, errors.New("description is required")
	}
	return Stage{id: id, manga: manga, order: order, name: name, description: description, picture: picture}, nil
}

func (s Stage) ID() StageID          { return s.id }
func (s Stage) Manga() enums.Manga   { return s.manga }
func (s Stage) Order() int           { return s.order }
func (s Stage) Name() string         { return s.name }
func (s Stage) Description() string  { return s.description }
func (s Stage) Picture() string      { return s.picture }
func (s Stage) PictureThumb() string { return s.pictureThumb }

// PictureStatus reports where this Stage's picture is in the async
// compression pipeline.
func (s Stage) PictureStatus() enums.PictureStatus { return s.pictureStatus }

// SetPictureRenditions replaces the stored main and thumbnail picture keys
// together with the pipeline status that produced them, so the three
// always change as one unit - same pattern as powers.Power.SetPictureRenditions.
func (s *Stage) SetPictureRenditions(main, thumb string, status enums.PictureStatus) {
	s.picture = main
	s.pictureThumb = thumb
	s.pictureStatus = status
}

// Interleave merges per-manga Stage lists into Gauntlet round order: the
// first Stage of each manga (in enums.Mangas() order), then the second of
// each, and so on - leftovers from the longer manga(s) land at the end
// automatically, since a shorter manga simply stops contributing once its
// slice is exhausted. Each manga's slice is sorted by Order first, so the
// caller does not need to pre-sort ports.IStageCatalog's results.
func Interleave(byManga map[enums.Manga][]Stage) []Stage {
	mangas := enums.Mangas()
	sorted := make(map[enums.Manga][]Stage, len(mangas))
	maxLen := 0
	for _, m := range mangas {
		stages := append([]Stage(nil), byManga[m]...)
		sort.Slice(stages, func(i, j int) bool { return stages[i].order < stages[j].order })
		sorted[m] = stages
		if len(stages) > maxLen {
			maxLen = len(stages)
		}
	}

	result := make([]Stage, 0, maxLen*len(mangas))
	for i := 0; i < maxLen; i++ {
		for _, m := range mangas {
			if i < len(sorted[m]) {
				result = append(result, sorted[m][i])
			}
		}
	}
	return result
}
