package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestLoadoutBuilder_JojoOnlyDrawsNoOnePieceAbilities(t *testing.T) {
	stand := mustStand(t, 1, "Star Platinum", enums.Legendary)
	pool := game.NewAvailablePowers([]*powers.Stand{stand}, nil)
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.Jojo}, weights, &fakeRandom{seq: []int{1}})

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.DevilFruit() != nil || loadout.FruitMastery() != enums.FruitMasteryNone {
		t.Fatalf("expected no one piece abilities, got fruit=%v mastery=%v", loadout.DevilFruit(), loadout.FruitMastery())
	}
	if loadout.ArmamentHaki() != enums.HakiPrivate || loadout.ObservationHaki() != enums.HakiPrivate ||
		loadout.ConquerorHaki() != enums.HakiPrivate || loadout.PhysicalForm() != enums.PhysicalFormPrivate {
		t.Fatalf("expected zero-value one piece stats, got %+v", loadout)
	}
}

func TestLoadoutBuilder_OnePieceOnlyDrawsNoJojoAbilities(t *testing.T) {
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.OnePiece}, weights, &fakeRandom{})
	pool := game.NewAvailablePowers(nil, nil)

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.Stand() != nil || loadout.Spin() != enums.SpinNone || loadout.Hamon() != enums.HamonNone {
		t.Fatalf("expected no jojo abilities, got %+v", loadout)
	}
}

func TestLoadoutBuilder_MixedAssignsBothMangas(t *testing.T) {
	stand := mustStand(t, 1, "Star Platinum", enums.Legendary)
	fruit := mustDevilFruit(t, 2, "Gomu Gomu no Mi", enums.Legendary, enums.Paramecia)
	pool := game.NewAvailablePowers([]*powers.Stand{stand}, []*powers.DevilFruit{fruit})
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder(enums.Mangas(), weights, &fakeRandom{seq: []int{1}})

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.Stand() == nil || loadout.DevilFruit() == nil {
		t.Fatalf("expected both a stand and a devil fruit, got %+v", loadout)
	}
}

func TestLoadoutBuilder_FruitMasteryFollowsFruitPresence(t *testing.T) {
	fruit := mustDevilFruit(t, 1, "Gomu Gomu no Mi", enums.Legendary, enums.Paramecia)
	weights := game.DefaultAssignmentWeights()
	// index 0 is "no fruit"; index 1 is the only seeded fruit.
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.OnePiece}, weights, &fakeRandom{seq: []int{1}})
	pool := game.NewAvailablePowers(nil, []*powers.DevilFruit{fruit})

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.DevilFruit() == nil {
		t.Fatalf("expected a devil fruit to be drawn")
	}
	if loadout.FruitMastery() == enums.FruitMasteryNone {
		t.Fatalf("expected fruit mastery >= REGULAR once a fruit is drawn, got NONE")
	}
}

func TestLoadoutBuilder_NoFruitForcesNoMastery(t *testing.T) {
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.OnePiece}, weights, &fakeRandom{seq: []int{0}})
	pool := game.NewAvailablePowers(nil, nil)

	loadout, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if loadout.DevilFruit() != nil || loadout.FruitMastery() != enums.FruitMasteryNone {
		t.Fatalf("expected no fruit and NONE mastery, got fruit=%v mastery=%v", loadout.DevilFruit(), loadout.FruitMastery())
	}
}

func TestLoadoutBuilder_DoesNotRepeatStandWithinTeam(t *testing.T) {
	standA := mustStand(t, 1, "Star Platinum", enums.Legendary)
	standB := mustStand(t, 2, "The World", enums.Legendary)
	pool := game.NewAvailablePowers([]*powers.Stand{standA, standB}, nil)
	weights := game.DefaultAssignmentWeights()
	builder := game.NewLoadoutBuilder([]enums.Manga{enums.Jojo}, weights, &fakeRandom{seq: []int{1}})

	l1, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build (1st): %v", err)
	}
	l2, err := builder.Build(pool)
	if err != nil {
		t.Fatalf("Build (2nd): %v", err)
	}
	if l1.Stand() == nil || l2.Stand() == nil {
		t.Fatalf("expected both draws to produce a stand")
	}
	if l1.Stand().ID() == l2.Stand().ID() {
		t.Fatalf("expected distinct stands within the team, got the same one twice: %v", l1.Stand().ID())
	}
}
