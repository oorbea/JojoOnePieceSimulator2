package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func newOnePieceCharacterRow(name string, rarity string) onePieceCharacterRow {
	return onePieceCharacterRow{
		ID:              pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:            name,
		Description:     name + " description",
		Rarity:          rarity,
		Picture:         "",
		PictureThumb:    "",
		PictureStatus:   "NONE",
		PhysicalForm:    "MARINE_CAPTAIN",
		ArmamentHaki:    "VICE_ADMIRAL",
		ObservationHaki: "PRIVATE",
		ConquerorHaki:   "NONE",
		FruitMastery:    "ADVANCED",
	}
}

// See TestBuildJojoCharacters_SkipsCorruptRow's doc for why this matters.
func TestBuildOnePieceCharacters_SkipsCorruptRow(t *testing.T) {
	good := newOnePieceCharacterRow("Good Character", "EPIC")
	bad := newOnePieceCharacterRow("Bad Character", "NOT_A_RARITY")

	got, err := buildOnePieceCharacters([]onePieceCharacterRow{good, bad})
	if err != nil {
		t.Fatalf("buildOnePieceCharacters: unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (only the valid row)", len(got))
	}
	if got[0].Name() != "Good Character" {
		t.Errorf("got[0].Name() = %q, want %q", got[0].Name(), "Good Character")
	}
}

func TestBuildOnePieceCharacter_FailsOnCorruptRow(t *testing.T) {
	bad := newOnePieceCharacterRow("Bad Character", "NOT_A_RARITY")

	if _, err := buildOnePieceCharacter(bad); err == nil {
		t.Fatal("buildOnePieceCharacter(bad row) = nil error, want an error")
	}
}

func TestBuildOnePieceCharacter_InvalidPhysicalFormFails(t *testing.T) {
	bad := newOnePieceCharacterRow("Bad Form", "EPIC")
	bad.PhysicalForm = "NOT_A_FORM"

	if _, err := buildOnePieceCharacter(bad); err == nil {
		t.Fatal("buildOnePieceCharacter(invalid physical form) = nil error, want an error")
	}
}

func TestBuildOnePieceCharacter_InvalidHakiFails(t *testing.T) {
	bad := newOnePieceCharacterRow("Bad Haki", "EPIC")
	bad.ArmamentHaki = "NOT_A_LEVEL"

	if _, err := buildOnePieceCharacter(bad); err == nil {
		t.Fatal("buildOnePieceCharacter(invalid armament haki) = nil error, want an error")
	}
}

func TestBuildOnePieceCharacter_InvalidFruitMasteryFails(t *testing.T) {
	bad := newOnePieceCharacterRow("Bad Mastery", "EPIC")
	bad.FruitMastery = "NOT_A_MASTERY"

	if _, err := buildOnePieceCharacter(bad); err == nil {
		t.Fatal("buildOnePieceCharacter(invalid fruit mastery) = nil error, want an error")
	}
}
