package repositories

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// stageRow is the minimal shape every stages.sql.go row type shares -
// ListStages/ListStagesByManga/GetStageByID/UpsertStage all return the same
// four columns under different generated row types, so this lets a single
// mapper function serve all of them via a small per-caller adapter.
type stageRow struct {
	ID       pgtype.UUID
	Manga    string
	Position int32
	Name     string
}

func toStage(r stageRow) (game.Stage, error) {
	manga, err := enums.ParseManga(r.Manga)
	if err != nil {
		return game.Stage{}, fmt.Errorf("stage %q: %w", r.Name, err)
	}
	return game.NewStage(r.ID.Bytes, manga, int(r.Position), r.Name)
}

func fromListStagesRow(r db.ListStagesRow) stageRow {
	return stageRow{ID: r.ID, Manga: r.Manga, Position: r.Position, Name: r.Name}
}

func fromListStagesByMangaRow(r db.ListStagesByMangaRow) stageRow {
	return stageRow{ID: r.ID, Manga: r.Manga, Position: r.Position, Name: r.Name}
}

func fromGetStageByIDRow(r db.GetStageByIDRow) stageRow {
	return stageRow{ID: r.ID, Manga: r.Manga, Position: r.Position, Name: r.Name}
}

func fromUpsertStageRow(r db.UpsertStageRow) stageRow {
	return stageRow{ID: r.ID, Manga: r.Manga, Position: r.Position, Name: r.Name}
}
