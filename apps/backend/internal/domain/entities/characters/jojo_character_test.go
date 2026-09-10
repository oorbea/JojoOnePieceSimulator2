package characters_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func mustBaseCharacter(t *testing.T) characters.Character {
	t.Helper()
	c, err := characters.NewCharacter(mustCharacterID(t, 1), enums.Jojo, "Jotaro", enums.Rare, "desc", "")
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}
	return c
}

func TestNewJojoCharacter_ValidatesEnums(t *testing.T) {
	base := mustBaseCharacter(t)

	if _, err := characters.NewJojoCharacter(base, enums.HamonLevel(99), enums.SpinBasic, 100); err == nil {
		t.Error("invalid hamon: err = nil, want an error")
	}
	if _, err := characters.NewJojoCharacter(base, enums.HamonBasic, enums.SpinLevel(99), 100); err == nil {
		t.Error("invalid spin: err = nil, want an error")
	}
}

// TestNewJojoCharacter_BattleIQZeroIsValid proves 0 is a legitimate (if
// absurd) admin-authored score - unlike game.BattleIQ, there is no
// present/absent distinction for a hand-authored character.
func TestNewJojoCharacter_BattleIQZeroIsValid(t *testing.T) {
	base := mustBaseCharacter(t)

	c, err := characters.NewJojoCharacter(base, enums.HamonBasic, enums.SpinBasic, 0)
	if err != nil {
		t.Fatalf("NewJojoCharacter(battleIQ=0): %v", err)
	}
	if c.BattleIQ() != 0 {
		t.Errorf("BattleIQ() = %d, want 0", c.BattleIQ())
	}
}

func TestJojoCharacter_Accessors(t *testing.T) {
	base := mustBaseCharacter(t)
	c, err := characters.NewJojoCharacter(base, enums.HamonAdvanced, enums.SpinGolden, 130)
	if err != nil {
		t.Fatalf("NewJojoCharacter: %v", err)
	}
	if c.Hamon() != enums.HamonAdvanced {
		t.Errorf("Hamon() = %v, want ADVANCED", c.Hamon())
	}
	if c.Spin() != enums.SpinGolden {
		t.Errorf("Spin() = %v, want GOLDEN", c.Spin())
	}
	if c.BattleIQ() != 130 {
		t.Errorf("BattleIQ() = %d, want 130", c.BattleIQ())
	}
	if c.Name() != "Jotaro" {
		t.Errorf("embedded Character.Name() = %q, want Jotaro", c.Name())
	}
}
