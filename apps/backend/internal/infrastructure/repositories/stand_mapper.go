package repositories

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// standRow is the common shape shared by GetStandRowsByNameRow,
// ListStandRowsRow and FilterStandRowsRow, so all three can be hydrated by
// the same builder below.
type standRow struct {
	ID            pgtype.UUID
	Name          string
	Description   string
	Rarity        string
	Picture       string
	PictureThumb  string
	PictureStatus string
	AttackPower   string
	Speed         string
	AttackRange   string
	Endurance     string
	Precision     string
	Potential     string
	EvolvesFromID pgtype.UUID
	Matched       bool
	Skills        []string
}

func standRowsFromGetByName(rs []db.GetStandRowsByNameRow) []standRow {
	rows := make([]standRow, len(rs))
	for i, r := range rs {
		rows[i] = standRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus,
			AttackPower: r.AttackPower, Speed: r.Speed, AttackRange: r.AttackRange, Endurance: r.Endurance,
			Precision: r.Precision, Potential: r.Potential, EvolvesFromID: r.EvolvesFromID,
			Matched: r.Matched, Skills: r.Skills,
		}
	}
	return rows
}

func standRowsFromGetByID(rs []db.GetStandRowsByIDRow) []standRow {
	rows := make([]standRow, len(rs))
	for i, r := range rs {
		rows[i] = standRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus,
			AttackPower: r.AttackPower, Speed: r.Speed, AttackRange: r.AttackRange, Endurance: r.Endurance,
			Precision: r.Precision, Potential: r.Potential, EvolvesFromID: r.EvolvesFromID,
			Matched: r.Matched, Skills: r.Skills,
		}
	}
	return rows
}

func standRowsFromList(rs []db.ListStandRowsRow) []standRow {
	rows := make([]standRow, len(rs))
	for i, r := range rs {
		rows[i] = standRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus,
			AttackPower: r.AttackPower, Speed: r.Speed, AttackRange: r.AttackRange, Endurance: r.Endurance,
			Precision: r.Precision, Potential: r.Potential, EvolvesFromID: r.EvolvesFromID,
			Matched: r.Matched, Skills: r.Skills,
		}
	}
	return rows
}

func standRowsFromFilter(rs []db.FilterStandRowsRow) []standRow {
	rows := make([]standRow, len(rs))
	for i, r := range rs {
		rows[i] = standRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureStatus: r.PictureStatus,
			AttackPower: r.AttackPower, Speed: r.Speed, AttackRange: r.AttackRange, Endurance: r.Endurance,
			Precision: r.Precision, Potential: r.Potential, EvolvesFromID: r.EvolvesFromID,
			Matched: r.Matched, Skills: r.Skills,
		}
	}
	return rows
}

// buildStands turns a flat set of standRow (a matched stand or two, plus the
// ancestor rows needed to hydrate their EvolvesFrom chain) into fully
// validated *powers.Stand values, resolving each evolvesFrom pointer before
// its descendant is built - Stand's fields are unexported, so the only way
// to construct one is powers.NewStand, which takes the parent by value.
// Only rows with Matched == true are returned, in their original (name)
// order; ancestor-only rows are used purely to satisfy EvolvesFrom.
func buildStands(rows []standRow) ([]*powers.Stand, error) {
	byID := make(map[powers.PowerID]standRow, len(rows))
	order := make([]powers.PowerID, 0, len(rows))
	for _, r := range rows {
		id := powers.PowerID(r.ID.Bytes)
		if _, ok := byID[id]; !ok {
			order = append(order, id)
		}
		byID[id] = r
	}

	built := make(map[powers.PowerID]*powers.Stand, len(rows))
	building := make(map[powers.PowerID]bool, len(rows))

	var resolve func(id powers.PowerID) (*powers.Stand, error)
	resolve = func(id powers.PowerID) (*powers.Stand, error) {
		if s, ok := built[id]; ok {
			return s, nil
		}
		row, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("stand %s: evolves_from references a stand outside the loaded set", id)
		}
		if building[id] {
			return nil, fmt.Errorf("stand %q: cyclic evolves_from chain detected", row.Name)
		}
		building[id] = true
		defer delete(building, id)

		var evolvesFrom *powers.Stand
		if row.EvolvesFromID.Valid {
			parent, err := resolve(powers.PowerID(row.EvolvesFromID.Bytes))
			if err != nil {
				return nil, fmt.Errorf("stand %q: resolving evolves_from: %w", row.Name, err)
			}
			evolvesFrom = parent
		}

		rarity, err := enums.ParsePowerRarity(row.Rarity)
		if err != nil {
			return nil, fmt.Errorf("stand %q: %w", row.Name, err)
		}
		skills := row.Skills
		power, err := powers.NewPower(id, row.Name, row.Description, rarity, &skills, row.Picture)
		if err != nil {
			return nil, fmt.Errorf("stand %q: %w", row.Name, err)
		}
		pictureStatus, err := enums.ParsePictureStatus(row.PictureStatus)
		if err != nil {
			return nil, fmt.Errorf("stand %q: picture_status: %w", row.Name, err)
		}
		power.SetPictureRenditions(row.Picture, row.PictureThumb, pictureStatus)

		attackPower, err := enums.ParseStandStat(row.AttackPower)
		if err != nil {
			return nil, fmt.Errorf("stand %q: attack_power: %w", row.Name, err)
		}
		speed, err := enums.ParseStandStat(row.Speed)
		if err != nil {
			return nil, fmt.Errorf("stand %q: speed: %w", row.Name, err)
		}
		attackRange, err := enums.ParseStandStat(row.AttackRange)
		if err != nil {
			return nil, fmt.Errorf("stand %q: attack_range: %w", row.Name, err)
		}
		endurance, err := enums.ParseStandStat(row.Endurance)
		if err != nil {
			return nil, fmt.Errorf("stand %q: endurance: %w", row.Name, err)
		}
		precision, err := enums.ParseStandStat(row.Precision)
		if err != nil {
			return nil, fmt.Errorf("stand %q: precision: %w", row.Name, err)
		}
		potential, err := enums.ParseStandStat(row.Potential)
		if err != nil {
			return nil, fmt.Errorf("stand %q: potential: %w", row.Name, err)
		}

		stand, err := powers.NewStand(*power, attackPower, speed, attackRange, endurance, precision, potential, evolvesFrom)
		if err != nil {
			return nil, fmt.Errorf("stand %q: %w", row.Name, err)
		}
		built[id] = stand
		return stand, nil
	}

	result := make([]*powers.Stand, 0, len(rows))
	for _, id := range order {
		if !byID[id].Matched {
			continue
		}
		stand, err := resolve(id)
		if err != nil {
			return nil, err
		}
		result = append(result, stand)
	}
	return result, nil
}
