package game

import (
	"sort"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// Combat conventions are the voters' rulebook for "who can see / touch /
// beat whom". They are pure data: nothing in the game resolves a round from
// them (the vote is human), but the manual renders them straight from this
// table, so what it says is by construction what the project considers true.
// A future LLM tiebreaker or bot heuristic can read the same table.
//
// Stand/fruit names are the catalogue's display names; they are matched
// exactly after normalizePowerName, like every other name-keyed rule (see
// power_effects.go), and db/migrations/power_effect_names_test.go checks them
// against the seeds via PowerEffectRuleNames.

// ConventionCategory groups conventions in the manual.
type ConventionCategory string

const (
	CategoryVisibility    ConventionCategory = "VISIBILITY"
	CategoryContact       ConventionCategory = "CONTACT"
	CategoryTime          ConventionCategory = "TIME"
	CategoryWill          ConventionCategory = "WILL"
	CategoryWeakness      ConventionCategory = "WEAKNESS"
	CategoryNullification ConventionCategory = "NULLIFICATION"
	CategorySupremacy     ConventionCategory = "SUPREMACY"
)

// SubjectKind says how a Subject picks who it refers to.
type SubjectKind string

const (
	// SubjectAnyStand: whoever holds any Stand.
	SubjectAnyStand SubjectKind = "ANY_STAND"
	// SubjectAnyFruit: whoever holds any Devil Fruit.
	SubjectAnyFruit SubjectKind = "ANY_FRUIT"
	// SubjectFruitType: whoever holds a fruit of one of FruitTypes.
	SubjectFruitType SubjectKind = "FRUIT_TYPE"
	// SubjectSlotMin: whoever has Slot at Min (a level wire string) or above.
	SubjectSlotMin SubjectKind = "SLOT_MIN"
	// SubjectPowers: whoever holds one of the named Stands/Fruits. Group, when
	// set, names a set the manual labels as a whole.
	SubjectPowers SubjectKind = "POWERS"
)

// Subject is one "who" of a convention. Only the fields its Kind uses are set.
type Subject struct {
	Kind       SubjectKind
	Slot       enums.LoadoutSlot // SubjectSlotMin
	Min        string            // SubjectSlotMin: level wire string, e.g. "GOLDEN"
	MinRank    int               // SubjectSlotMin: the level's ordinal, what rules compare
	FruitTypes []enums.FruitType // SubjectFruitType
	Names      []string          // SubjectPowers: catalogue display names
	Group      string            // SubjectPowers: optional set label
}

// Named groups of powers that several conventions share.
const (
	GroupTimeStoppers = "TIME_STOPPERS"
	GroupAbsolutes    = "ABSOLUTE_STANDS"
)

// Convention is one rule. Actors are the "anyone of" who do it (or who can);
// Targets, when set, who it acts on (empty: anyone/anything); Immune who it
// cannot affect; Attenuated who only feel it one step weaker. Gap is the level
// difference a Conqueror's Haki convention needs, 0 elsewhere. Judgement marks
// a convention the voters settle case by case rather than a hard rule.
type Convention struct {
	ID         string
	Category   ConventionCategory
	Actors     []Subject
	Targets    []Subject
	Immune     []Subject
	Attenuated []Subject
	Gap        int
	Judgement  bool
}

// Catalogue display names the conventions refer to.
const (
	convTheWorld       = "The World"
	convStarPlatinumTW = "Star Platinum: The World"
	convTheWorldSBR    = "The World (Steel Ball Run)"
	convKingCrimson    = "King Crimson"
	convMadeInHeaven   = "Made in Heaven"
	convGER            = "Gold Experience: Requiem"
	convWonderOfU      = "Wonder of U"
	convLoveTrain      = "D4C: Love Train"
	convTuskAct4       = "Tusk: Act 4"
	convGoBeyond       = "Soft & Wet: Go Beyond"
	convWeatherReport  = "Weather Report"
	convYamiYami       = "Yami Yami no mi"
	convHermitPurple   = "Hermit Purple"
	convNika           = "Hito Hito no mi: Model Nika"
	convGomuGomu       = "Gomu Gomu no mi"
)

func anyStand() Subject { return Subject{Kind: SubjectAnyStand} }
func anyFruit() Subject { return Subject{Kind: SubjectAnyFruit} }

func fruitTypes(t ...enums.FruitType) Subject {
	return Subject{Kind: SubjectFruitType, FruitTypes: t}
}

// slotMin is "has slot at rank (the level's ordinal) or above"; Min carries the
// level's wire string for the manual.
func slotMin(slot enums.LoadoutSlot, rank int) Subject {
	return Subject{Kind: SubjectSlotMin, Slot: slot, Min: slotLabel(slot, rank), MinRank: rank}
}

func powersNamed(group string, names ...string) Subject {
	return Subject{Kind: SubjectPowers, Group: group, Names: names}
}

func timeStoppers() Subject {
	return powersNamed(GroupTimeStoppers, convTheWorld, convStarPlatinumTW, convTheWorldSBR, convKingCrimson)
}

func absoluteStands() Subject {
	return powersNamed(GroupAbsolutes, convGER, convWonderOfU, convLoveTrain, convTuskAct4)
}

// CombatConventions returns the full table, built fresh on every call.
func CombatConventions() []Convention {
	spinBasic := slotMin(enums.SlotSpin, int(enums.SpinBasic))
	spinGolden := slotMin(enums.SlotSpin, int(enums.SpinGolden))
	spinInfinite := slotMin(enums.SlotSpin, int(enums.SpinInfinite))
	armament := slotMin(enums.SlotArmamentHaki, int(enums.HakiPrivate))
	conqueror := slotMin(enums.SlotConquerorHaki, int(enums.HakiPrivate))
	observation := slotMin(enums.SlotObservationHaki, int(enums.HakiPrivate))
	hamon := slotMin(enums.SlotHamon, int(enums.HamonBasic))
	hamonPerfect := slotMin(enums.SlotHamon, int(enums.HamonPerfect))
	logia := fruitTypes(enums.Logia)
	madeInHeaven := powersNamed("", convMadeInHeaven)
	conquerorYonkoPlus := slotMin(enums.SlotConquerorHaki, int(enums.HakiYonkoPlus))
	observationYonkoPlus := slotMin(enums.SlotObservationHaki, int(enums.HakiYonkoPlus))

	return []Convention{
		{ID: "SEE_STANDS", Category: CategoryVisibility,
			Actors: []Subject{anyStand(), spinBasic, observation}, Targets: []Subject{anyStand()}},

		{ID: "TOUCH_LOGIA", Category: CategoryContact,
			Actors: []Subject{armament, anyStand(), spinGolden}, Targets: []Subject{logia}},
		{ID: "TOUCH_LOGIA_HAMON", Category: CategoryContact, Judgement: true,
			Actors: []Subject{hamon}, Targets: []Subject{logia}},
		{ID: "DAMAGE_STAND", Category: CategoryContact,
			Actors: []Subject{anyStand(), armament, conqueror, logia, spinBasic}, Targets: []Subject{anyStand()}},

		{ID: "TIME_STOP", Category: CategoryTime,
			Actors: []Subject{timeStoppers()}},
		{ID: "MOVE_IN_TIME_STOP", Category: CategoryTime,
			Actors: []Subject{timeStoppers(), spinInfinite, conquerorYonkoPlus}, Targets: []Subject{timeStoppers()}},
		{ID: "MADE_IN_HEAVEN_STOPPED_BY", Category: CategoryTime,
			Actors: []Subject{timeStoppers()}, Targets: []Subject{madeInHeaven}},
		{ID: "MADE_IN_HEAVEN_KEEP_PACE", Category: CategoryTime,
			Actors: []Subject{spinInfinite}, Targets: []Subject{madeInHeaven}},
		{ID: "MADE_IN_HEAVEN_SLOWED_BY", Category: CategoryTime,
			Actors: []Subject{conquerorYonkoPlus}, Targets: []Subject{madeInHeaven}},
		{ID: "MADE_IN_HEAVEN_TRACKED_BY", Category: CategoryTime,
			Actors: []Subject{observationYonkoPlus}, Targets: []Subject{madeInHeaven}},

		{ID: "CONQUEROR_KNOCKOUT", Category: CategoryWill, Gap: 2,
			Actors: []Subject{conqueror}, Immune: []Subject{spinInfinite}, Attenuated: []Subject{hamonPerfect}},
		{ID: "CONQUEROR_HINDER", Category: CategoryWill, Gap: 1,
			Actors: []Subject{conqueror}, Immune: []Subject{spinInfinite}, Attenuated: []Subject{hamonPerfect}},

		{ID: "WATER_WEAKNESS", Category: CategoryWeakness,
			Actors: []Subject{powersNamed("", convWeatherReport)}, Targets: []Subject{anyFruit()}},

		{ID: "YAMI_NULLIFIES", Category: CategoryNullification,
			Actors:  []Subject{powersNamed("", convYamiYami)},
			Targets: []Subject{anyFruit(), anyStand()},
			Immune: []Subject{
				absoluteStands(), powersNamed("", convGoBeyond), timeStoppers(), madeInHeaven,
			}},

		{ID: "SPIN_INFINITE_SUPREME", Category: CategorySupremacy,
			Actors: []Subject{spinInfinite},
			Immune: []Subject{powersNamed("", convGER), spinInfinite, powersNamed("", convWonderOfU)}},
		// Two INFINITE Spins: no obvious winner, the voters decide.
		{ID: "SPIN_INFINITE_DUEL", Category: CategorySupremacy, Judgement: true,
			Actors: []Subject{spinInfinite}, Targets: []Subject{spinInfinite}},

		// Who can overcome each absolute Stand (Actors), and nobody else.
		{ID: "ABSOLUTE_GER", Category: CategorySupremacy,
			Targets: []Subject{powersNamed("", convGER)},
			Actors:  []Subject{powersNamed("", convGoBeyond)}},
		{ID: "ABSOLUTE_WONDER_OF_U", Category: CategorySupremacy,
			Targets: []Subject{powersNamed("", convWonderOfU)},
			Actors:  []Subject{powersNamed("", convGoBeyond, convGER)}},
		{ID: "ABSOLUTE_LOVE_TRAIN", Category: CategorySupremacy,
			Targets: []Subject{powersNamed("", convLoveTrain)},
			Actors:  []Subject{spinInfinite, powersNamed("", convGER, convGoBeyond, convWonderOfU)}},
		{ID: "ABSOLUTE_TUSK_ACT_4", Category: CategorySupremacy,
			Targets: []Subject{powersNamed("", convTuskAct4)},
			Actors:  []Subject{powersNamed("", convGER, convWonderOfU)}},
	}
}

// ConventionPowerNames lists every power name (normalized, sorted, unique) the
// conventions refer to, for the catalogue test that checks they exist.
func ConventionPowerNames() []string {
	set := map[string]struct{}{}
	add := func(subjects []Subject) {
		for _, s := range subjects {
			for _, n := range s.Names {
				set[normalizePowerName(n)] = struct{}{}
			}
		}
	}
	for _, c := range CombatConventions() {
		add(c.Actors)
		add(c.Targets)
		add(c.Immune)
		add(c.Attenuated)
	}
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
