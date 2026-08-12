package game

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

var (
	// ErrFruitMasteryMismatch is returned when FruitMastery does not agree
	// with whether a DevilFruit is present: NONE without a fruit, at least
	// REGULAR with one.
	ErrFruitMasteryMismatch = errors.New("fruit mastery must be NONE without a devil fruit, and at least REGULAR with one")

	// ErrSpin4Required is returned when a Stand carrying
	// enums.RequiresSpin4 is paired with any Spin level other than
	// enums.SpinInfinite.
	ErrSpin4Required = errors.New("this stand requires spin level INFINITE")
)

// Loadout is the immutable set of abilities assigned to a Participant for a
// game (Gauntlet) or a single round (Versus). Spin and Hamon are
// independent of the Stand and of each other - a player can hold any
// combination of the three - except for the handful of Stands carrying
// enums.RequiresSpin4, which force SpinInfinite. FruitMastery is coupled
// to DevilFruit: no fruit forces FruitMasteryNone, any fruit forces at
// least FruitMasteryRegular. NewLoadout enforces both invariants
// regardless of how the values were produced (random draw or, later,
// inventory).
type Loadout struct {
	stand           *powers.Stand
	devilFruit      *powers.DevilFruit
	spin            enums.SpinLevel
	hamon           enums.HamonLevel
	fruitMastery    enums.FruitMastery
	armamentHaki    enums.HakiLevel
	observationHaki enums.HakiLevel
	conquerorHaki   enums.HakiLevel
	physicalForm    enums.PhysicalForm
}

// NewLoadout validates and builds a Loadout. Pass nil for stand/devilFruit
// when the player draws neither ability.
func NewLoadout(
	stand *powers.Stand,
	devilFruit *powers.DevilFruit,
	spin enums.SpinLevel,
	hamon enums.HamonLevel,
	fruitMastery enums.FruitMastery,
	armamentHaki enums.HakiLevel,
	observationHaki enums.HakiLevel,
	conquerorHaki enums.HakiLevel,
	physicalForm enums.PhysicalForm,
) (*Loadout, error) {
	if !spin.IsValid() {
		return nil, enums.ErrInvalidSpinLevel
	}
	if !hamon.IsValid() {
		return nil, enums.ErrInvalidHamonLevel
	}
	if !fruitMastery.IsValid() {
		return nil, enums.ErrInvalidFruitMastery
	}
	if !armamentHaki.IsValid() {
		return nil, enums.ErrInvalidHakiLevel
	}
	if !observationHaki.IsValid() {
		return nil, enums.ErrInvalidHakiLevel
	}
	if !conquerorHaki.IsValid() {
		return nil, enums.ErrInvalidHakiLevel
	}
	if !physicalForm.IsValid() {
		return nil, enums.ErrInvalidPhysicalForm
	}
	if devilFruit == nil && fruitMastery != enums.FruitMasteryNone {
		return nil, ErrFruitMasteryMismatch
	}
	if devilFruit != nil && fruitMastery == enums.FruitMasteryNone {
		return nil, ErrFruitMasteryMismatch
	}
	if stand != nil && HasTrait(&stand.Power, enums.RequiresSpin4) && spin != enums.SpinInfinite {
		return nil, ErrSpin4Required
	}
	return &Loadout{
		stand:           stand,
		devilFruit:      devilFruit,
		spin:            spin,
		hamon:           hamon,
		fruitMastery:    fruitMastery,
		armamentHaki:    armamentHaki,
		observationHaki: observationHaki,
		conquerorHaki:   conquerorHaki,
		physicalForm:    physicalForm,
	}, nil
}

func (l *Loadout) Stand() *powers.Stand             { return l.stand }
func (l *Loadout) DevilFruit() *powers.DevilFruit   { return l.devilFruit }
func (l *Loadout) Spin() enums.SpinLevel            { return l.spin }
func (l *Loadout) Hamon() enums.HamonLevel          { return l.hamon }
func (l *Loadout) FruitMastery() enums.FruitMastery { return l.fruitMastery }
func (l *Loadout) ArmamentHaki() enums.HakiLevel    { return l.armamentHaki }
func (l *Loadout) ObservationHaki() enums.HakiLevel { return l.observationHaki }
func (l *Loadout) ConquerorHaki() enums.HakiLevel   { return l.conquerorHaki }
func (l *Loadout) PhysicalForm() enums.PhysicalForm { return l.physicalForm }
