package game

// RandomSource is the minimal randomness primitive the domain needs
// (weighted power/level draws, picking a random Versus Stage). It is
// declared locally, instead of importing ports.RandomGenerator, so that
// entities/game never depends on the ports package (ports already depends
// on entities/game for Stage/GameResult, and Go forbids the cycle). Any
// type satisfying ports.RandomGenerator[T] for any T already satisfies
// this interface structurally - infrastructure/random.StdRandomGenerator[T]
// can be passed in as-is.
type RandomSource interface {
	IntN(n int) int
}

// weightedPick returns an index into weights chosen with probability
// proportional to its (non-negative) value. If every weight is zero (or
// weights is such that the total is non-positive), it falls back to a
// uniform pick so a misconfigured weight table never panics or always
// picks index 0.
func weightedPick(rng RandomSource, weights []int) int {
	total := 0
	for _, w := range weights {
		if w > 0 {
			total += w
		}
	}
	if total <= 0 {
		return rng.IntN(len(weights))
	}
	r := rng.IntN(total)
	cum := 0
	for i, w := range weights {
		if w > 0 {
			cum += w
		}
		if r < cum {
			return i
		}
	}
	return len(weights) - 1
}
