package cache

import (
	"encoding/json"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// devilFruitSnapshot is the JSON-serializable shape of a *powers.DevilFruit.
// DevilFruit's fields are all unexported, so it can't be marshaled directly -
// snapshot reads it through its public getters, same reasoning as
// standSnapshot.
type devilFruitSnapshot struct {
	ID            [16]byte `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Rarity        string   `json:"rarity"`
	Skills        []string `json:"skills"`
	Picture       string   `json:"picture"`
	PictureThumb  string   `json:"pictureThumb"`
	PictureStatus string   `json:"pictureStatus"`
	FruitType     string   `json:"fruitType"`
}

func snapshotDevilFruit(fruit *powers.DevilFruit) devilFruitSnapshot {
	return devilFruitSnapshot{
		ID:            fruit.ID(),
		Name:          fruit.Name(),
		Description:   fruit.Description(),
		Rarity:        fruit.Rarity().String(),
		Skills:        fruit.Skills(),
		Picture:       fruit.Picture(),
		PictureThumb:  fruit.PictureThumb(),
		PictureStatus: fruit.PictureStatus().String(),
		FruitType:     fruit.FruitType().String(),
	}
}

func (s devilFruitSnapshot) hydrate() (*powers.DevilFruit, error) {
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
	power.SetPictureRenditions(s.Picture, s.PictureThumb, pictureStatus)

	fruitType, err := enums.ParseFruitType(s.FruitType)
	if err != nil {
		return nil, fmt.Errorf("devil fruit %q: fruit_type: %w", s.Name, err)
	}

	return powers.NewDevilFruit(*power, fruitType)
}

func marshalDevilFruit(fruit *powers.DevilFruit) ([]byte, error) {
	return json.Marshal(snapshotDevilFruit(fruit))
}

func unmarshalDevilFruit(data []byte) (*powers.DevilFruit, error) {
	var s devilFruitSnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshaling cached devil fruit: %w", err)
	}
	return s.hydrate()
}

func marshalDevilFruits(fruits []*powers.DevilFruit) ([]byte, error) {
	snapshots := make([]devilFruitSnapshot, len(fruits))
	for i, fruit := range fruits {
		snapshots[i] = snapshotDevilFruit(fruit)
	}
	return json.Marshal(snapshots)
}

func unmarshalDevilFruits(data []byte) ([]*powers.DevilFruit, error) {
	var snapshots []devilFruitSnapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, fmt.Errorf("unmarshaling cached devil fruits: %w", err)
	}
	fruits := make([]*powers.DevilFruit, len(snapshots))
	for i, s := range snapshots {
		fruit, err := s.hydrate()
		if err != nil {
			return nil, err
		}
		fruits[i] = fruit
	}
	return fruits, nil
}
