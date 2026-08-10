package game

import (
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// requiresSpin4Names lists the Stands whose assignment must force the
// player's Spin to enums.SpinInfinite. Matching by name is a stopgap - it
// is the single place to update, and the single place to replace, once
// powers gains a persisted traits column (see ADR).
var requiresSpin4Names = map[string]struct{}{
	"Tusk ACT4":             {},
	"Ball Breaker":          {},
	"Soft & Wet: Go Beyond": {},
}

// TraitsOf returns the PowerTraits carried by p. p may be nil (no
// traits), e.g. when a Loadout has no Stand.
func TraitsOf(p *powers.Power) []enums.PowerTrait {
	if p == nil {
		return nil
	}
	if _, ok := requiresSpin4Names[strings.TrimSpace(p.Name())]; ok {
		return []enums.PowerTrait{enums.RequiresSpin4}
	}
	return nil
}

// HasTrait reports whether p carries trait t.
func HasTrait(p *powers.Power, t enums.PowerTrait) bool {
	for _, pt := range TraitsOf(p) {
		if pt == t {
			return true
		}
	}
	return false
}
