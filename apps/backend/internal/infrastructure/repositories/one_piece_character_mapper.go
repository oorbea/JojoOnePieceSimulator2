package repositories

import (
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// onePieceCharacterRow is the common shape shared by every One Piece
// character query row - see jojoCharacterRow's doc.
type onePieceCharacterRow struct {
	ID              pgtype.UUID
	Name            string
	Description     string
	Rarity          string
	Picture         string
	PictureThumb    string
	PictureCard     string
	PictureMediaID  string
	PictureStatus   string
	PictureLqip     string
	PhysicalForm    string
	ArmamentHaki    string
	ObservationHaki string
	ConquerorHaki   string
	FruitMastery    string
}

func onePieceCharacterRowFromGetByID(r db.GetOnePieceCharacterRowByIDRow) onePieceCharacterRow {
	return onePieceCharacterRow{
		ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
		PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
		PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
		PhysicalForm: r.PhysicalForm, ArmamentHaki: r.ArmamentHaki,
		ObservationHaki: r.ObservationHaki, ConquerorHaki: r.ConquerorHaki, FruitMastery: r.FruitMastery,
	}
}

func onePieceCharacterRowFromGetByName(r db.GetOnePieceCharacterRowByNameRow) onePieceCharacterRow {
	return onePieceCharacterRow{
		ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
		PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
		PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
		PhysicalForm: r.PhysicalForm, ArmamentHaki: r.ArmamentHaki,
		ObservationHaki: r.ObservationHaki, ConquerorHaki: r.ConquerorHaki, FruitMastery: r.FruitMastery,
	}
}

func onePieceCharacterRowsFromList(rs []db.ListOnePieceCharacterRowsRow) []onePieceCharacterRow {
	rows := make([]onePieceCharacterRow, len(rs))
	for i, r := range rs {
		rows[i] = onePieceCharacterRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
			PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
			PhysicalForm: r.PhysicalForm, ArmamentHaki: r.ArmamentHaki,
			ObservationHaki: r.ObservationHaki, ConquerorHaki: r.ConquerorHaki, FruitMastery: r.FruitMastery,
		}
	}
	return rows
}

func onePieceCharacterRowsFromFilter(rs []db.FilterOnePieceCharacterRowsRow) []onePieceCharacterRow {
	rows := make([]onePieceCharacterRow, len(rs))
	for i, r := range rs {
		rows[i] = onePieceCharacterRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
			PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
			PhysicalForm: r.PhysicalForm, ArmamentHaki: r.ArmamentHaki,
			ObservationHaki: r.ObservationHaki, ConquerorHaki: r.ConquerorHaki, FruitMastery: r.FruitMastery,
		}
	}
	return rows
}

func onePieceCharacterRowsFromPage(rs []db.PageOnePieceCharacterRowsRow) []onePieceCharacterRow {
	rows := make([]onePieceCharacterRow, len(rs))
	for i, r := range rs {
		rows[i] = onePieceCharacterRow{
			ID: r.ID, Name: r.Name, Description: r.Description, Rarity: r.Rarity, Picture: r.Picture,
			PictureThumb: r.PictureThumb, PictureCard: r.PictureCard, PictureStatus: r.PictureStatus,
			PictureLqip: r.PictureLqip, PictureMediaID: r.PictureMediaID,
			PhysicalForm: r.PhysicalForm, ArmamentHaki: r.ArmamentHaki,
			ObservationHaki: r.ObservationHaki, ConquerorHaki: r.ConquerorHaki, FruitMastery: r.FruitMastery,
		}
	}
	return rows
}

// buildOnePieceCharacter turns a single onePieceCharacterRow into a fully
// validated *characters.OnePieceCharacter.
func buildOnePieceCharacter(row onePieceCharacterRow) (*characters.OnePieceCharacter, error) {
	id := characters.CharacterID(row.ID.Bytes)

	rarity, err := enums.ParsePowerRarity(row.Rarity)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: %w", row.Name, err)
	}
	character, err := characters.NewCharacter(id, enums.OnePiece, row.Name, rarity, row.Description, row.Picture)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: %w", row.Name, err)
	}
	pictureStatus, err := enums.ParsePictureStatus(row.PictureStatus)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: picture_status: %w", row.Name, err)
	}
	character.SetPictureRenditions(row.Picture, row.PictureThumb, row.PictureCard, row.PictureLqip, pictureStatus)
	character.SetMediaID(row.PictureMediaID)

	physicalForm, err := enums.ParsePhysicalForm(row.PhysicalForm)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: physical_form: %w", row.Name, err)
	}
	armamentHaki, err := enums.ParseHakiLevel(row.ArmamentHaki)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: armament_haki: %w", row.Name, err)
	}
	observationHaki, err := enums.ParseHakiLevel(row.ObservationHaki)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: observation_haki: %w", row.Name, err)
	}
	conquerorHaki, err := enums.ParseHakiLevel(row.ConquerorHaki)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: conqueror_haki: %w", row.Name, err)
	}
	fruitMastery, err := enums.ParseFruitMastery(row.FruitMastery)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: fruit_mastery: %w", row.Name, err)
	}

	onePiece, err := characters.NewOnePieceCharacter(character, physicalForm, armamentHaki, observationHaki, conquerorHaki, fruitMastery)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: %w", row.Name, err)
	}
	return onePiece, nil
}

// buildOnePieceCharacters builds every row into a *characters.OnePieceCharacter,
// in order, skipping (and logging) any row that fails validation - same
// convention as buildDevilFruits/buildJojoCharacters.
func buildOnePieceCharacters(rows []onePieceCharacterRow) ([]*characters.OnePieceCharacter, error) {
	result := make([]*characters.OnePieceCharacter, 0, len(rows))
	for _, row := range rows {
		c, err := buildOnePieceCharacter(row)
		if err != nil {
			log.Printf("one piece character %s: skipping corrupt row: %v", characters.CharacterID(row.ID.Bytes), err)
			continue
		}
		result = append(result, c)
	}
	return result, nil
}
