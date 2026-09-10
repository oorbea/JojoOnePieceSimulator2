package characters_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestNewOnePieceCharacter_ValidatesEnums(t *testing.T) {
	base := mustBaseCharacter(t)

	cases := []struct {
		name            string
		physicalForm    enums.PhysicalForm
		armamentHaki    enums.HakiLevel
		observationHaki enums.HakiLevel
		conquerorHaki   enums.HakiLevel
		fruitMastery    enums.FruitMastery
	}{
		{"invalid physical form", enums.PhysicalForm(99), enums.HakiNone, enums.HakiNone, enums.HakiNone, enums.FruitMasteryNone},
		{"invalid armament haki", enums.PhysicalFormPrivate, enums.HakiLevel(99), enums.HakiNone, enums.HakiNone, enums.FruitMasteryNone},
		{"invalid observation haki", enums.PhysicalFormPrivate, enums.HakiNone, enums.HakiLevel(99), enums.HakiNone, enums.FruitMasteryNone},
		{"invalid conqueror haki", enums.PhysicalFormPrivate, enums.HakiNone, enums.HakiNone, enums.HakiLevel(99), enums.FruitMasteryNone},
		{"invalid fruit mastery", enums.PhysicalFormPrivate, enums.HakiNone, enums.HakiNone, enums.HakiNone, enums.FruitMastery(99)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := characters.NewOnePieceCharacter(base, c.physicalForm, c.armamentHaki, c.observationHaki, c.conquerorHaki, c.fruitMastery); err == nil {
				t.Errorf("%s: err = nil, want an error", c.name)
			}
		})
	}
}

func TestOnePieceCharacter_Accessors(t *testing.T) {
	base := mustBaseCharacter(t)
	c, err := characters.NewOnePieceCharacter(base, enums.PhysicalFormMarineCaptain, enums.HakiViceAdmiral, enums.HakiPrivate, enums.HakiNone, enums.FruitMasteryAdvanced)
	if err != nil {
		t.Fatalf("NewOnePieceCharacter: %v", err)
	}
	if c.PhysicalForm() != enums.PhysicalFormMarineCaptain {
		t.Errorf("PhysicalForm() = %v, want MARINE_CAPTAIN", c.PhysicalForm())
	}
	if c.ArmamentHaki() != enums.HakiViceAdmiral {
		t.Errorf("ArmamentHaki() = %v, want VICE_ADMIRAL", c.ArmamentHaki())
	}
	if c.ObservationHaki() != enums.HakiPrivate {
		t.Errorf("ObservationHaki() = %v, want PRIVATE", c.ObservationHaki())
	}
	if c.ConquerorHaki() != enums.HakiNone {
		t.Errorf("ConquerorHaki() = %v, want NONE", c.ConquerorHaki())
	}
	if c.FruitMastery() != enums.FruitMasteryAdvanced {
		t.Errorf("FruitMastery() = %v, want ADVANCED", c.FruitMastery())
	}
}
