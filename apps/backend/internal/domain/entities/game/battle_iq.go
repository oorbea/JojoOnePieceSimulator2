package game

// BattleIQ is a JoJo-only Loadout stat: the raw IQ score behind a
// participant's tactical intelligence, deliberately unbounded (0-255, no
// validation range - allows intentionally absurd values, e.g. 255 for a
// Pucci or a Kars). It is a value object rather than a bare byte/pointer so
// "absent" (no JoJo manga in the lobby) can be represented distinctly from
// a present score of 0, which is itself a legitimate absurd value. The zero
// value of BattleIQ is absent, so every existing call site that doesn't
// know about this stat gets NoBattleIQ() for free.
type BattleIQ struct {
	value   byte
	present bool
}

// NoBattleIQ is the absent BattleIQ - the zero value.
func NoBattleIQ() BattleIQ {
	return BattleIQ{}
}

// NewBattleIQ builds a present BattleIQ holding v. No range validation: the
// full byte range (0-255) is intentional (see the type doc comment).
func NewBattleIQ(v byte) BattleIQ {
	return BattleIQ{value: v, present: true}
}

// Present reports whether this Loadout carries a BattleIQ at all.
func (b BattleIQ) Present() bool {
	return b.present
}

// Value returns the raw 0-255 score. Meaningless when Present() is false.
func (b BattleIQ) Value() byte {
	return b.value
}

// BattleIQBand names one of the seven WAIS-IV classification bands a
// BattleIQ score falls into. It is a game-package type, not an enums one -
// like HakiSet (see weights.go), it is a domain-internal concept the
// frontend re-derives from the raw score rather than one the wire carries,
// so it is deliberately absent from enums.WireEnums / cmd/typegen.
type BattleIQBand byte

const (
	BattleIQExtremelyLow BattleIQBand = iota
	BattleIQBorderline
	BattleIQLowAverage
	BattleIQAverage
	BattleIQHighAverage
	BattleIQSuperior
	BattleIQVerySuperior
)

// Band classifies Value() per the standard Wechsler (WAIS-IV) bands.
// Everything at or above 130 collapses into BattleIQVerySuperior - there is
// no higher WAIS-IV band, that is expected. Meaningless when Present() is
// false.
func (b BattleIQ) Band() BattleIQBand {
	switch {
	case b.value < 70:
		return BattleIQExtremelyLow
	case b.value < 80:
		return BattleIQBorderline
	case b.value < 90:
		return BattleIQLowAverage
	case b.value < 110:
		return BattleIQAverage
	case b.value < 120:
		return BattleIQHighAverage
	case b.value < 130:
		return BattleIQSuperior
	default:
		return BattleIQVerySuperior
	}
}
