package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func newJojoCharacterRow(name string, rarity string) jojoCharacterRow {
	return jojoCharacterRow{
		ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:          name,
		Description:   name + " description",
		Rarity:        rarity,
		Picture:       "",
		PictureThumb:  "",
		PictureStatus: "NONE",
		Hamon:         "ADVANCED",
		Spin:          "GOLDEN",
		BattleIQ:      130,
	}
}

// A corrupt row (here: an unrecognized rarity) must not poison the whole
// batch - buildJojoCharacters is what GetAll/Filter use, so one bad row
// must not 500 the entire admin catalogue. Same contract as
// TestBuildDevilFruits_SkipsCorruptRow.
func TestBuildJojoCharacters_SkipsCorruptRow(t *testing.T) {
	good := newJojoCharacterRow("Good Character", "RARE")
	bad := newJojoCharacterRow("Bad Character", "NOT_A_RARITY")

	got, err := buildJojoCharacters([]jojoCharacterRow{good, bad})
	if err != nil {
		t.Fatalf("buildJojoCharacters: unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (only the valid row)", len(got))
	}
	if got[0].Name() != "Good Character" {
		t.Errorf("got[0].Name() = %q, want %q", got[0].Name(), "Good Character")
	}
}

// buildJojoCharacter (singular - used by FindByID/FindByName) still fails
// loudly when the one requested character is itself corrupt.
func TestBuildJojoCharacter_FailsOnCorruptRow(t *testing.T) {
	bad := newJojoCharacterRow("Bad Character", "NOT_A_RARITY")

	if _, err := buildJojoCharacter(bad); err == nil {
		t.Fatal("buildJojoCharacter(bad row) = nil error, want an error")
	}
}

func TestBuildJojoCharacter_InvalidHamonFails(t *testing.T) {
	bad := newJojoCharacterRow("Bad Hamon", "RARE")
	bad.Hamon = "NOT_A_LEVEL"

	if _, err := buildJojoCharacter(bad); err == nil {
		t.Fatal("buildJojoCharacter(invalid hamon) = nil error, want an error")
	}
}

func TestBuildJojoCharacter_InvalidSpinFails(t *testing.T) {
	bad := newJojoCharacterRow("Bad Spin", "RARE")
	bad.Spin = "NOT_A_LEVEL"

	if _, err := buildJojoCharacter(bad); err == nil {
		t.Fatal("buildJojoCharacter(invalid spin) = nil error, want an error")
	}
}
