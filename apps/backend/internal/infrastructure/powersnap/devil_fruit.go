package powersnap

import (
	"encoding/json"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// DevilFruitSnapshot is the JSON-serializable shape of a *powers.DevilFruit.
// DevilFruit's fields are all unexported, so it can't be marshaled directly -
// this reads it through its public getters, same reasoning as StandSnapshot.
type DevilFruitSnapshot struct {
	ID             [16]byte `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Rarity         string   `json:"rarity"`
	Skills         []string `json:"skills"`
	Picture        string   `json:"picture"`
	PictureThumb   string   `json:"pictureThumb"`
	PictureCard    string   `json:"pictureCard"`
	PictureStatus  string   `json:"pictureStatus"`
	PictureLqip    string   `json:"pictureLqip"`
	PictureMediaID string   `json:"pictureMediaId"`
	FocalX         float64  `json:"focalX"`
	FocalY         float64  `json:"focalY"`
	FruitType      string   `json:"fruitType"`
	// EvolvesFrom is only ever set on a fruit a game drew with its evolution
	// attached (Hito Hito no mi: Model Nika <- Gomu Gomu no mi - see
	// game.LinkFruitEvolutions); catalogue-cached fruits never carry one, so
	// payloads written before this field existed decode to nil unchanged.
	EvolvesFrom *DevilFruitSnapshot `json:"evolvesFrom,omitempty"`
}

func OfDevilFruit(fruit *powers.DevilFruit) DevilFruitSnapshot {
	var evolvesFrom *DevilFruitSnapshot
	if parent := fruit.EvolvesFrom(); parent != nil {
		s := OfDevilFruit(parent)
		evolvesFrom = &s
	}
	return DevilFruitSnapshot{
		ID:             fruit.ID(),
		Name:           fruit.Name(),
		Description:    fruit.Description(),
		Rarity:         fruit.Rarity().String(),
		Skills:         fruit.Skills(),
		Picture:        fruit.Picture(),
		PictureThumb:   fruit.PictureThumb(),
		PictureCard:    fruit.PictureCard(),
		PictureStatus:  fruit.PictureStatus().String(),
		PictureLqip:    fruit.PictureLqip(),
		PictureMediaID: fruit.PictureMediaID(),
		FocalX:         fruit.FocalX(),
		FocalY:         fruit.FocalY(),
		FruitType:      fruit.FruitType().String(),
		EvolvesFrom:    evolvesFrom,
	}
}

func (s DevilFruitSnapshot) Hydrate() (*powers.DevilFruit, error) {
	var evolvesFrom *powers.DevilFruit
	if s.EvolvesFrom != nil {
		parent, err := s.EvolvesFrom.Hydrate()
		if err != nil {
			return nil, err
		}
		evolvesFrom = parent
	}

	rarity, err := enums.ParsePowerRarity(s.Rarity)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: %w", s.Name, err)
	}
	skills := s.Skills
	power, err := powers.NewPower(s.ID, s.Name, s.Description, rarity, &skills, s.Picture)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: %w", s.Name, err)
	}
	pictureStatus, err := enums.ParsePictureStatus(s.PictureStatus)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: picture_status: %w", s.Name, err)
	}
	power.SetPictureRenditions(s.Picture, s.PictureThumb, s.PictureCard, s.PictureLqip, pictureStatus)
	power.SetMediaID(s.PictureMediaID)
	if err := power.SetFocalPoint(s.FocalX, s.FocalY); err != nil {
		return nil, fmt.Errorf("devil fruit %q: focal: %w", s.Name, err)
	}

	fruitType, err := enums.ParseFruitType(s.FruitType)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: fruit_type: %w", s.Name, err)
	}

	fruit, err := powers.NewDevilFruit(*power, fruitType)
	if err != nil {
		return nil, err
	}
	if evolvesFrom != nil {
		fruit = fruit.WithEvolvesFrom(evolvesFrom)
	}
	return fruit, nil
}

func MarshalDevilFruit(fruit *powers.DevilFruit) ([]byte, error) {
	return json.Marshal(OfDevilFruit(fruit))
}

func UnmarshalDevilFruit(data []byte) (*powers.DevilFruit, error) {
	var s DevilFruitSnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshaling devil fruit: %w", err)
	}
	return s.Hydrate()
}

func MarshalDevilFruits(fruits []*powers.DevilFruit) ([]byte, error) {
	snapshots := make([]DevilFruitSnapshot, len(fruits))
	for i, fruit := range fruits {
		snapshots[i] = OfDevilFruit(fruit)
	}
	return json.Marshal(snapshots)
}

func UnmarshalDevilFruits(data []byte) ([]*powers.DevilFruit, error) {
	var snapshots []DevilFruitSnapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, fmt.Errorf("unmarshaling devil fruits: %w", err)
	}
	fruits := make([]*powers.DevilFruit, len(snapshots))
	for i, s := range snapshots {
		fruit, err := s.Hydrate()
		if err != nil {
			return nil, err
		}
		fruits[i] = fruit
	}
	return fruits, nil
}
