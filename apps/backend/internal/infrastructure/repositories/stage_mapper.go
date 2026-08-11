package repositories

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// stageRow is the minimal shape every stages.sql.go row type shares -
// ListStages/FilterStageRows/GetStageByID/UpsertStage all return the same
// columns under different generated row types, so this lets a single
// mapper function serve all of them via a small per-caller adapter.
type stageRow struct {
	ID            pgtype.UUID
	Manga         string
	Position      int32
	Name          string
	Description   string
	Picture       string
	PictureThumb  string
	PictureStatus string
}

func toStage(r stageRow) (game.Stage, error) {
	manga, err := enums.ParseManga(r.Manga)
	if err != nil {
		return game.Stage{}, fmt.Errorf("stage %q: %w", r.Name, err)
	}
	status, err := enums.ParsePictureStatus(r.PictureStatus)
	if err != nil {
		return game.Stage{}, fmt.Errorf("stage %q: picture_status: %w", r.Name, err)
	}
	st, err := game.NewStage(r.ID.Bytes, manga, int(r.Position), r.Name, r.Description, r.Picture)
	if err != nil {
		return game.Stage{}, err
	}
	st.SetPictureRenditions(r.Picture, r.PictureThumb, status)
	return st, nil
}

func fromListStagesRow(r db.ListStagesRow) stageRow {
	return stageRow{
		ID: r.ID, Manga: r.Manga, Position: r.Position, Name: r.Name, Description: r.Description,
		Picture: r.Picture, PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus,
	}
}

func fromFilterStageRow(r db.FilterStageRowsRow) stageRow {
	return stageRow{
		ID: r.ID, Manga: r.Manga, Position: r.Position, Name: r.Name, Description: r.Description,
		Picture: r.Picture, PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus,
	}
}

func fromGetStageByIDRow(r db.GetStageByIDRow) stageRow {
	return stageRow{
		ID: r.ID, Manga: r.Manga, Position: r.Position, Name: r.Name, Description: r.Description,
		Picture: r.Picture, PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus,
	}
}
