package migrations_test

import (
	"strings"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/db/migrations"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

// prodOnlyRuleNames are the rule-table names that exist only in prod's
// catalogue (admin-created), not in any seed migration. They can't be checked
// against the seeds, so they are pinned here by hand: if an admin renames one
// in prod, the rule stops firing silently (power effects match by exact name).
var prodOnlyRuleNames = map[string]bool{
	"tusk: acto 1":          true,
	"tusk: acto 2":          true,
	"tusk: acto 3":          true,
	"tusk: acto 4":          true,
	"ball breaker":          true,
	"soft & wet: go beyond": true,
}

// TestPowerEffectRuleNamesExistInCatalogue guards game's name-keyed rule table
// (power_effects.go) against drifting from the catalogue: every rule name must
// appear as a quoted value in a seed migration, or be in the prod-only
// allowlist above. It also fails when an allowlisted name becomes seeded, so
// the allowlist can't go stale.
func TestPowerEffectRuleNamesExistInCatalogue(t *testing.T) {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var sql strings.Builder
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		b, err := migrations.FS.ReadFile(e.Name())
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", e.Name(), err)
		}
		sql.WriteString(strings.ToLower(string(b)))
	}
	seeds := sql.String()

	rules := map[string]bool{}
	for _, name := range game.PowerEffectRuleNames() {
		rules[name] = true
		seeded := strings.Contains(seeds, "'"+name+"'")
		switch {
		case prodOnlyRuleNames[name] && seeded:
			t.Errorf("%q is now seeded - remove it from prodOnlyRuleNames", name)
		case !prodOnlyRuleNames[name] && !seeded:
			t.Errorf("rule name %q is in neither a seed migration nor prodOnlyRuleNames - was it renamed?", name)
		}
	}
	for name := range prodOnlyRuleNames {
		if !rules[name] {
			t.Errorf("prodOnlyRuleNames lists %q, which no rule uses any more", name)
		}
	}
}
