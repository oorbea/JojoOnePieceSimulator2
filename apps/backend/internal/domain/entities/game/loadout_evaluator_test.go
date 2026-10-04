package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// TestDefaultLoadoutEvaluator_RarityBonus pins the Rare/Epic/Legendary/
// Mythical bonus scale (0/1/2/4/8) - a Legendary is no longer just +1 over
// Epic, and Mythical is meant to weigh twice a Legendary.
func TestDefaultLoadoutEvaluator_RarityBonus(t *testing.T) {
	eval := game.DefaultLoadoutEvaluator{}

	baseline, err := game.NewLoadout(nil, nil, enums.SpinNone, enums.HamonNone, enums.FruitMasteryNone, enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate)
	if err != nil {
		t.Fatalf("NewLoadout(baseline): %v", err)
	}
	baseScore := eval.Score(baseline)

	cases := []struct {
		rarity enums.PowerRarity
		bonus  int
	}{
		{enums.Common, 0},
		{enums.Rare, 1},
		{enums.Epic, 2},
		{enums.Legendary, 4},
		{enums.Mythical, 8},
	}
	for _, tc := range cases {
		fruit := mustDevilFruit(t, 1, "Test Fruit "+tc.rarity.String(), tc.rarity, enums.Paramecia)
		l, err := game.NewLoadout(nil, fruit, enums.SpinNone, enums.HamonNone, enums.FruitMasteryRegular, enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate)
		if err != nil {
			t.Fatalf("NewLoadout(%s): %v", tc.rarity, err)
		}
		got := eval.Score(l) - baseScore - 1 /* FruitMasteryRegular weight */
		if got != tc.bonus {
			t.Fatalf("rarity %s: expected bonus %d, got %d", tc.rarity, tc.bonus, got)
		}
	}
}

// TestDefaultLoadoutEvaluator_BattleIQScore pins the WAIS-IV band -> 0..6
// score mapping - moderate and sublinear, never the raw 0-255 value (a 255
// would otherwise be ~5x the rest of the score combined).
func TestDefaultLoadoutEvaluator_BattleIQScore(t *testing.T) {
	eval := game.DefaultLoadoutEvaluator{}

	baseline, err := game.NewLoadout(nil, nil, enums.SpinNone, enums.HamonNone, enums.FruitMasteryNone, enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate)
	if err != nil {
		t.Fatalf("NewLoadout(baseline): %v", err)
	}
	if baseline.BattleIQ().Present() {
		t.Fatalf("NewLoadout should leave BattleIQ absent")
	}
	baseScore := eval.Score(baseline)

	cases := []struct {
		value byte
		want  int
	}{
		{0, 0},   // ExtremelyLow
		{75, 1},  // Borderline
		{85, 2},  // LowAverage
		{100, 3}, // Average
		{115, 4}, // HighAverage
		{125, 5}, // Superior
		{255, 6}, // VerySuperior - the extreme, still capped at 6
	}
	for _, tc := range cases {
		l, err := game.NewLoadoutFromSpec(game.LoadoutSpec{
			Spin: enums.SpinNone, Hamon: enums.HamonNone, FruitMastery: enums.FruitMasteryNone,
			ArmamentHaki: enums.HakiPrivate, ObservationHaki: enums.HakiPrivate, ConquerorHaki: enums.HakiPrivate,
			PhysicalForm: enums.PhysicalFormPrivate,
			BattleIQ:     game.NewBattleIQ(tc.value),
		})
		if err != nil {
			t.Fatalf("NewLoadoutFromSpec(battleIQ=%d): %v", tc.value, err)
		}
		got := eval.Score(l) - baseScore
		if got != tc.want {
			t.Fatalf("battleIQ %d: expected score contribution %d, got %d", tc.value, tc.want, got)
		}
	}
}

// TestDefaultLoadoutEvaluator_AbilityWeights pins every ability weight table.
// Each case varies one slot from an all-zero baseline and expects exactly the
// table value, so a reshuffled enum can't silently rebalance the score.
func TestDefaultLoadoutEvaluator_AbilityWeights(t *testing.T) {
	eval := game.DefaultLoadoutEvaluator{}
	build := func(spin enums.SpinLevel, hamon enums.HamonLevel, m enums.FruitMastery, arm, obs, con enums.HakiLevel, form enums.PhysicalForm) int {
		l, err := game.NewLoadout(nil, nil, spin, hamon, m, arm, obs, con, form)
		if err != nil {
			t.Fatalf("NewLoadout: %v", err)
		}
		return eval.Score(l)
	}
	base := build(enums.SpinNone, enums.HamonNone, enums.FruitMasteryNone, enums.HakiNone, enums.HakiNone, enums.HakiNone, enums.PhysicalFormPrivate)
	if base != 0 {
		t.Fatalf("all-zero baseline: expected 0, got %d", base)
	}

	hakis := []enums.HakiLevel{enums.HakiNone, enums.HakiPrivate, enums.HakiViceAdmiral, enums.HakiYonkoCommander, enums.HakiYonkoPlus}
	forms := []enums.PhysicalForm{enums.PhysicalFormPrivate, enums.PhysicalFormStrongFishman, enums.PhysicalFormMarineCaptain, enums.PhysicalFormViceAdmiral, enums.PhysicalFormYonkoCommander, enums.PhysicalFormYonkoPlus}
	spins := []enums.SpinLevel{enums.SpinNone, enums.SpinBasic, enums.SpinGolden, enums.SpinInfinite}
	hamons := []enums.HamonLevel{enums.HamonNone, enums.HamonBasic, enums.HamonAdvanced, enums.HamonPerfect}
	masteries := []enums.FruitMastery{enums.FruitMasteryNone, enums.FruitMasteryRegular, enums.FruitMasteryAdvanced, enums.FruitMasteryAwakened}

	for i, want := range []int{0, 1, 3, 6} {
		if got := build(spins[i], enums.HamonNone, enums.FruitMasteryNone, enums.HakiNone, enums.HakiNone, enums.HakiNone, enums.PhysicalFormPrivate); got != want {
			t.Errorf("spin %s: expected %d, got %d", spins[i], want, got)
		}
	}
	// Hamon carries power-effect floors (Advanced+ needs MarineCaptain form,
	// Perfect also needs Spin Basic and Armament Private), so build with the
	// companions it forces and expect hamon weight + their weights.
	hamonCases := []struct {
		spin enums.SpinLevel
		arm  enums.HakiLevel
		form enums.PhysicalForm
		want int
	}{
		{enums.SpinNone, enums.HakiNone, enums.PhysicalFormPrivate, 0},
		{enums.SpinNone, enums.HakiNone, enums.PhysicalFormPrivate, 1},
		{enums.SpinNone, enums.HakiNone, enums.PhysicalFormMarineCaptain, 2 + 2},
		{enums.SpinBasic, enums.HakiPrivate, enums.PhysicalFormMarineCaptain, 3 + 1 + 1 + 2},
	}
	for i, tc := range hamonCases {
		if got := build(tc.spin, hamons[i], enums.FruitMasteryNone, tc.arm, enums.HakiNone, enums.HakiNone, tc.form); got != tc.want {
			t.Errorf("hamon %s: expected %d, got %d", hamons[i], tc.want, got)
		}
	}
	// Mastery needs a Devil Fruit (None only without one, >= Regular with
	// one); a Common fruit adds no rarity bonus, so the score is the weight.
	fruit := mustDevilFruit(t, 1, "Weight Fruit", enums.Common, enums.Paramecia)
	for i, want := range []int{1, 3, 5} {
		m := masteries[i+1]
		l, err := game.NewLoadout(nil, fruit, enums.SpinNone, enums.HamonNone, m, enums.HakiNone, enums.HakiNone, enums.HakiNone, enums.PhysicalFormPrivate)
		if err != nil {
			t.Fatalf("NewLoadout(mastery %s): %v", m, err)
		}
		if got := eval.Score(l); got != want {
			t.Errorf("mastery %s: expected %d, got %d", m, want, got)
		}
	}
	for i, want := range []int{0, 1, 2, 3, 4} {
		if got := build(enums.SpinNone, enums.HamonNone, enums.FruitMasteryNone, hakis[i], enums.HakiNone, enums.HakiNone, enums.PhysicalFormPrivate); got != want {
			t.Errorf("armament haki %s: expected %d, got %d", hakis[i], want, got)
		}
		if got := build(enums.SpinNone, enums.HamonNone, enums.FruitMasteryNone, enums.HakiNone, hakis[i], enums.HakiNone, enums.PhysicalFormPrivate); got != want {
			t.Errorf("observation haki %s: expected %d, got %d", hakis[i], want, got)
		}
	}
	for i, want := range []int{0, 1, 2, 4, 6} {
		if got := build(enums.SpinNone, enums.HamonNone, enums.FruitMasteryNone, enums.HakiNone, enums.HakiNone, hakis[i], enums.PhysicalFormPrivate); got != want {
			t.Errorf("conqueror haki %s: expected %d, got %d", hakis[i], want, got)
		}
	}
	for i, want := range []int{0, 1, 2, 3, 4, 5} {
		if got := build(enums.SpinNone, enums.HamonNone, enums.FruitMasteryNone, enums.HakiNone, enums.HakiNone, enums.HakiNone, forms[i]); got != want {
			t.Errorf("physical form %s: expected %d, got %d", forms[i], want, got)
		}
	}
}
