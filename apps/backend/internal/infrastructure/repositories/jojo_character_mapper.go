package repositories

import (
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// jojoCharacterRow is the common shape shared by every JoJo character query
// row (GetByID/GetByName/List/Filter/Page), so all of them can be hydrated
// by the same builder below - same pattern as devilFruitRow.
type jojoCharacterRow struct {
	ID             pgtype.UUID
	Name           string
	Description    string
	Rarity         string
	Picture        string
	PictureThumb   string
	PictureCard    string
	PictureMediaID string
	PictureStatus  string
	PictureLqip    string
	Hamon          string
	Spin           string
	BattleIQ       int16
}

func jojoCharacterRowFromGetByID(r db.GetJojoCharacterRowByIDRow) jojoCharacterRow {
	return jojoCharacterRow{
		ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
		PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
		PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
		Hamon: r.Hamon, Spin: r.Spin, BattleIQ: r.BattleIq,
	}
}

func jojoCharacterRowFromGetByName(r db.GetJojoCharacterRowByNameRow) jojoCharacterRow {
	return jojoCharacterRow{
		ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
		PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
		PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
		Hamon: r.Hamon, Spin: r.Spin, BattleIQ: r.BattleIq,
	}
}

func jojoCharacterRowsFromList(rs []db.ListJojoCharacterRowsRow) []jojoCharacterRow {
	rows := make([]jojoCharacterRow, len(rs))
	for i, r := range rs {
		rows[i] = jojoCharacterRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
			PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
			Hamon: r.Hamon, Spin: r.Spin, BattleIQ: r.BattleIq,
		}
	}
	return rows
}

func jojoCharacterRowsFromFilter(rs []db.FilterJojoCharacterRowsRow) []jojoCharacterRow {
	rows := make([]jojoCharacterRow, len(rs))
	for i, r := range rs {
		rows[i] = jojoCharacterRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
			PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
			Hamon: r.Hamon, Spin: r.Spin, BattleIQ: r.BattleIq,
		}
	}
	return rows
}

func jojoCharacterRowsFromPage(rs []db.PageJojoCharacterRowsRow) []jojoCharacterRow {
	rows := make([]jojoCharacterRow, len(rs))
	for i, r := range rs {
		rows[i] = jojoCharacterRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
			PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
			Hamon: r.Hamon, Spin: r.Spin, BattleIQ: r.BattleIq,
		}
	}
	return rows
}

// buildJojoCharacter turns a single jojoCharacterRow into a fully validated
// *characters.JojoCharacter.
func buildJojoCharacter(row jojoCharacterRow) (*characters.JojoCharacter, error) {
	id := characters.CharacterID(row.ID.Bytes)

	rarity, err := enums.ParsePowerRarity(row.Rarity)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: %w", row.Name, err)
	}
	character, err := characters.NewCharacter(id, enums.Jojo, row.Name, rarity, row.Description, row.Picture)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: %w", row.Name, err)
	}
	pictureStatus, err := enums.ParsePictureStatus(row.PictureStatus)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: picture_status: %w", row.Name, err)
	}
	character.SetPictureRenditions(row.Picture, row.PictureThumb, row.PictureCard, row.PictureLqip, pictureStatus)
	character.SetMediaID(row.PictureMediaID)

	hamon, err := enums.ParseHamonLevel(row.Hamon)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: hamon: %w", row.Name, err)
	}
	spin, err := enums.ParseSpinLevel(row.Spin)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: spin: %w", row.Name, err)
	}

	jojo, err := characters.NewJojoCharacter(character, hamon, spin, byte(row.BattleIQ))
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: %w", row.Name, err)
	}
	return jojo, nil
}

// buildJojoCharacters builds every row into a *characters.JojoCharacter, in
// order, skipping (and logging) any row that fails validation - same
// convention as buildDevilFruits.
func buildJojoCharacters(rows []jojoCharacterRow) ([]*characters.JojoCharacter, error) {
	result := make([]*characters.JojoCharacter, 0, len(rows))
	for _, row := range rows {
		c, err := buildJojoCharacter(row)
		if err != nil {
			log.Printf("jojo character %s: skipping corrupt row: %v", characters.CharacterID(row.ID.Bytes), err)
			continue
		}
		result = append(result, c)
	}
	return result, nil
}
