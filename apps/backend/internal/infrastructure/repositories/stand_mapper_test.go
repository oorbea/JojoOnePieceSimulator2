package repositories

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
)

func newStandRow(name string, skills []string) standRow {
	return standRow{
		Name:          name,
		Description:   name + " description",
		Rarity:        "RARE",
		Picture:       "",
		PictureThumb:  "",
		PictureStatus: "NONE",
		AttackPower:   "C",
		Speed:         "C",
		AttackRange:   "C",
		Endurance:     "C",
		Precision:     "C",
		Potential:     "C",
		Skills:        skills,
		ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Matched:       true,
	}
}

// A corrupt row (here: no skills, which powers.NewPower rejects) must not
// poison the whole batch - buildStandsLenient is what GetAll/Filter use, so
// one bad legacy row must not 500 the entire admin catalogue.
func TestBuildStandsLenient_SkipsCorruptRow(t *testing.T) {
	good := newStandRow("Good Stand", []string{"punch"})
	bad := newStandRow("Bad Stand", nil)

	stands := buildStandsLenient([]standRow{good, bad})

	if len(stands) != 1 {
		t.Fatalf("len(stands) = %d, want 1 (only the valid row)", len(stands))
	}
	if stands[0].Name() != "Good Stand" {
		t.Errorf("stands[0].Name() = %q, want %q", stands[0].Name(), "Good Stand")
	}
}

// buildStands (strict - used by FindByID/FindByName) still fails loudly when
// the one requested Stand is itself corrupt.
func TestBuildStands_FailsOnCorruptRow(t *testing.T) {
	bad := newStandRow("Bad Stand", nil)

	if _, err := buildStands([]standRow{bad}); err == nil {
		t.Fatal("buildStands(bad row) = nil error, want an error")
	}
}
