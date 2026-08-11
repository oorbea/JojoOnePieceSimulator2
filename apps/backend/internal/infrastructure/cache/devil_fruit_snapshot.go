package cache

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/powersnap"
)

// marshalDevilFruit/marshalDevilFruits/unmarshalDevilFruit/
// unmarshalDevilFruits are thin wrappers over infrastructure/powersnap - see
// stand_snapshot.go's doc comment.

func marshalDevilFruit(fruit *powers.DevilFruit) ([]byte, error) {
	return powersnap.MarshalDevilFruit(fruit)
}

func unmarshalDevilFruit(data []byte) (*powers.DevilFruit, error) {
	return powersnap.UnmarshalDevilFruit(data)
}

func marshalDevilFruits(fruits []*powers.DevilFruit) ([]byte, error) {
	return powersnap.MarshalDevilFruits(fruits)
}

func unmarshalDevilFruits(data []byte) ([]*powers.DevilFruit, error) {
	return powersnap.UnmarshalDevilFruits(data)
}
