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
