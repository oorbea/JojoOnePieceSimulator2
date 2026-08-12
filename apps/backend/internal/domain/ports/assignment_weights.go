package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

// IAssignmentWeights resolves the weighted-draw policy game.LoadoutBuilder
// uses (probability of no stand/no fruit, rarity skew, per-level weights -
// including the deliberately-lower weight for Conqueror Haki). It is a
// port purely because the policy is meant to be admin-configurable later;
// until that adapter exists, game.DefaultAssignmentWeights is a reasonable
// implementation to return.
type IAssignmentWeights interface {
	Load(ctx context.Context) (game.AssignmentWeights, error)
}
