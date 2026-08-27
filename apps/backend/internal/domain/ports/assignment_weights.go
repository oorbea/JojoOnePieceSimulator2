package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

// IAssignmentWeights resolves the weighted-draw policy game.LoadoutBuilder
// uses (probability of no stand/no fruit, per-level weights, the haki
// set/mastery tables). It is a port purely because the policy is meant to
// be admin-configurable later; until that adapter exists,
// game.DefaultAssignmentWeights - a 1:1 port of JoJoOnePiece_Simulator V1's
// probabilities - is a reasonable implementation to return.
type IAssignmentWeights interface {
	Load(ctx context.Context) (game.AssignmentWeights, error)
}
