package game

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

// LoadoutEvaluator scores a Loadout's aggregate power. It is an interface
// so the scoring heuristic can be swapped (e.g. once real game balance
// data exists) without touching Game, IGameMode, or BotVoter.
type LoadoutEvaluator interface {
	Score(l *Loadout) int
}

// DefaultLoadoutEvaluator sums normalized ability levels plus Stand stats
// and a rarity bonus for the Stand/DevilFruit, if any.
type DefaultLoadoutEvaluator struct{}

func (DefaultLoadoutEvaluator) Score(l *Loadout) int {
	if l == nil {
		return 0
	}
	score := int(l.Spin()) + int(l.Hamon()) + int(l.FruitMastery()) +
		int(l.ArmamentHaki()) + int(l.ObservationHaki()) + int(l.ConquerorHaki()) +
		int(l.PhysicalForm())

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

func rarityBonus(r enums.PowerRarity) int {
	switch r {
	case enums.Rare:
		return 1
	case enums.Epic:
		return 2
	case enums.Legendary:
		return 3
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
