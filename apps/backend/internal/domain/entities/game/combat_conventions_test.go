package game_test

import (
	"strings"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func conventionByID(t *testing.T, id string) game.Convention {
	t.Helper()
	for _, c := range game.CombatConventions() {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("no convention %q", id)
	return game.Convention{}
}

func subjectsHaveName(subjects []game.Subject, name string) bool {
	for _, s := range subjects {
		for _, n := range s.Names {
			if strings.EqualFold(n, name) {
				return true
			}
		}
	}
	return false
}

func subjectsHaveSpinInfinite(subjects []game.Subject) bool {
	for _, s := range subjects {
		if s.Kind == game.SubjectSlotMin && s.Slot == enums.SlotSpin && s.Min == enums.SpinInfinite.String() {
			return true
		}
	}
	return false
}

func TestCombatConventions_WellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range game.CombatConventions() {
		if c.ID == "" || c.Category == "" {
			t.Errorf("convention %+v lacks an ID or category", c)
		}
		if seen[c.ID] {
			t.Errorf("duplicate convention ID %q", c.ID)
		}
		seen[c.ID] = true
		if len(c.Actors) == 0 {
			t.Errorf("%s has no actors", c.ID)
		}
		for _, group := range [][]game.Subject{c.Actors, c.Targets, c.Immune, c.Attenuated} {
			for _, s := range group {
				switch s.Kind {
				case game.SubjectSlotMin:
					if s.Min == "" {
						t.Errorf("%s: slot subject %v without a minimum level", c.ID, s.Slot)
					}
				case game.SubjectPowers:
					if len(s.Names) == 0 {
						t.Errorf("%s: power subject without names", c.ID)
					}
				case game.SubjectFruitType:
					if len(s.FruitTypes) == 0 {
						t.Errorf("%s: fruit-type subject without types", c.ID)
					}
				case game.SubjectAnyStand, game.SubjectAnyFruit:
				default:
					t.Errorf("%s: unknown subject kind %q", c.ID, s.Kind)
				}
			}
		}
	}
}

// Spin INFINITE beats every defence except the Stands it lists as immune. An
// absolute Stand therefore must be listed there exactly when its own list of
// who can overcome it does NOT include Spin INFINITE - otherwise the manual
// would tell voters both that Spin INFINITE beats it and that it does not.
func TestCombatConventions_SpinInfiniteAgreesWithAbsolutes(t *testing.T) {
	supreme := conventionByID(t, "SPIN_INFINITE_SUPREME")
	for _, c := range game.CombatConventions() {
		if !strings.HasPrefix(c.ID, "ABSOLUTE_") {
			continue
		}
		stand := c.Targets[0].Names[0]
		beatenBySpin := subjectsHaveSpinInfinite(c.Actors)
		immune := subjectsHaveName(supreme.Immune, stand)
		// Tusk: Act 4 is itself a Spin INFINITE user, so it is covered by the
		// "another Spin INFINITE" immunity (a duel for the voters) rather than
		// by name.
		if stand == "Tusk: Act 4" {
			immune = subjectsHaveSpinInfinite(supreme.Immune)
		}
		if beatenBySpin == immune {
			t.Errorf("%s: Spin INFINITE beats it = %v, but SPIN_INFINITE_SUPREME lists it as immune = %v",
				c.ID, beatenBySpin, immune)
		}
	}
}

// Yami Yami never nullifies an absolute Stand, a time stopper, Made in Heaven
// or Go Beyond.
func TestCombatConventions_YamiExceptions(t *testing.T) {
	yami := conventionByID(t, "YAMI_NULLIFIES")
	for _, c := range game.CombatConventions() {
		if strings.HasPrefix(c.ID, "ABSOLUTE_") {
			if !subjectsHaveName(yami.Immune, c.Targets[0].Names[0]) {
				t.Errorf("%s is absolute but Yami Yami is not told to leave it alone", c.ID)
			}
		}
	}
	for _, name := range []string{"The World", "Star Platinum: The World", "The World (Steel Ball Run)",
		"King Crimson", "Made in Heaven", "Soft & Wet: Go Beyond"} {
		if !subjectsHaveName(yami.Immune, name) {
			t.Errorf("Yami Yami should not nullify %s", name)
		}
	}
}

func TestCombatConventions_TimeStoppersAreOneGroup(t *testing.T) {
	stop := conventionByID(t, "TIME_STOP")
	move := conventionByID(t, "MOVE_IN_TIME_STOP")
	if len(stop.Actors) != 1 || len(move.Targets) != 1 ||
		strings.Join(stop.Actors[0].Names, "|") != strings.Join(move.Targets[0].Names, "|") {
		t.Fatalf("TIME_STOP actors and MOVE_IN_TIME_STOP targets must be the same group")
	}
	// A time stopper can move inside a stop, and so can the extra counters.
	if !subjectsHaveName(move.Actors, "The World") {
		t.Errorf("time stoppers should move in a time stop")
	}
}

func TestCombatConventions_ConquerorGaps(t *testing.T) {
	if got := conventionByID(t, "CONQUEROR_KNOCKOUT").Gap; got != 2 {
		t.Errorf("knockout gap = %d, want 2", got)
	}
	if got := conventionByID(t, "CONQUEROR_HINDER").Gap; got != 1 {
		t.Errorf("hinder gap = %d, want 1", got)
	}
}

func TestConventionPowerNames_NormalizedAndIncluded(t *testing.T) {
	got := game.ConventionPowerNames()
	if len(got) == 0 {
		t.Fatal("expected convention power names")
	}
	for _, n := range got {
		if n != strings.Join(strings.Fields(strings.ToLower(n)), " ") {
			t.Errorf("%q is not normalized", n)
		}
	}
	rules := map[string]bool{}
	for _, n := range game.PowerEffectRuleNames() {
		rules[n] = true
	}
	for _, n := range got {
		if !rules[n] {
			t.Errorf("%q is missing from PowerEffectRuleNames, so the catalogue test would not cover it", n)
		}
	}
}
