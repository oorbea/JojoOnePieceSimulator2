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
	battleIQ        BattleIQ
}

// LoadoutSpec is the full set of fields NewLoadoutFromSpec validates and
// assembles into a Loadout. It exists so new fields (like BattleIQ) don't
// grow NewLoadout's parameter list - a Loadout already has three adjacent
// enums.HakiLevel parameters that a positional call site can silently
// transpose; a tenth positional parameter would only make that worse.
type LoadoutSpec struct {
	Stand           *powers.Stand
	DevilFruit      *powers.DevilFruit
	Spin            enums.SpinLevel
	Hamon           enums.HamonLevel
	FruitMastery    enums.FruitMastery
	ArmamentHaki    enums.HakiLevel
	ObservationHaki enums.HakiLevel
	ConquerorHaki   enums.HakiLevel
	PhysicalForm    enums.PhysicalForm
	BattleIQ        BattleIQ // zero value (NoBattleIQ()) is fine for a non-JoJo lobby
}

// NewLoadoutFromSpec validates and builds a Loadout from spec. BattleIQ
// carries no manga gate here - Loadout itself doesn't know which mangas are
// in play (neither does Spin/Hamon), so the JoJo-only gate lives in
// LoadoutBuilder instead.
func NewLoadoutFromSpec(spec LoadoutSpec) (*Loadout, error) {
	if !spec.Spin.IsValid() {
		return nil, enums.ErrInvalidSpinLevel
	}
	if !spec.Hamon.IsValid() {
		return nil, enums.ErrInvalidHamonLevel
	}
	if !spec.FruitMastery.IsValid() {
		return nil, enums.ErrInvalidFruitMastery
	}
	if !spec.ArmamentHaki.IsValid() {
		return nil, enums.ErrInvalidHakiLevel
	}
	if !spec.ObservationHaki.IsValid() {
		return nil, enums.ErrInvalidHakiLevel
	}
	if !spec.ConquerorHaki.IsValid() {
		return nil, enums.ErrInvalidHakiLevel
	}
	if !spec.PhysicalForm.IsValid() {
		return nil, enums.ErrInvalidPhysicalForm
	}
	if spec.DevilFruit == nil && spec.FruitMastery != enums.FruitMasteryNone {
		return nil, ErrFruitMasteryMismatch
	}
	if spec.DevilFruit != nil && spec.FruitMastery == enums.FruitMasteryNone {
		return nil, ErrFruitMasteryMismatch
	}
	if spec.Stand != nil && HasTrait(&spec.Stand.Power, enums.RequiresSpin4) && spec.Spin != enums.SpinInfinite {
		return nil, ErrSpin4Required
	}
	return &Loadout{
		stand:           spec.Stand,
		devilFruit:      spec.DevilFruit,
		spin:            spec.Spin,
		hamon:           spec.Hamon,
		fruitMastery:    spec.FruitMastery,
		armamentHaki:    spec.ArmamentHaki,
		observationHaki: spec.ObservationHaki,
		conquerorHaki:   spec.ConquerorHaki,
		physicalForm:    spec.PhysicalForm,
		battleIQ:        spec.BattleIQ,
	}, nil
}

// NewLoadout validates and builds a Loadout. Pass nil for stand/devilFruit
// when the player draws neither ability. Thin wrapper over
// NewLoadoutFromSpec for the many existing 9-argument call sites; BattleIQ
// is always absent through this constructor.
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
	return NewLoadoutFromSpec(LoadoutSpec{
		Stand:           stand,
		DevilFruit:      devilFruit,
		Spin:            spin,
		Hamon:           hamon,
		FruitMastery:    fruitMastery,
		ArmamentHaki:    armamentHaki,
		ObservationHaki: observationHaki,
		ConquerorHaki:   conquerorHaki,
		PhysicalForm:    physicalForm,
	})
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
func (l *Loadout) BattleIQ() BattleIQ               { return l.battleIQ }
