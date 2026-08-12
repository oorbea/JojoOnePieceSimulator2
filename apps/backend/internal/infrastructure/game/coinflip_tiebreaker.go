// Package game provides infrastructure implementations of the game-related
// domain ports (ports.ITiebreaker, ports.IAssignmentWeights,
// ports.IGamePowerPool, ports.IStageCatalog). None of it depends on
// entities/game directly beyond what those ports themselves require.
package game

import (
	"context"
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// randIntN is the minimal randomness CoinFlipTiebreaker needs - satisfied
// structurally by ports.RandomGenerator[T] (any T) and by
// infrastructure/random.StdRandomGenerator[T].
type randIntN interface {
	IntN(n int) int
}

// CoinFlipTiebreaker is today's ports.ITiebreaker: a uniform random pick
// among the tied options. An LLM-backed adapter can replace it later
// without any other change (see ports.ITiebreaker's doc comment).
type CoinFlipTiebreaker struct {
	rng randIntN
}

// NewCoinFlipTiebreaker builds a CoinFlipTiebreaker over rng.
func NewCoinFlipTiebreaker(rng randIntN) *CoinFlipTiebreaker {
	return &CoinFlipTiebreaker{rng: rng}
}

var _ ports.ITiebreaker = (*CoinFlipTiebreaker)(nil)

func (t *CoinFlipTiebreaker) Break(_ context.Context, options []string) (string, error) {
	if len(options) == 0 {
		return "", errors.New("no options to break a tie between")
	}
	return options[t.rng.IntN(len(options))], nil
}
