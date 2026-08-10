package game

import (
	"errors"
	"sort"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// Stage is a single unit of round content - a JoJo part or a One Piece
// saga. It is deliberately opaque beyond Manga and Order: the actual
// catalog (names, descriptions, admin CRUD) is served by
// ports.IStageCatalog and lives outside the domain layer.
type Stage struct {
	id    StageID
	manga enums.Manga
	order int
	name  string
}

// NewStage validates and builds a Stage.
func NewStage(id StageID, manga enums.Manga, order int, name string) (Stage, error) {
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
	return Stage{id: id, manga: manga, order: order, name: name}, nil
}

func (s Stage) ID() StageID        { return s.id }
func (s Stage) Manga() enums.Manga { return s.manga }
func (s Stage) Order() int         { return s.order }
func (s Stage) Name() string       { return s.name }

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
