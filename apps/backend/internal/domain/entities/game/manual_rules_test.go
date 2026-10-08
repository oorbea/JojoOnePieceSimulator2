package game_test

import (
	"math"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

func chainByStage(t *testing.T, rules game.ManualRules, stage string) game.ManualChain {
	t.Helper()
	for _, c := range rules.Evolutions {
		for _, s := range c.Stages {
			if s.Name == stage {
				return c
			}
		}
	}
	t.Fatalf("no evolution chain holds %q", stage)
	return game.ManualChain{}
}

func outcome(t *testing.T, c game.ManualChain, drawn, level string) game.ManualOutcome {
	t.Helper()
	for _, r := range c.Rows {
		if r.Drawn != drawn {
			continue
		}
		for _, o := range r.Outcomes {
			if o.Level == level {
				return o
			}
		}
	}
	t.Fatalf("no outcome for %q at %q", drawn, level)
	return game.ManualOutcome{}
}

// The matrix is computed by the real resolver; these are the outcomes the
// owner agreed to (2026-10-09), pinned so a rule change shows up here and in
// the regenerated contracts/rules.ts at the same time.
func TestManualRules_TuskMatrix(t *testing.T) {
	c := chainByStage(t, game.BuildManualRules(), "Tusk: Act 1")
	tests := []struct {
		drawn, spin, wantPower, wantSpin, wantChange string
	}{
		{"Tusk: Act 1", "NONE", "Tusk: Act 1", "BASIC", "LEVEL_RAISED"},
		{"Tusk: Act 1", "BASIC", "Tusk: Act 1", "BASIC", "NONE"},
		{"Tusk: Act 1", "GOLDEN", "Tusk: Act 3", "GOLDEN", "EVOLVED"},
		{"Tusk: Act 1", "INFINITE", "Tusk: Act 4", "INFINITE", "EVOLVED"},
		{"Tusk: Act 2", "NONE", "Tusk: Act 2", "GOLDEN", "LEVEL_RAISED"},
		{"Tusk: Act 2", "BASIC", "Tusk: Act 2", "GOLDEN", "LEVEL_RAISED"},
		{"Tusk: Act 2", "GOLDEN", "Tusk: Act 2", "GOLDEN", "NONE"},
		{"Tusk: Act 2", "INFINITE", "Tusk: Act 4", "INFINITE", "EVOLVED"},
		{"Tusk: Act 3", "GOLDEN", "Tusk: Act 3", "GOLDEN", "NONE"},
		{"Tusk: Act 3", "INFINITE", "Tusk: Act 4", "INFINITE", "EVOLVED"},
		{"Tusk: Act 4", "NONE", "Tusk: Act 4", "INFINITE", "LEVEL_RAISED"},
		{"Tusk: Act 4", "GOLDEN", "Tusk: Act 4", "INFINITE", "LEVEL_RAISED"},
		{"Tusk: Act 4", "INFINITE", "Tusk: Act 4", "INFINITE", "NONE"},
	}
	for _, tc := range tests {
		got := outcome(t, c, tc.drawn, tc.spin)
		if got.Power != tc.wantPower || got.LevelAfter != tc.wantSpin || got.Change != tc.wantChange {
			t.Errorf("%s + %s = %+v, want %s / %s / %s", tc.drawn, tc.spin, got, tc.wantPower, tc.wantSpin, tc.wantChange)
		}
	}
}

func TestManualRules_SoftAndWetAndNika(t *testing.T) {
	rules := game.BuildManualRules()
	sw := chainByStage(t, rules, "Soft & Wet")
	if got := outcome(t, sw, "Soft & Wet", "INFINITE"); got.Power != "Soft & Wet: Go Beyond" || got.Change != "EVOLVED" {
		t.Errorf("Soft & Wet + INFINITE = %+v, want Go Beyond", got)
	}
	if got := outcome(t, sw, "Soft & Wet", "GOLDEN"); got.Change != "NONE" {
		t.Errorf("Soft & Wet + GOLDEN = %+v, want no change", got)
	}
	if got := outcome(t, sw, "Soft & Wet: Go Beyond", "NONE"); got.LevelAfter != "INFINITE" || got.Change != "LEVEL_RAISED" {
		t.Errorf("Go Beyond + NONE = %+v, want spin raised to INFINITE", got)
	}

	gomu := chainByStage(t, rules, "Gomu Gomu no mi")
	if got := outcome(t, gomu, "Gomu Gomu no mi", "AWAKENED"); got.Power != "Hito Hito no mi: Model Nika" || got.Change != "EVOLVED" {
		t.Errorf("Gomu + AWAKENED = %+v, want Nika", got)
	}
	if got := outcome(t, gomu, "Gomu Gomu no mi", "ADVANCED"); got.Change != "NONE" {
		t.Errorf("Gomu + ADVANCED = %+v, want no change", got)
	}
	if got := outcome(t, gomu, "Hito Hito no mi: Model Nika", "REGULAR"); got.LevelAfter != "AWAKENED" || got.Change != "LEVEL_RAISED" {
		t.Errorf("Nika + REGULAR = %+v, want mastery raised to AWAKENED", got)
	}
}

// Every stage the rule tables give a tier must appear in a manual chain,
// otherwise the manual would silently omit a Stand or fruit that evolves.
func TestManualRules_CoverEveryTier(t *testing.T) {
	rules := game.BuildManualRules()
	tiered := map[string]bool{}
	for _, c := range rules.Evolutions {
		for _, s := range c.Stages {
			if s.Tier != "" {
				tiered[s.Name] = true
			}
		}
	}
	for _, want := range []string{"Tusk: Act 1", "Tusk: Act 2", "Tusk: Act 3", "Tusk: Act 4",
		"Ball Breaker", "Soft & Wet: Go Beyond", "Hito Hito no mi: Model Nika"} {
		if !tiered[want] {
			t.Errorf("tiered stage %q is missing from the manual's evolution chains", want)
		}
	}
}

func TestManualRules_StatFloorsMirrorResolverTable(t *testing.T) {
	floors := game.BuildManualRules().StatFloors
	if len(floors) != 8 {
		t.Fatalf("expected the 8 agreed floor rules, got %d", len(floors))
	}
	kc := floors[1]
	if kc.When.Names[0] != "King Crimson" || kc.Target != "OBSERVATION_HAKI" || kc.Floor != "YONKO_PLUS" || !kc.Cross {
		t.Errorf("King Crimson rule = %+v", kc)
	}
}

func TestManualRules_OddsAreDistributions(t *testing.T) {
	odds := game.BuildManualRules().Odds
	for name, ps := range map[string][]game.ManualPct{
		"spin": odds.Spin, "hamon": odds.Hamon, "fruitMastery": odds.FruitMastery,
		"physicalForm": odds.PhysicalForm, "hakiMastery": odds.HakiMastery,
	} {
		sum := 0.0
		for _, p := range ps {
			sum += p.Percent
		}
		if math.Abs(sum-100) > 0.1 {
			t.Errorf("%s sums to %.2f%%", name, sum)
		}
	}
	if odds.HakiPresence.Armament != 65 || odds.HakiPresence.Observation != 65 || odds.HakiPresence.Conqueror != 36 {
		t.Errorf("haki presence = %+v, want 65/65/36", odds.HakiPresence)
	}
	iq := game.BuildManualRules().BattleIQ
	if len(iq) != 7 || iq[0].Lo != 0 || iq[6].Hi != 255 || iq[3].Percent != 50 {
		t.Errorf("battle IQ bands = %+v", iq)
	}
}
