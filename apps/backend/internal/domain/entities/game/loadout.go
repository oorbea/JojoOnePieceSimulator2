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
)

// Loadout is the immutable set of abilities assigned to a Participant for a
// game (Gauntlet) or a single round (Versus). Spin and Hamon are
// independent of the Stand and of each other - a player can hold any
// combination of the three - except where a power demands a floor (see
// power_effects.go: Tusk: Acto 4 forces SpinInfinite, a Zoan raises Physical
// Form, ...). FruitMastery is coupled to DevilFruit: no fruit forces
// FruitMasteryNone, any fruit forces at least FruitMasteryRegular.
// NewLoadout enforces both regardless of how the values were produced
// (random draw or, later, inventory). effects records what the builder's
// resolver did on the way - the Stand/DevilFruit held here are the final
// ones, after any evolution.
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
	effects         []PowerEffect
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
	// Effects is what the LoadoutBuilder's resolver did, in order. Empty for
	// a Loadout assembled by hand.
	Effects []PowerEffect
	// Mangas are the mangas in play, so cross-manga floors are only enforced
	// when both are. Nil (tests, hand-built loadouts) means "not both".
	Mangas []enums.Manga
}

// NewLoadoutFromSpec validates and builds a Loadout from spec. BattleIQ
// carries no manga gate here - Loadout itself doesn't know which mangas are
// in play (neither does Spin/Hamon), so the JoJo-only gate lives in
// LoadoutBuilder instead. Power-effect floors are enforced too (see
// ErrPowerEffectFloorViolated); restoring a persisted Loadout skips them, see
// newLoadout.
func NewLoadoutFromSpec(spec LoadoutSpec) (*Loadout, error) {
	return newLoadout(spec, true)
}

// newLoadout is NewLoadoutFromSpec with the power-effect floor check made
// optional. Restore passes false: a snapshot already holds an assigned
// Loadout, and re-judging it against today's rule table would make every
// in-flight game assigned under an older table fail to restore.
func newLoadout(spec LoadoutSpec, enforceFloors bool) (*Loadout, error) {
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
	if enforceFloors && newEffectStateFromSpec(spec).floorsViolated() {
		return nil, ErrPowerEffectFloorViolated
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
		effects:         append([]PowerEffect(nil), spec.Effects...),
	}, nil
}

// newEffectStateFromSpec builds the resolver's state from a spec, without any
// evolution family - enough to judge floors, which need no evolution targets.
func newEffectStateFromSpec(spec LoadoutSpec) *effectState {
	st := &effectState{stand: spec.Stand, fruit: spec.DevilFruit}
	st.values[enums.SlotPhysicalForm] = int(spec.PhysicalForm)
	st.values[enums.SlotFruitMastery] = int(spec.FruitMastery)
	st.values[enums.SlotHamon] = int(spec.Hamon)
	st.values[enums.SlotArmamentHaki] = int(spec.ArmamentHaki)
	st.values[enums.SlotObservationHaki] = int(spec.ObservationHaki)
	st.values[enums.SlotConquerorHaki] = int(spec.ConquerorHaki)
	st.values[enums.SlotSpin] = int(spec.Spin)
	var jojo, onePiece bool
	for _, m := range spec.Mangas {
		switch m {
		case enums.Jojo:
			jojo = true
		case enums.OnePiece:
			onePiece = true
		}
	}
	st.bothMangas = jojo && onePiece
	return st
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

// Effects returns a copy of what the power-effect resolver did to this
// Loadout during assignment, in order (empty if nothing).
func (l *Loadout) Effects() []PowerEffect { return append([]PowerEffect(nil), l.effects...) }

// DrawnStand is the Stand as it was drawn, before any evolution effect: the
// final Stand itself if it never evolved, else the ancestor the first Stand
// evolution started from. The sorteo reveals this one first.
func (l *Loadout) DrawnStand() *powers.Stand {
	if l.stand == nil {
		return nil
	}
	for _, e := range l.effects {
		if e.Kind != enums.EffectEvolution || e.Slot != enums.SlotStand {
			continue
		}
		for cur := l.stand; cur != nil; cur = cur.EvolvesFrom() {
			if cur.ID().String() == e.From {
				return cur
			}
		}
		break
	}
	return l.stand
}

// DrawnDevilFruit is DrawnStand for the DevilFruit.
func (l *Loadout) DrawnDevilFruit() *powers.DevilFruit {
	if l.devilFruit == nil {
		return nil
	}
	for _, e := range l.effects {
		if e.Kind != enums.EffectEvolution || e.Slot != enums.SlotDevilFruit {
			continue
		}
		for cur := l.devilFruit; cur != nil; cur = cur.EvolvesFrom() {
			if cur.ID().String() == e.From {
				return cur
			}
		}
		break
	}
	return l.devilFruit
}
