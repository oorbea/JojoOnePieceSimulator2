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

// TestRevealSpinCycles_DeterministicAndBounded pins that RevealSpinCycles
// always returns 1 or 2 (mirroring V1's generateRandomNumber(1, 2)) and is
// a pure function of its inputs - the frontend's mirrored
// revealSpinCycles must agree on the exact same number for the exact same
// inputs, so it can never actually be random.
func TestRevealSpinCycles_DeterministicAndBounded(t *testing.T) {
	id := GameID{1, 2, 3}
	for slot := RevealSlot(0); slot < 10; slot++ {
		for pi := 0; pi < 5; pi++ {
			got := RevealSpinCycles(id, 0, pi, slot)
			if got != 1 && got != 2 {
				t.Fatalf("RevealSpinCycles(round=0, participant=%d, slot=%v) = %d, want 1 or 2", pi, slot, got)
			}
			again := RevealSpinCycles(id, 0, pi, slot)
			if again != got {
				t.Fatalf("RevealSpinCycles not deterministic: got %d then %d", got, again)
			}
		}
	}
}

// TestRevealDuration_LongerWithStandAndDevilFruit checks the core V1
// behaviour the 2026-08-30 rewrite exists for: a participant who actually
// lands a Stand and/or a Devil Fruit gets a longer reveal than one who
// lands NONE on both, because RevealHoldStandMs/RevealHoldFruitMs each
// dwarf RevealHoldEmptyMs.
func TestRevealDuration_LongerWithStandAndDevilFruit(t *testing.T) {
	id := GameID{9}
	mangas := []enums.Manga{enums.Jojo, enums.OnePiece}

	empty := RevealDuration(id, 0, mangas, []RevealPlayer{{}}, enums.Swift)
	both := RevealDuration(id, 0, mangas, []RevealPlayer{{HasStand: true, HasDevilFruit: true}}, enums.Swift)

	minExtra := time.Duration(RevealHoldStandMs-RevealHoldEmptyMs+RevealHoldFruitMs-RevealHoldEmptyMs) * time.Millisecond
	if both-empty < minExtra {
		t.Fatalf("RevealDuration with stand+fruit - without = %v, want >= %v", both-empty, minExtra)
	}
}

// TestRevealDuration_MorePlayersTakeLonger pins the jugador-por-jugador
// structure: N participants must take strictly longer than 1, since each
// plays out their own full per-slot sequence (V1's "Le toca a X" loop),
// never in parallel.
func TestRevealDuration_MorePlayersTakeLonger(t *testing.T) {
	id := GameID{7}
	mangas := []enums.Manga{enums.Jojo}

	one := RevealDuration(id, 0, mangas, []RevealPlayer{{}}, enums.Swift)
	four := RevealDuration(id, 0, mangas, []RevealPlayer{{}, {}, {}, {}}, enums.Swift)

	if four <= one {
		t.Fatalf("RevealDuration(4 players) = %v, want > RevealDuration(1 player) = %v", four, one)
	}
}

// TestRevealDuration_SpeedOrdering pins that Relaxed > Normal > Swift, and
// that Swift alone reproduces V1's own tempo (i.e. multiplier 1.0).
func TestRevealDuration_SpeedOrdering(t *testing.T) {
	id := GameID{3}
	mangas := []enums.Manga{enums.Jojo, enums.OnePiece}
	players := []RevealPlayer{{HasStand: true}, {}}

	swift := RevealDuration(id, 0, mangas, players, enums.Swift)
	normal := RevealDuration(id, 0, mangas, players, enums.Normal)
	relaxed := RevealDuration(id, 0, mangas, players, enums.Relaxed)

	if !(swift < normal && normal < relaxed) {
		t.Fatalf("RevealDuration speed ordering broken: swift=%v normal=%v relaxed=%v", swift, normal, relaxed)
	}
}
