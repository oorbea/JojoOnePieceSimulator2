// Package cache decorates ports.IStandRepository and ports.IPictureStorage
// with a ports.ICache, so read-heavy calls can skip Postgres/R2 entirely on
// a hit. Both decorators are fail-open: any cache error is treated as a
// miss, never as a request failure.
package cache

import (
	"encoding/json"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// standSnapshot is the JSON-serializable shape of a *powers.Stand. Stand's
// fields are all unexported, so it can't be marshaled directly - snapshot
// reads it through its public getters instead of reaching into the
// repositories package's own (also unexported) row types, keeping this
// package's only coupling to the domain's public API.
type standSnapshot struct {
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
	EvolvesFrom   *standSnapshot `json:"evolvesFrom,omitempty"`
}

// snapshot captures stand (and, recursively, its whole EvolvesFrom
// ancestor chain - short in practice, so no dedup) into its serializable
// form.
func snapshot(stand *powers.Stand) standSnapshot {
	var evolvesFrom *standSnapshot
	if parent := stand.EvolvesFrom(); parent != nil {
		s := snapshot(parent)
		evolvesFrom = &s
	}
	return standSnapshot{
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

// hydrate rebuilds a *powers.Stand from s, resolving EvolvesFrom first since
// powers.NewStand takes its parent by value - the same parent-first
// ordering repositories.buildStands relies on.
func (s standSnapshot) hydrate() (*powers.Stand, error) {
	var evolvesFrom *powers.Stand
	if s.EvolvesFrom != nil {
		parent, err := s.EvolvesFrom.hydrate()
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

// marshalStand/marshalStands/unmarshalStand/unmarshalStands are the
// (de)serialization entry points the repository decorator uses; kept
// separate from snapshot/hydrate so JSON errors are reported at one place.

func marshalStand(stand *powers.Stand) ([]byte, error) {
	return json.Marshal(snapshot(stand))
}

func unmarshalStand(data []byte) (*powers.Stand, error) {
	var s standSnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshaling cached stand: %w", err)
	}
	return s.hydrate()
}

func marshalStands(stands []*powers.Stand) ([]byte, error) {
	snapshots := make([]standSnapshot, len(stands))
	for i, stand := range stands {
		snapshots[i] = snapshot(stand)
	}
	return json.Marshal(snapshots)
}

func unmarshalStands(data []byte) ([]*powers.Stand, error) {
	var snapshots []standSnapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, fmt.Errorf("unmarshaling cached stands: %w", err)
	}
	stands := make([]*powers.Stand, len(snapshots))
	for i, s := range snapshots {
		stand, err := s.hydrate()
		if err != nil {
			return nil, err
		}
		stands[i] = stand
	}
	return stands, nil
}
