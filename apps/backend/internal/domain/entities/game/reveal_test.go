package game

import (
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestRevealSlots_GatedByManga(t *testing.T) {
	tests := []struct {
		name   string
		mangas []enums.Manga
		want   []RevealSlot
	}{
		{
			name:   "jojo only",
			mangas: []enums.Manga{enums.Jojo},
			want:   []RevealSlot{RevealStand, RevealHamon, RevealSpin},
		},
		{
			name:   "one piece only",
			mangas: []enums.Manga{enums.OnePiece},
			want: []RevealSlot{
				RevealPhysicalForm, RevealDevilFruit, RevealFruitMastery,
				RevealHakiSet, RevealArmamentHaki, RevealObservationHaki, RevealConquerorHaki,
			},
		},
		{
			name:   "both mangas",
			mangas: []enums.Manga{enums.Jojo, enums.OnePiece},
			want: []RevealSlot{
				RevealPhysicalForm, RevealStand, RevealDevilFruit, RevealFruitMastery,
				RevealHamon, RevealHakiSet, RevealArmamentHaki, RevealObservationHaki, RevealConquerorHaki,
				RevealSpin,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RevealSlots(tt.mangas)
			if len(got) != len(tt.want) {
				t.Fatalf("RevealSlots(%v) = %v, want %v", tt.mangas, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("RevealSlots(%v)[%d] = %v, want %v", tt.mangas, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestRevealDuration_PinnedTotals pins the exact millisecond totals for each
// manga combination, computed by hand from the timing table in reveal.go's
// doc comment. Any accidental change to a constant or to which slots are
// gated must fail this test loudly, since the frontend (loadout-reveal.ts)
// independently computes the same numbers and the two must never drift.
func TestRevealDuration_PinnedTotals(t *testing.T) {
	tests := []struct {
		name   string
		mangas []enums.Manga
		wantMs int
	}{
		{"both mangas", []enums.Manga{enums.Jojo, enums.OnePiece}, 48900},
		{"jojo only", []enums.Manga{enums.Jojo}, 18350},
		{"one piece only", []enums.Manga{enums.OnePiece}, 34950},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RevealDuration(tt.mangas)
			want := time.Duration(tt.wantMs) * time.Millisecond
			if got != want {
				t.Fatalf("RevealDuration(%v) = %v, want %v", tt.mangas, got, want)
			}
		})
	}
}
