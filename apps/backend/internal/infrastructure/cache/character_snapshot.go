package cache

import (
	"encoding/json"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// jojoCharacterSnapshot/onePieceCharacterSnapshot are this package's own
// JSON-serializable shape for the two Character subtypes - kept local to
// the cache package rather than added to infrastructure/powersnap, since
// (per the owner's decision) a Character never gets embedded inside a live
// Loadout the way a Stand/DevilFruit does; this cache is the only consumer.

type jojoCharacterSnapshot struct {
	ID             [16]byte `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Rarity         string   `json:"rarity"`
	Picture        string   `json:"picture"`
	PictureThumb   string   `json:"pictureThumb"`
	PictureCard    string   `json:"pictureCard"`
	PictureStatus  string   `json:"pictureStatus"`
	PictureLqip    string   `json:"pictureLqip"`
	PictureMediaID string   `json:"pictureMediaId"`
	Hamon          string   `json:"hamon"`
	Spin           string   `json:"spin"`
	BattleIQ       byte     `json:"battleIq"`
}

func ofJojoCharacter(c *characters.JojoCharacter) jojoCharacterSnapshot {
	return jojoCharacterSnapshot{
		ID: c.ID(), Name: c.Name(), Description: c.Description(), Rarity: c.Rarity().String(),
		Picture: c.Picture(), PictureThumb: c.PictureThumb(), PictureCard: c.PictureCard(),
		PictureStatus: c.PictureStatus().String(), PictureLqip: c.PictureLqip(), PictureMediaID: c.PictureMediaID(),
		Hamon: c.Hamon().String(), Spin: c.Spin().String(), BattleIQ: c.BattleIQ(),
	}
}

func (s jojoCharacterSnapshot) hydrate() (*characters.JojoCharacter, error) {
	rarity, err := enums.ParsePowerRarity(s.Rarity)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: %w", s.Name, err)
	}
	character, err := characters.NewCharacter(s.ID, enums.Jojo, s.Name, rarity, s.Description, s.Picture)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: %w", s.Name, err)
	}
	pictureStatus, err := enums.ParsePictureStatus(s.PictureStatus)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: picture_status: %w", s.Name, err)
	}
	character.SetPictureRenditions(s.Picture, s.PictureThumb, s.PictureCard, s.PictureLqip, pictureStatus)
	character.SetMediaID(s.PictureMediaID)

	hamon, err := enums.ParseHamonLevel(s.Hamon)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: hamon: %w", s.Name, err)
	}
	spin, err := enums.ParseSpinLevel(s.Spin)
	if err != nil {
		return nil, fmt.Errorf("jojo character %q: spin: %w", s.Name, err)
	}
	return characters.NewJojoCharacter(character, hamon, spin, s.BattleIQ)
}

func marshalJojoCharacter(c *characters.JojoCharacter) ([]byte, error) {
	return json.Marshal(ofJojoCharacter(c))
}

func unmarshalJojoCharacter(data []byte) (*characters.JojoCharacter, error) {
	var s jojoCharacterSnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return s.hydrate()
}

func marshalJojoCharacters(list []*characters.JojoCharacter) ([]byte, error) {
	snapshots := make([]jojoCharacterSnapshot, len(list))
	for i, c := range list {
		snapshots[i] = ofJojoCharacter(c)
	}
	return json.Marshal(snapshots)
}

func unmarshalJojoCharacters(data []byte) ([]*characters.JojoCharacter, error) {
	var snapshots []jojoCharacterSnapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, err
	}
	result := make([]*characters.JojoCharacter, len(snapshots))
	for i, s := range snapshots {
		c, err := s.hydrate()
		if err != nil {
			return nil, err
		}
		result[i] = c
	}
	return result, nil
}

type onePieceCharacterSnapshot struct {
	ID              [16]byte `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Rarity          string   `json:"rarity"`
	Picture         string   `json:"picture"`
	PictureThumb    string   `json:"pictureThumb"`
	PictureCard     string   `json:"pictureCard"`
	PictureStatus   string   `json:"pictureStatus"`
	PictureLqip     string   `json:"pictureLqip"`
	PictureMediaID  string   `json:"pictureMediaId"`
	PhysicalForm    string   `json:"physicalForm"`
	ArmamentHaki    string   `json:"armamentHaki"`
	ObservationHaki string   `json:"observationHaki"`
	ConquerorHaki   string   `json:"conquerorHaki"`
	FruitMastery    string   `json:"fruitMastery"`
}

func ofOnePieceCharacter(c *characters.OnePieceCharacter) onePieceCharacterSnapshot {
	return onePieceCharacterSnapshot{
		ID: c.ID(), Name: c.Name(), Description: c.Description(), Rarity: c.Rarity().String(),
		Picture: c.Picture(), PictureThumb: c.PictureThumb(), PictureCard: c.PictureCard(),
		PictureStatus: c.PictureStatus().String(), PictureLqip: c.PictureLqip(), PictureMediaID: c.PictureMediaID(),
		PhysicalForm: c.PhysicalForm().String(), ArmamentHaki: c.ArmamentHaki().String(),
		ObservationHaki: c.ObservationHaki().String(), ConquerorHaki: c.ConquerorHaki().String(),
		FruitMastery: c.FruitMastery().String(),
	}
}

func (s onePieceCharacterSnapshot) hydrate() (*characters.OnePieceCharacter, error) {
	rarity, err := enums.ParsePowerRarity(s.Rarity)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: %w", s.Name, err)
	}
	character, err := characters.NewCharacter(s.ID, enums.OnePiece, s.Name, rarity, s.Description, s.Picture)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: %w", s.Name, err)
	}
	pictureStatus, err := enums.ParsePictureStatus(s.PictureStatus)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: picture_status: %w", s.Name, err)
	}
	character.SetPictureRenditions(s.Picture, s.PictureThumb, s.PictureCard, s.PictureLqip, pictureStatus)
	character.SetMediaID(s.PictureMediaID)

	physicalForm, err := enums.ParsePhysicalForm(s.PhysicalForm)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: physical_form: %w", s.Name, err)
	}
	armamentHaki, err := enums.ParseHakiLevel(s.ArmamentHaki)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: armament_haki: %w", s.Name, err)
	}
	observationHaki, err := enums.ParseHakiLevel(s.ObservationHaki)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: observation_haki: %w", s.Name, err)
	}
	conquerorHaki, err := enums.ParseHakiLevel(s.ConquerorHaki)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: conqueror_haki: %w", s.Name, err)
	}
	fruitMastery, err := enums.ParseFruitMastery(s.FruitMastery)
	if err != nil {
		return nil, fmt.Errorf("one piece character %q: fruit_mastery: %w", s.Name, err)
	}
	return characters.NewOnePieceCharacter(character, physicalForm, armamentHaki, observationHaki, conquerorHaki, fruitMastery)
}

func marshalOnePieceCharacter(c *characters.OnePieceCharacter) ([]byte, error) {
	return json.Marshal(ofOnePieceCharacter(c))
}

func unmarshalOnePieceCharacter(data []byte) (*characters.OnePieceCharacter, error) {
	var s onePieceCharacterSnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return s.hydrate()
}

func marshalOnePieceCharacters(list []*characters.OnePieceCharacter) ([]byte, error) {
	snapshots := make([]onePieceCharacterSnapshot, len(list))
	for i, c := range list {
		snapshots[i] = ofOnePieceCharacter(c)
	}
	return json.Marshal(snapshots)
}

func unmarshalOnePieceCharacters(data []byte) ([]*characters.OnePieceCharacter, error) {
	var snapshots []onePieceCharacterSnapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, err
	}
	result := make([]*characters.OnePieceCharacter, len(snapshots))
	for i, s := range snapshots {
		c, err := s.hydrate()
		if err != nil {
			return nil, err
		}
		result[i] = c
	}
	return result, nil
}
