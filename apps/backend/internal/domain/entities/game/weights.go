package game

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

// AssignmentWeights configures LoadoutBuilder's weighted random draws. All
// fields are relative weights (they need not sum to anything in
// particular) consumed via weightedPick. A weight of 0 for a given option
// means "never draw this"; a missing map entry is treated as 0.
//
// This is a plain value, not a port: the actual policy (probability of no
// stand, rarity skew, the reduced weight for Conqueror Haki, ...) is
// configured by an adapter behind ports.IAssignmentWeights, which the
// application layer resolves and passes in. DefaultAssignmentWeights is a
// uniform-ish reference implementation, useful until that adapter exists.
type AssignmentWeights struct {
	// NoStandWeight/NoDevilFruitWeight weigh drawing no stand / no devil
	// fruit at all, alongside the pool's per-rarity weights below.
	NoStandWeight      int
	NoDevilFruitWeight int

	// RarityWeight weighs which Stand/DevilFruit gets drawn, by rarity.
	RarityWeight map[enums.PowerRarity]int

	// Per-level weight tables, keyed by the level itself.
	SpinLevelWeights    map[enums.SpinLevel]int
	HamonLevelWeights   map[enums.HamonLevel]int
	FruitMasteryWeights map[enums.FruitMastery]int // FruitMasteryNone is ignored: NewLoadout forces it once a fruit is/isn't drawn.

	ArmamentHakiWeights    map[enums.HakiLevel]int
	ObservationHakiWeights map[enums.HakiLevel]int
	// ConquerorHakiWeights should skew towards HakiPrivate far more than
	// the other two haki tables - Conqueror Haki is meant to be rare.
	ConquerorHakiWeights map[enums.HakiLevel]int

	PhysicalFormWeights map[enums.PhysicalForm]int
}

// DefaultAssignmentWeights is a uniform reference table (equal weight per
// level/rarity), except Conqueror Haki, which is skewed low by design.
func DefaultAssignmentWeights() AssignmentWeights {
	return AssignmentWeights{
		NoStandWeight:      1,
		NoDevilFruitWeight: 1,
		RarityWeight: map[enums.PowerRarity]int{
			enums.Common:    4,
			enums.Rare:      3,
			enums.Epic:      2,
			enums.Legendary: 1,
		},
		SpinLevelWeights: map[enums.SpinLevel]int{
			enums.SpinNone:     1,
			enums.SpinBasic:    1,
			enums.SpinAdvanced: 1,
			enums.SpinGolden:   1,
			enums.SpinInfinite: 1,
		},
		HamonLevelWeights: map[enums.HamonLevel]int{
			enums.HamonNone:     1,
			enums.HamonBasic:    1,
			enums.HamonAdvanced: 1,
			enums.HamonPerfect:  1,
		},
		FruitMasteryWeights: map[enums.FruitMastery]int{
			enums.FruitMasteryRegular:  1,
			enums.FruitMasteryAdvanced: 1,
			enums.FruitMasteryAwakened: 1,
		},
		ArmamentHakiWeights: map[enums.HakiLevel]int{
			enums.HakiPrivate:        1,
			enums.HakiViceAdmiral:    1,
			enums.HakiYonkoCommander: 1,
			enums.HakiYonkoPlus:      1,
		},
		ObservationHakiWeights: map[enums.HakiLevel]int{
			enums.HakiPrivate:        1,
			enums.HakiViceAdmiral:    1,
			enums.HakiYonkoCommander: 1,
			enums.HakiYonkoPlus:      1,
		},
		ConquerorHakiWeights: map[enums.HakiLevel]int{
			enums.HakiPrivate:        6,
			enums.HakiViceAdmiral:    2,
			enums.HakiYonkoCommander: 1,
			enums.HakiYonkoPlus:      1,
		},
		PhysicalFormWeights: map[enums.PhysicalForm]int{
			enums.PhysicalFormPrivate:        1,
			enums.PhysicalFormViceAdmiral:    1,
			enums.PhysicalFormYonkoCommander: 1,
			enums.PhysicalFormYonkoPlus:      1,
		},
	}
}
