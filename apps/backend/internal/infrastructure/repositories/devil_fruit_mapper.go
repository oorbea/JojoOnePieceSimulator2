package repositories

import (
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// devilFruitRow is the common shape shared by GetDevilFruitRowByIDRow,
// GetDevilFruitRowByNameRow, ListDevilFruitRowsRow and FilterDevilFruitRowsRow,
// so all four can be hydrated by the same builder below.
type devilFruitRow struct {
	ID            pgtype.UUID
	Name          string
	Description   string
	Rarity        string
	Picture       string
	PictureThumb  string
	PictureStatus string
	FruitType     string
	Skills        []string
}

func devilFruitRowFromGetByID(r db.GetDevilFruitRowByIDRow) devilFruitRow {
	return devilFruitRow{
		ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
		PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus, FruitType: r.FruitType, Skills: r.Skills,
	}
}

func devilFruitRowFromGetByName(r db.GetDevilFruitRowByNameRow) devilFruitRow {
	return devilFruitRow{
		ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
		PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus, FruitType: r.FruitType, Skills: r.Skills,
	}
}

func devilFruitRowsFromList(rs []db.ListDevilFruitRowsRow) []devilFruitRow {
	rows := make([]devilFruitRow, len(rs))
	for i, r := range rs {
		rows[i] = devilFruitRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus, FruitType: r.FruitType, Skills: r.Skills,
		}
	}
	return rows
}

func devilFruitRowsFromFilter(rs []db.FilterDevilFruitRowsRow) []devilFruitRow {
	rows := make([]devilFruitRow, len(rs))
	for i, r := range rs {
		rows[i] = devilFruitRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus, FruitType: r.FruitType, Skills: r.Skills,
		}
	}
	return rows
}

// buildDevilFruit turns a single devilFruitRow into a fully validated
// *powers.DevilFruit. Unlike Stand, DevilFruit has no evolves_from chain, so
// there is no topological resolution or cycle detection to do here.
func buildDevilFruit(row devilFruitRow) (*powers.DevilFruit, error) {
	id := powers.PowerID(row.ID.Bytes)

	rarity, err := enums.ParsePowerRarity(row.Rarity)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: %w", row.Name, err)
	}
	skills := row.Skills
	power, err := powers.NewPower(id, row.Name, row.Description, rarity, &skills, row.Picture)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: %w", row.Name, err)
	}
	pictureStatus, err := enums.ParsePictureStatus(row.PictureStatus)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: picture_status: %w", row.Name, err)
	}
	power.SetPictureRenditions(row.Picture, row.PictureThumb, pictureStatus)

	fruitType, err := enums.ParseFruitType(row.FruitType)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: fruit_type: %w", row.Name, err)
	}

	fruit, err := powers.NewDevilFruit(*power, fruitType)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: %w", row.Name, err)
	}
	return fruit, nil
}

// buildDevilFruits builds every row into a *powers.DevilFruit, in order,
// skipping (and logging) any row that fails validation instead of failing
// the whole batch - used by GetAll/Filter, where one corrupt legacy row must
// not 500 the entire admin catalogue. FindByID/FindByName go through
// buildDevilFruit directly and still fail loudly for their one requested row.
func buildDevilFruits(rows []devilFruitRow) ([]*powers.DevilFruit, error) {
	result := make([]*powers.DevilFruit, 0, len(rows))
	for _, row := range rows {
		fruit, err := buildDevilFruit(row)
		if err != nil {
			log.Printf("devil fruit %s: skipping corrupt row: %v", powers.PowerID(row.ID.Bytes), err)
			continue
		}
		result = append(result, fruit)
	}
	return result, nil
}
