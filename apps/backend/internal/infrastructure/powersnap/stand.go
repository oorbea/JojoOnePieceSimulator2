// Package powersnap holds the JSON-serializable shape of the two power
// entities (*powers.Stand, *powers.DevilFruit), promoted out of
// infrastructure/cache so it can be shared by any adapter that needs to
// embed a full power inside a larger payload - today that's the game store
// (a Loadout embeds powers in full, see entities/game.LoadoutSnapshot) and
// infrastructure/cache's own read-through repositories.
package powersnap

import (
	"encoding/json"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// StandSnapshot is the JSON-serializable shape of a *powers.Stand. Stand's
// fields are all unexported, so it can't be marshaled directly - this reads
// it through its public getters instead of reaching into any repository's
// own (also unexported) row types.
type StandSnapshot struct {
	ID            [16]byte       `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Rarity        string         `json:"rarity"`
	Skills        []string       `json:"skills"`
	Picture       string         `json:"picture"`
	PictureThumb  string         `json:"pictureThumb"`
	PictureStatus string         `json:"pictureStatus"`
	AttackPower   string         `json:"attackPower"`
	Speed         string         `json:"speed"`
	AttackRange   string         `json:"attackRange"`
	Endurance     string         `json:"endurance"`
	Precision     string         `json:"precision"`
	Potential     string         `json:"potential"`
	EvolvesFrom   *StandSnapshot `json:"evolvesFrom,omitempty"`
}

// OfStand captures stand (and, recursively, its whole EvolvesFrom ancestor
// chain - short in practice, so no dedup) into its serializable form.
func OfStand(stand *powers.Stand) StandSnapshot {
	var evolvesFrom *StandSnapshot
	if parent := stand.EvolvesFrom(); parent != nil {
		s := OfStand(parent)
		evolvesFrom = &s
	}
	return StandSnapshot{
		ID:            stand.ID(),
		Name:          stand.Name(),
		Description:   stand.Description(),
		Rarity:        stand.Rarity().String(),
		Skills:        stand.Skills(),
		Picture:       stand.Picture(),
		PictureThumb:  stand.PictureThumb(),
		PictureStatus: stand.PictureStatus().String(),
		AttackPower:   stand.AttackPower().String(),
		Speed:         stand.Speed().String(),
		AttackRange:   stand.AttackRange().String(),
		Endurance:     stand.Endurance().String(),
		Precision:     stand.Precision().String(),
		Potential:     stand.Potential().String(),
		EvolvesFrom:   evolvesFrom,
	}
}

// Hydrate rebuilds a *powers.Stand from s, resolving EvolvesFrom first since
// powers.NewStand takes its parent by value - the same parent-first
// ordering repositories.buildStands relies on.
func (s StandSnapshot) Hydrate() (*powers.Stand, error) {
	var evolvesFrom *powers.Stand
	if s.EvolvesFrom != nil {
		parent, err := s.EvolvesFrom.Hydrate()
		if err != nil {
			return nil, err
		}
		evolvesFrom = parent
	}

	rarity, err := enums.ParsePowerRarity(s.Rarity)
	if err != nil {
		return nil, fmt.Errorf("stand %q: %w", s.Name, err)
	}
	skills := s.Skills
	power, err := powers.NewPower(s.ID, s.Name, s.Description, rarity, &skills, s.Picture)
	if err != nil {
		return nil, fmt.Errorf("stand %q: %w", s.Name, err)
	}
	pictureStatus, err := enums.ParsePictureStatus(s.PictureStatus)
	if err != nil {
		return nil, fmt.Errorf("stand %q: picture_status: %w", s.Name, err)
	}
	power.SetPictureRenditions(s.Picture, s.PictureThumb, pictureStatus)

	attackPower, err := enums.ParseStandStat(s.AttackPower)
	if err != nil {
		return nil, fmt.Errorf("stand %q: attack_power: %w", s.Name, err)
	}
	speed, err := enums.ParseStandStat(s.Speed)
	if err != nil {
		return nil, fmt.Errorf("stand %q: speed: %w", s.Name, err)
	}
	attackRange, err := enums.ParseStandStat(s.AttackRange)
	if err != nil {
		return nil, fmt.Errorf("stand %q: attack_range: %w", s.Name, err)
	}
	endurance, err := enums.ParseStandStat(s.Endurance)
	if err != nil {
		return nil, fmt.Errorf("stand %q: endurance: %w", s.Name, err)
	}
	precision, err := enums.ParseStandStat(s.Precision)
	if err != nil {
		return nil, fmt.Errorf("stand %q: precision: %w", s.Name, err)
	}
	potential, err := enums.ParseStandStat(s.Potential)
	if err != nil {
		return nil, fmt.Errorf("stand %q: potential: %w", s.Name, err)
	}

	return powers.NewStand(*power, attackPower, speed, attackRange, endurance, precision, potential, evolvesFrom)
}

// MarshalStand/MarshalStands/UnmarshalStand/UnmarshalStands are the
// (de)serialization entry points adapters use; kept separate from
// OfStand/Hydrate so JSON errors are reported at one place.

func MarshalStand(stand *powers.Stand) ([]byte, error) {
	return json.Marshal(OfStand(stand))
}

func UnmarshalStand(data []byte) (*powers.Stand, error) {
	var s StandSnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshaling stand: %w", err)
	}
	return s.Hydrate()
}

func MarshalStands(stands []*powers.Stand) ([]byte, error) {
	snapshots := make([]StandSnapshot, len(stands))
	for i, stand := range stands {
		snapshots[i] = OfStand(stand)
	}
	return json.Marshal(snapshots)
}

func UnmarshalStands(data []byte) ([]*powers.Stand, error) {
	var snapshots []StandSnapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, fmt.Errorf("unmarshaling stands: %w", err)
	}
	stands := make([]*powers.Stand, len(snapshots))
	for i, s := range snapshots {
		stand, err := s.Hydrate()
		if err != nil {
			return nil, err
		}
		stands[i] = stand
	}
	return stands, nil
}
