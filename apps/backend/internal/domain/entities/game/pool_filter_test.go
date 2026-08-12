package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestPoolFilter_EmptyMeansUnrestricted(t *testing.T) {
	f, err := game.NewPoolFilter(nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}
	stand := mustStand(t, 1, "Star Platinum", enums.Legendary)
	fruit := mustDevilFruit(t, 2, "Gomu Gomu no Mi", enums.Rare, enums.Paramecia)
	if !f.AllowsStand(stand) {
		t.Fatalf("expected an unrestricted filter to allow any Stand")
	}
	if !f.AllowsDevilFruit(fruit) {
		t.Fatalf("expected an unrestricted filter to allow any DevilFruit")
	}
}

func TestPoolFilter_RarityAllowlist(t *testing.T) {
	f, err := game.NewPoolFilter([]enums.PowerRarity{enums.Legendary}, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}
	legendary := mustStand(t, 1, "Star Platinum", enums.Legendary)
	common := mustStand(t, 2, "Hermit Purple", enums.Common)
	if !f.AllowsStand(legendary) {
		t.Fatalf("expected the legendary Stand to be allowed")
	}
	if f.AllowsStand(common) {
		t.Fatalf("expected the common Stand to be excluded by the rarity allowlist")
	}
}

func TestPoolFilter_FruitTypeAllowlist(t *testing.T) {
	f, err := game.NewPoolFilter(nil, nil, []enums.FruitType{enums.Logia}, nil)
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}
	logia := mustDevilFruit(t, 1, "Mera Mera no Mi", enums.Epic, enums.Logia)
	paramecia := mustDevilFruit(t, 2, "Gomu Gomu no Mi", enums.Rare, enums.Paramecia)
	if !f.AllowsDevilFruit(logia) {
		t.Fatalf("expected the Logia fruit to be allowed")
	}
	if f.AllowsDevilFruit(paramecia) {
		t.Fatalf("expected the Paramecia fruit to be excluded by the fruit-type allowlist")
	}
}

func TestPoolFilter_BanlistOverridesAllowlist(t *testing.T) {
	stand := mustStand(t, 1, "Star Platinum", enums.Legendary)
	f, err := game.NewPoolFilter([]enums.PowerRarity{enums.Legendary}, nil, nil, []powers.PowerID{stand.ID()})
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}
	if f.AllowsStand(stand) {
		t.Fatalf("expected a banned Stand to be excluded even though its rarity is allowed")
	}
}

func TestPoolFilter_RejectsInvalidEnumMembers(t *testing.T) {
	if _, err := game.NewPoolFilter([]enums.PowerRarity{99}, nil, nil, nil); err == nil {
		t.Fatalf("expected an invalid stand rarity to be rejected")
	}
	if _, err := game.NewPoolFilter(nil, nil, []enums.FruitType{99}, nil); err == nil {
		t.Fatalf("expected an invalid fruit type to be rejected")
	}
}

func TestPoolFilter_RejectsNilBannedID(t *testing.T) {
	if _, err := game.NewPoolFilter(nil, nil, nil, []powers.PowerID{powers.NilPowerID}); err != game.ErrInvalidPoolFilter {
		t.Fatalf("expected ErrInvalidPoolFilter for a nil banned id, got %v", err)
	}
}

func TestPoolFilter_Apply(t *testing.T) {
	legendary := mustStand(t, 1, "Star Platinum", enums.Legendary)
	common := mustStand(t, 2, "Hermit Purple", enums.Common)
	logia := mustDevilFruit(t, 3, "Mera Mera no Mi", enums.Epic, enums.Logia)
	paramecia := mustDevilFruit(t, 4, "Gomu Gomu no Mi", enums.Rare, enums.Paramecia)

	f, err := game.NewPoolFilter([]enums.PowerRarity{enums.Legendary}, nil, []enums.FruitType{enums.Logia}, nil)
	if err != nil {
		t.Fatalf("NewPoolFilter: %v", err)
	}
	stands, fruits := f.Apply([]*powers.Stand{legendary, common}, []*powers.DevilFruit{logia, paramecia})
	if len(stands) != 1 || stands[0] != legendary {
		t.Fatalf("expected only the legendary Stand to survive Apply, got %v", stands)
	}
	if len(fruits) != 1 || fruits[0] != logia {
		t.Fatalf("expected only the Logia fruit to survive Apply, got %v", fruits)
	}
}
