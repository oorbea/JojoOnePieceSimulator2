package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestNewLoadout_NoFruitRequiresNoneMastery(t *testing.T) {
	_, err := game.NewLoadout(nil, nil, enums.SpinNone, enums.HamonNone, enums.FruitMasteryRegular,
		enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate)
	if err != game.ErrFruitMasteryMismatch {
		t.Fatalf("expected ErrFruitMasteryMismatch, got %v", err)
	}
}

func TestNewLoadout_FruitRequiresAtLeastRegularMastery(t *testing.T) {
	fruit := mustDevilFruit(t, 1, "Gomu Gomu no Mi", enums.Legendary, enums.Paramecia)

	if _, err := game.NewLoadout(nil, fruit, enums.SpinNone, enums.HamonNone, enums.FruitMasteryNone,
		enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate); err != game.ErrFruitMasteryMismatch {
		t.Fatalf("expected ErrFruitMasteryMismatch, got %v", err)
	}

	if _, err := game.NewLoadout(nil, fruit, enums.SpinNone, enums.HamonNone, enums.FruitMasteryRegular,
		enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate); err != nil {
		t.Fatalf("expected a valid loadout, got %v", err)
	}
}

func TestNewLoadout_RequiresSpin4Trait(t *testing.T) {
	tusk := mustStand(t, 1, "Tusk ACT4", enums.Legendary)

	if _, err := game.NewLoadout(tusk, nil, enums.SpinBasic, enums.HamonNone, enums.FruitMasteryNone,
		enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate); err != game.ErrSpin4Required {
		t.Fatalf("expected ErrSpin4Required, got %v", err)
	}

	l, err := game.NewLoadout(tusk, nil, enums.SpinInfinite, enums.HamonNone, enums.FruitMasteryNone,
		enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate)
	if err != nil {
		t.Fatalf("expected a valid loadout, got %v", err)
	}
	if l.Spin() != enums.SpinInfinite {
		t.Fatalf("expected spin INFINITE, got %v", l.Spin())
	}
}

func TestNewLoadout_SpinAndHamonIndependentOfStand(t *testing.T) {
	// A player with no stand at all can still have spin and hamon (e.g. a
	// pure Hamon/Ripple user like Zeppeli).
	l, err := game.NewLoadout(nil, nil, enums.SpinGolden, enums.HamonPerfect, enums.FruitMasteryNone,
		enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate)
	if err != nil {
		t.Fatalf("expected a valid loadout, got %v", err)
	}
	if l.Spin() != enums.SpinGolden || l.Hamon() != enums.HamonPerfect {
		t.Fatalf("expected spin/hamon preserved without a stand, got %v/%v", l.Spin(), l.Hamon())
	}

	// A regular (non-trait) stand doesn't force any particular spin level.
	star := mustStand(t, 2, "Star Platinum", enums.Legendary)
	l2, err := game.NewLoadout(star, nil, enums.SpinNone, enums.HamonNone, enums.FruitMasteryNone,
		enums.HakiPrivate, enums.HakiPrivate, enums.HakiPrivate, enums.PhysicalFormPrivate)
	if err != nil {
		t.Fatalf("expected a valid loadout, got %v", err)
	}
	if l2.Spin() != enums.SpinNone {
		t.Fatalf("expected spin unaffected by a non-trait stand, got %v", l2.Spin())
	}
}
