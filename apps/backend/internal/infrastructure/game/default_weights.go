package game

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// DefaultWeights is today's ports.IAssignmentWeights: it always returns
// game.DefaultAssignmentWeights(), a uniform-ish reference table. The
// weighting policy is meant to be admin-configurable later; this adapter is
// the placeholder until that exists (see ports.IAssignmentWeights).
type DefaultWeights struct{}

// NewDefaultWeights builds a DefaultWeights adapter.
func NewDefaultWeights() DefaultWeights { return DefaultWeights{} }

var _ ports.IAssignmentWeights = DefaultWeights{}

func (DefaultWeights) Load(context.Context) (game.AssignmentWeights, error) {
	return game.DefaultAssignmentWeights(), nil
}
