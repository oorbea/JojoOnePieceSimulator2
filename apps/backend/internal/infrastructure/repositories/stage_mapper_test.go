package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func newStageRow(name string) stageRow {
	return stageRow{
		ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Manga:         "JOJO",
		Position:      3,
		Name:          name,
		Description:   name + " description",
		Picture:       "",
		PictureThumb:  "",
		PictureStatus: "NONE",
	}
}

func TestToStage(t *testing.T) {
	row := newStageRow("Stardust Crusaders")

	st, err := toStage(row)
	if err != nil {
		t.Fatalf("toStage: %v", err)
	}
	if st.Name() != "Stardust Crusaders" {
		t.Errorf("Name() = %q, want %q", st.Name(), "Stardust Crusaders")
	}
	if st.Order() != 3 {
		t.Errorf("Order() = %d, want 3", st.Order())
	}
	if st.Description() != "Stardust Crusaders description" {
		t.Errorf("Description() = %q, want %q", st.Description(), "Stardust Crusaders description")
	}
}

func TestToStage_InvalidManga(t *testing.T) {
	row := newStageRow("Bad Stage")
	row.Manga = "NOT_A_MANGA"

	if _, err := toStage(row); err == nil {
		t.Fatal("toStage(invalid manga) = nil error, want an error")
	}
}

func TestToStage_InvalidPictureStatus(t *testing.T) {
	row := newStageRow("Bad Stage")
	row.PictureStatus = "NOT_A_STATUS"

	if _, err := toStage(row); err == nil {
		t.Fatal("toStage(invalid picture status) = nil error, want an error")
	}
}

func TestToStage_EmptyDescription(t *testing.T) {
	row := newStageRow("Bad Stage")
	row.Description = ""

	if _, err := toStage(row); err == nil {
		t.Fatal("toStage(empty description) = nil error, want an error (Stage.NewStage requires it)")
	}
}
