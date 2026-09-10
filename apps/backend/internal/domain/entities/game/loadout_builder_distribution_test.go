package game_test

import (
	"math/rand"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// mathRandSource adapts math/rand to game.RandomSource with a fixed seed,
// so this test is deterministic despite drawing a large sample.
type mathRandSource struct{ r *rand.Rand }

func (m mathRandSource) IntN(n int) int { return m.r.Intn(n) }

// TestLoadoutBuilder_MatchesV1Distributions draws a large sample and checks
// the observed frequencies against JoJoOnePiece_Simulator V1's published
// probabilities (github.com/oorbea/JoJoOnePiece_Simulator, powers.cc),
// within a tolerance - the only way to actually demonstrate the port is
// faithful, and a permanent regression guard against an accidental weight
// edit. Stand/DevilFruit draws are checked for uniformity across a 4-item
// pool instead, since V1's uniform_int_distribution has no "probability
// table" to compare against directly.
func TestLoadoutBuilder_MatchesV1Distributions(t *testing.T) {
	const n = 100_000
	const tolerance = 0.02 // absolute probability points

	stands := []*powers.Stand{
		mustStand(t, 1, "Star Platinum", enums.Common),
		mustStand(t, 2, "The World", enums.Rare),
		mustStand(t, 3, "Crazy Diamond", enums.Epic),
		mustStand(t, 4, "Gold Experience", enums.Legendary),
	}
	weights := game.DefaultAssignmentWeights()
	rng := mathRandSource{rand.New(rand.NewSource(42))}
	builder := game.NewLoadoutBuilder(enums.Mangas(), weights, rng)

	standCounts := map[string]int{"none": 0}
	for _, s := range stands {
		standCounts[s.Name()] = 0
	}
	spinCounts := map[enums.SpinLevel]int{}
	hamonCounts := map[enums.HamonLevel]int{}
	physicalFormCounts := map[enums.PhysicalForm]int{}
	hakiPresence := map[string]int{"armament": 0, "observation": 0, "conqueror": 0}
	battleIQBandCounts := map[game.BattleIQBand]int{}
	verySuperiorBelow160 := 0
	verySuperiorTotal := 0

	for i := 0; i < n; i++ {
		pool := game.NewAvailablePowers(append([]*powers.Stand(nil), stands...), nil)
		l, err := builder.Build(pool)
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		if l.Stand() == nil {
			standCounts["none"]++
		} else {
			standCounts[l.Stand().Name()]++
		}
		spinCounts[l.Spin()]++
		hamonCounts[l.Hamon()]++
		physicalFormCounts[l.PhysicalForm()]++
		if l.ArmamentHaki() != enums.HakiNone {
			hakiPresence["armament"]++
		}
		if l.ObservationHaki() != enums.HakiNone {
			hakiPresence["observation"]++
		}
		if l.ConquerorHaki() != enums.HakiNone {
			hakiPresence["conqueror"]++
		}
		if l.BattleIQ().Present() {
			band := l.BattleIQ().Band()
			battleIQBandCounts[band]++
			if band == game.BattleIQVerySuperior {
				verySuperiorTotal++
				if l.BattleIQ().Value() < 160 {
					verySuperiorBelow160++
				}
			}
		}
	}

	// Stand: uniform over "none" + 4 stands => each ~20%.
	for name, count := range standCounts {
		got := float64(count) / n
		if diff := got - 0.20; diff < -tolerance || diff > tolerance {
			t.Errorf("stand %q: got p=%.4f, want ~0.20", name, got)
		}
	}

	// Spin: V1's 15/30/30/25.
	assertProb(t, "spin NONE", spinCounts[enums.SpinNone], n, 0.15, tolerance)
	assertProb(t, "spin BASIC", spinCounts[enums.SpinBasic], n, 0.30, tolerance)
	assertProb(t, "spin GOLDEN", spinCounts[enums.SpinGolden], n, 0.30, tolerance)
	assertProb(t, "spin INFINITE", spinCounts[enums.SpinInfinite], n, 0.25, tolerance)

	// Hamon: owner-chosen 25/35/35/5.
	assertProb(t, "hamon NONE", hamonCounts[enums.HamonNone], n, 0.25, tolerance)
	assertProb(t, "hamon BASIC", hamonCounts[enums.HamonBasic], n, 0.35, tolerance)
	assertProb(t, "hamon ADVANCED", hamonCounts[enums.HamonAdvanced], n, 0.35, tolerance)
	assertProb(t, "hamon PERFECT", hamonCounts[enums.HamonPerfect], n, 0.05, tolerance)

	// PhysicalForm: V1's uniform 1/6.
	for level, count := range physicalFormCounts {
		got := float64(count) / n
		if diff := got - 1.0/6.0; diff < -tolerance || diff > tolerance {
			t.Errorf("physical form %v: got p=%.4f, want ~0.1667", level, got)
		}
	}

	// Haki presence marginals derived from V1's 8-case set table:
	// armament = 20+20+15+10 = 65%, observation = 20+20+15+10 = 65%,
	// conqueror = 1+15+10+10 = 36%.
	assertProb(t, "armament present", hakiPresence["armament"], n, 0.65, tolerance)
	assertProb(t, "observation present", hakiPresence["observation"], n, 0.65, tolerance)
	assertProb(t, "conqueror present", hakiPresence["conqueror"], n, 0.36, tolerance)

	// BattleIQ band: owner-chosen WAIS-IV normative frequencies
	// 2/7/16/50/16/7/2 (%). Every loadout here is JoJo, so BattleIQ is
	// present for all n draws.
	assertProb(t, "battleIQ EXTREMELY_LOW", battleIQBandCounts[game.BattleIQExtremelyLow], n, 0.02, tolerance)
	assertProb(t, "battleIQ BORDERLINE", battleIQBandCounts[game.BattleIQBorderline], n, 0.07, tolerance)
	assertProb(t, "battleIQ LOW_AVERAGE", battleIQBandCounts[game.BattleIQLowAverage], n, 0.16, tolerance)
	assertProb(t, "battleIQ AVERAGE", battleIQBandCounts[game.BattleIQAverage], n, 0.50, tolerance)
	assertProb(t, "battleIQ HIGH_AVERAGE", battleIQBandCounts[game.BattleIQHighAverage], n, 0.16, tolerance)
	assertProb(t, "battleIQ SUPERIOR", battleIQBandCounts[game.BattleIQSuperior], n, 0.07, tolerance)
	assertProb(t, "battleIQ VERY_SUPERIOR", battleIQBandCounts[game.BattleIQVerySuperior], n, 0.02, tolerance)

	// Within VERY_SUPERIOR (130-255), the 30-point half-life decay means
	// roughly half the draws land in its bottom half (130-159) rather than
	// spreading uniformly across the full 126-value range.
	if verySuperiorTotal > 0 {
		got := float64(verySuperiorBelow160) / float64(verySuperiorTotal)
		if diff := got - 0.52; diff < -0.08 || diff > 0.08 {
			t.Errorf("battleIQ VERY_SUPERIOR values below 160: got p=%.4f, want ~0.52", got)
		}
	}
}

func assertProb(t *testing.T, label string, count, n int, want, tolerance float64) {
	t.Helper()
	got := float64(count) / float64(n)
	if diff := got - want; diff < -tolerance || diff > tolerance {
		t.Errorf("%s: got p=%.4f, want ~%.4f (tolerance %.4f)", label, got, want, tolerance)
	}
}
