package game

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

// LoadoutEvaluator scores a Loadout's aggregate power. It is an interface
// so the scoring heuristic can be swapped (e.g. once real game balance
// data exists) without touching Game, IGameMode, or BotVoter.
type LoadoutEvaluator interface {
	Score(l *Loadout) int
}

// DefaultLoadoutEvaluator sums per-level ability weights, Stand stats, a
// BattleIQ band score and a rarity bonus for the Stand/DevilFruit, if any.
// Every ability weight is an explicit table (spinScore, hamonScore, ...),
// never the enum's raw ordinal, so reshuffling or growing an enum can't
// silently rebalance bot votes.
type DefaultLoadoutEvaluator struct{}

func (DefaultLoadoutEvaluator) Score(l *Loadout) int {
	if l == nil {
		return 0
	}
	score := spinScore(l.Spin()) + hamonScore(l.Hamon()) + fruitMasteryScore(l.FruitMastery()) +
		hakiScore(l.ArmamentHaki()) + hakiScore(l.ObservationHaki()) + conquerorHakiScore(l.ConquerorHaki()) +
		physicalFormScore(l.PhysicalForm()) + battleIQScore(l.BattleIQ())

	if s := l.Stand(); s != nil {
		score += standStatScore(s.AttackPower()) + standStatScore(s.Speed()) +
			standStatScore(s.AttackRange()) + standStatScore(s.Endurance()) +
			standStatScore(s.Precision()) + standStatScore(s.Potential())
		score += rarityBonus(s.Rarity())
	}
	if f := l.DevilFruit(); f != nil {
		score += rarityBonus(f.Rarity())
	}
	return score
}

// spinScore weighs Spin 0/1/3/6: Golden is worth more than one step over
// Basic, and Infinite (Stand-class power) is the single biggest ability tier.
func spinScore(l enums.SpinLevel) int {
	switch l {
	case enums.SpinBasic:
		return 1
	case enums.SpinGolden:
		return 3
	case enums.SpinInfinite:
		return 6
	default:
		return 0
	}
}

// hamonScore weighs Hamon 0/1/2/3.
func hamonScore(l enums.HamonLevel) int {
	switch l {
	case enums.HamonBasic:
		return 1
	case enums.HamonAdvanced:
		return 2
	case enums.HamonPerfect:
		return 3
	default:
		return 0
	}
}

// fruitMasteryScore weighs Devil Fruit mastery 0/1/3/5.
func fruitMasteryScore(m enums.FruitMastery) int {
	switch m {
	case enums.FruitMasteryRegular:
		return 1
	case enums.FruitMasteryAdvanced:
		return 3
	case enums.FruitMasteryAwakened:
		return 5
	default:
		return 0
	}
}

// hakiScore weighs Armament and Observation Haki 0/1/2/3/4.
func hakiScore(h enums.HakiLevel) int {
	switch h {
	case enums.HakiPrivate:
		return 1
	case enums.HakiViceAdmiral:
		return 2
	case enums.HakiYonkoCommander:
		return 3
	case enums.HakiYonkoPlus:
		return 4
	default:
		return 0
	}
}

// conquerorHakiScore weighs Conqueror's Haki 0/1/2/4/6 - rarer and steeper
// at the top than the other two Haki.
func conquerorHakiScore(h enums.HakiLevel) int {
	switch h {
	case enums.HakiPrivate:
		return 1
	case enums.HakiViceAdmiral:
		return 2
	case enums.HakiYonkoCommander:
		return 4
	case enums.HakiYonkoPlus:
		return 6
	default:
		return 0
	}
}

// physicalFormScore weighs PhysicalForm 0/1/2/3/4/5.
func physicalFormScore(f enums.PhysicalForm) int {
	switch f {
	case enums.PhysicalFormStrongFishman:
		return 1
	case enums.PhysicalFormMarineCaptain:
		return 2
	case enums.PhysicalFormViceAdmiral:
		return 3
	case enums.PhysicalFormYonkoCommander:
		return 4
	case enums.PhysicalFormYonkoPlus:
		return 5
	default:
		return 0
	}
}

// standStatScore maps a StandStat to a magnitude: E..A become 1..5,
// Infinite becomes 6 (above A), and Null (a sentinel, not a magnitude)
// becomes 0.
func standStatScore(stat enums.StandStat) int {
	switch stat {
	case enums.E:
		return 1
	case enums.D:
		return 2
	case enums.C:
		return 3
	case enums.B:
		return 4
	case enums.A:
		return 5
	case enums.Infinite:
		return 6
	default:
		return 0
	}
}

// battleIQScore maps a BattleIQ's WAIS-IV band to a magnitude 0..6, the
// same 1..6 scale standStatScore uses - a moderate, sublinear contribution
// deliberately not proportional to the raw 0-255 value, so a 255 doesn't
// dominate the rest of the score (see BattleIQVerySuperior's own internal
// spread in loadout_builder.go, which this collapses to a single point).
// Absent (no JoJo manga in the lobby) contributes 0, same as every other
// participant in that lobby, so relative BotVoter comparisons are
// unaffected.
func battleIQScore(b BattleIQ) int {
	if !b.Present() {
		return 0
	}
	switch b.Band() {
	case BattleIQExtremelyLow:
		return 0
	case BattleIQBorderline:
		return 1
	case BattleIQLowAverage:
		return 2
	case BattleIQAverage:
		return 3
	case BattleIQHighAverage:
		return 4
	case BattleIQSuperior:
		return 5
	default: // BattleIQVerySuperior
		return 6
	}
}

func rarityBonus(r enums.PowerRarity) int {
	switch r {
	case enums.Rare:
		return 1
	case enums.Epic:
		return 2
	case enums.Legendary:
		return 4
	case enums.Mythical:
		return 8
	default:
		return 0
	}
}

// BotVoter casts a bot's Versus vote by comparing each option's aggregate
// LoadoutEvaluator score - the option with the highest combined squad
// score wins the bot's vote.
type BotVoter struct {
	Evaluator LoadoutEvaluator
}

// NewBotVoter builds a BotVoter, defaulting to DefaultLoadoutEvaluator
// when evaluator is nil.
func NewBotVoter(evaluator LoadoutEvaluator) BotVoter {
	if evaluator == nil {
		evaluator = DefaultLoadoutEvaluator{}
	}
	return BotVoter{Evaluator: evaluator}
}

// Vote picks the option with the highest score in scores, breaking ties by
// picking the first (in options order) - deterministic given the same
// inputs.
func (v BotVoter) Vote(options []OptionID, scores map[OptionID]int) OptionID {
	best := options[0]
	bestScore := scores[options[0]]
	for _, o := range options[1:] {
		if s := scores[o]; s > bestScore {
			bestScore = s
			best = o
		}
	}
	return best
}
