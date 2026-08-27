package game

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

// AssignmentWeights configures LoadoutBuilder's weighted random draws. All
// fields are relative weights (they need not sum to anything in
// particular) consumed via weightedPick. A weight of 0 for a given option
// means "never draw this"; a missing map entry is treated as 0.
//
// This is a plain value, not a port: the actual policy is configured by an
// adapter behind ports.IAssignmentWeights, which the application layer
// resolves and passes in. DefaultAssignmentWeights ports the probabilities
// of the original JoJoOnePiece_Simulator
// (github.com/oorbea/JoJoOnePiece_Simulator) 1:1, and is the only
// implementation until an admin-configurable adapter exists. Note what it
// deliberately does NOT weigh: which Stand/DevilFruit gets drawn is a
// uniform pick over the pool by LoadoutBuilder itself (see
// loadout_builder.go's drawStand/drawDevilFruit) - rarity has no bearing on
// random assignment, matching V1's uniform_int_distribution over the whole
// power array. Rarity is meant to matter only for the (not yet built)
// gachapon boxes - see ObsidianVault/gameplay-versus-inventory-characters.md.
type AssignmentWeights struct {
	// NoStandWeight/NoDevilFruitWeight weigh drawing no stand / no devil
	// fruit at all, alongside a uniform weight of 1 for every candidate in
	// the pool.
	NoStandWeight      int
	NoDevilFruitWeight int

	// Per-level weight tables, keyed by the level itself.
	SpinLevelWeights    map[enums.SpinLevel]int
	HamonLevelWeights   map[enums.HamonLevel]int
	FruitMasteryWeights map[enums.FruitMastery]int // FruitMasteryNone is ignored: NewLoadout forces it once a fruit is/isn't drawn.

	// HakiSetWeights weighs which combination of the three haki types
	// (Armament/Observation/Conqueror) a player has at all, drawn as a
	// single correlated pick - see HakiSet and its doc comment.
	HakiSetWeights map[HakiSet]int

	// HakiMasteryWeights weighs the mastery level drawn independently for
	// each haki type the player has, per HakiSetWeights. HakiNone is
	// ignored here - a present haki is always at least HakiPrivate.
	HakiMasteryWeights map[enums.HakiLevel]int

	PhysicalFormWeights map[enums.PhysicalForm]int
}

// HakiSet names one of the eight combinations of Armament/Observation/
// Conqueror haki a player can be drawn to have (or not) - ported from
// JoJoOnePiece_Simulator's generateHaki (powers.cc), which draws the *set*
// of hakis a player has as one correlated pick (Conqueror rarely appears
// alone or without at least one of the other two) rather than three
// independent per-type draws.
type HakiSet byte

const (
	HakiSetNone HakiSet = iota
	HakiSetArmament
	HakiSetObservation
	HakiSetArmamentObservation
	HakiSetArmamentObservationConqueror
	HakiSetArmamentConqueror
	HakiSetObservationConqueror
	HakiSetConqueror
)

// HasArmament, HasObservation and HasConqueror report whether this HakiSet
// includes the given haki type.
func (s HakiSet) HasArmament() bool {
	switch s {
	case HakiSetArmament, HakiSetArmamentObservation, HakiSetArmamentObservationConqueror, HakiSetArmamentConqueror:
		return true
	default:
		return false
	}
}

func (s HakiSet) HasObservation() bool {
	switch s {
	case HakiSetObservation, HakiSetArmamentObservation, HakiSetArmamentObservationConqueror, HakiSetObservationConqueror:
		return true
	default:
		return false
	}
}

func (s HakiSet) HasConqueror() bool {
	switch s {
	case HakiSetArmamentObservationConqueror, HakiSetArmamentConqueror, HakiSetObservationConqueror, HakiSetConqueror:
		return true
	default:
		return false
	}
}

// hakiSets is every HakiSet value, in the fixed order weightedPick indexes
// into - must stay in sync with DefaultAssignmentWeights.HakiSetWeights and
// with drawHakiSet in loadout_builder.go.
var hakiSets = []HakiSet{
	HakiSetNone,
	HakiSetArmament,
	HakiSetObservation,
	HakiSetArmamentObservation,
	HakiSetArmamentObservationConqueror,
	HakiSetArmamentConqueror,
	HakiSetObservationConqueror,
	HakiSetConqueror,
}

// DefaultAssignmentWeights ports JoJoOnePiece_Simulator V1's probability
// tables 1:1 (github.com/oorbea/JoJoOnePiece_Simulator, powers.cc):
// generateStand/generateFruit are uniform over the pool (handled outside
// this table, see the struct doc), generateStrength is uniform 1/6,
// generateSpin is 15/30/30/25 over its 4 levels, generateHaki draws the set
// with the 4/20/20/20/15/10/10/1 (%) table before generateHakiMastery
// draws a uniform 1/4 mastery per haki present. Hamon has no V1 equivalent
// (V1 predates it) - its weights were chosen by the owner directly:
// 25/35/35/5.
func DefaultAssignmentWeights() AssignmentWeights {
	return AssignmentWeights{
		NoStandWeight:      1,
		NoDevilFruitWeight: 1,
		SpinLevelWeights: map[enums.SpinLevel]int{
			enums.SpinNone:     15,
			enums.SpinBasic:    30,
			enums.SpinGolden:   30,
			enums.SpinInfinite: 25,
		},
		HamonLevelWeights: map[enums.HamonLevel]int{
			enums.HamonNone:     25,
			enums.HamonBasic:    35,
			enums.HamonAdvanced: 35,
			enums.HamonPerfect:  5,
		},
		FruitMasteryWeights: map[enums.FruitMastery]int{
			enums.FruitMasteryRegular:  1,
			enums.FruitMasteryAdvanced: 1,
			enums.FruitMasteryAwakened: 1,
		},
		HakiSetWeights: map[HakiSet]int{
			HakiSetNone:                         4,
			HakiSetArmament:                     20,
			HakiSetObservation:                  20,
			HakiSetArmamentObservation:          20,
			HakiSetArmamentObservationConqueror: 15,
			HakiSetArmamentConqueror:            10,
			HakiSetObservationConqueror:         10,
			HakiSetConqueror:                    1,
		},
		HakiMasteryWeights: map[enums.HakiLevel]int{
			enums.HakiPrivate:        1,
			enums.HakiViceAdmiral:    1,
			enums.HakiYonkoCommander: 1,
			enums.HakiYonkoPlus:      1,
		},
		PhysicalFormWeights: map[enums.PhysicalForm]int{
			enums.PhysicalFormPrivate:        1,
			enums.PhysicalFormStrongFishman:  1,
			enums.PhysicalFormMarineCaptain:  1,
			enums.PhysicalFormViceAdmiral:    1,
			enums.PhysicalFormYonkoCommander: 1,
			enums.PhysicalFormYonkoPlus:      1,
		},
	}
}
