package repositories

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
)

func newDevilFruitRow(name string, skills []string) devilFruitRow {
	return devilFruitRow{
		ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:          name,
		Description:   name + " description",
		Rarity:        "RARE",
		Picture:       "",
		PictureThumb:  "",
		PictureStatus: "NONE",
		FruitType:     "PARAMECIA",
		Skills:        skills,
	}
}

// A corrupt row (here: no skills, which powers.NewPower rejects) must not
// poison the whole batch - buildDevilFruits is what GetAll/Filter use, so
// one bad legacy row must not 500 the entire admin catalogue.
func TestBuildDevilFruits_SkipsCorruptRow(t *testing.T) {
	good := newDevilFruitRow("Good Fruit", []string{"punch"})
	bad := newDevilFruitRow("Bad Fruit", nil)

	fruits, err := buildDevilFruits([]devilFruitRow{good, bad})
	if err != nil {
		t.Fatalf("buildDevilFruits: unexpected error: %v", err)
	}
	if len(fruits) != 1 {
		t.Fatalf("len(fruits) = %d, want 1 (only the valid row)", len(fruits))
	}
	if fruits[0].Name() != "Good Fruit" {
		t.Errorf("fruits[0].Name() = %q, want %q", fruits[0].Name(), "Good Fruit")
	}
}

// buildDevilFruit (singular - used by FindByID/FindByName) still fails
// loudly when the one requested DevilFruit is itself corrupt.
func TestBuildDevilFruit_FailsOnCorruptRow(t *testing.T) {
	bad := newDevilFruitRow("Bad Fruit", nil)

	if _, err := buildDevilFruit(bad); err == nil {
		t.Fatal("buildDevilFruit(bad row) = nil error, want an error")
	}
}
