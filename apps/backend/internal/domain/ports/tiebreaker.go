package ports

import "context"

// ITiebreaker decides a round that is still tied after its revote window.
// Deliberately typed over plain strings, not game.OptionID, so this port
// never needs to import entities/game - the application layer converts
// game.OptionID <-> string on both sides of the call, then feeds the
// result back into game.Game.ResolveTiebreak. Today's implementation is a
// 50/50 coin flip; an LLM-backed adapter can replace it later without any
// other change.
type ITiebreaker interface {
	// Break picks and returns one of options.
	Break(ctx context.Context, options []string) (winner string, err error)
}
