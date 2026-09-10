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
		got := eval.Score(l) - baseScore - int(enums.FruitMasteryRegular)
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
