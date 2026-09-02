package cache

import (
	"encoding/json"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// stageSnapshot is the JSON-serializable shape of a game.Stage. Stage's
// fields are all unexported and its snapshot helpers
// (game.snapshotStage/restoreStage) are package-private to the game
// package, so this package carries its own - deliberately local rather
// than exported from game or bolted onto infrastructure/powersnap, which
// is strictly about powers.
type stageSnapshot struct {
	ID            [16]byte `json:"id"`
	Manga         string   `json:"manga"`
	Order         int      `json:"order"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Picture       string   `json:"picture"`
	PictureThumb  string   `json:"pictureThumb"`
	PictureStatus string   `json:"pictureStatus"`
}

func ofStage(st game.Stage) stageSnapshot {
	return stageSnapshot{
		ID:            st.ID(),
		Manga:         st.Manga().String(),
		Order:         st.Order(),
		Name:          st.Name(),
		Description:   st.Description(),
		Picture:       st.Picture(),
		PictureThumb:  st.PictureThumb(),
		PictureStatus: st.PictureStatus().String(),
	}
}

// hydrate rebuilds a Stage in the same order game.restoreStage does: parse
// the manga, run the value through NewStage's invariants, then apply the
// picture renditions as one unit. Every failure is returned as an ordinary
// error - never a panic - because each call site treats an unmarshal error
// as a cache miss and falls through to the wrapped repository.
func (s stageSnapshot) hydrate() (game.Stage, error) {
	manga, err := enums.ParseManga(s.Manga)
	if err != nil {
		return game.Stage{}, fmt.Errorf("stage %q: %w", s.Name, err)
	}
	st, err := game.NewStage(s.ID, manga, s.Order, s.Name, s.Description, s.Picture)
	if err != nil {
		return game.Stage{}, fmt.Errorf("stage %q: %w", s.Name, err)
	}
	status, err := enums.ParsePictureStatus(s.PictureStatus)
	if err != nil {
		return game.Stage{}, fmt.Errorf("stage %q: picture_status: %w", s.Name, err)
	}
	st.SetPictureRenditions(s.Picture, s.PictureThumb, status)
	return st, nil
}

func marshalStage(st game.Stage) ([]byte, error) {
	return json.Marshal(ofStage(st))
}

func unmarshalStage(data []byte) (game.Stage, error) {
	var s stageSnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return game.Stage{}, fmt.Errorf("unmarshaling stage: %w", err)
	}
	return s.hydrate()
}

func marshalStages(sts []game.Stage) ([]byte, error) {
	snapshots := make([]stageSnapshot, len(sts))
	for i, st := range sts {
		snapshots[i] = ofStage(st)
	}
	return json.Marshal(snapshots)
}

func unmarshalStages(data []byte) ([]game.Stage, error) {
	var snapshots []stageSnapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, fmt.Errorf("unmarshaling stages: %w", err)
	}
	stages := make([]game.Stage, len(snapshots))
	for i, s := range snapshots {
		st, err := s.hydrate()
		if err != nil {
			return nil, err
		}
		stages[i] = st
	}
	return stages, nil
}
