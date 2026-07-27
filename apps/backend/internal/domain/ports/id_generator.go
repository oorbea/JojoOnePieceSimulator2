package ports

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"

type IIDGenerator interface {
	GeneratePowerId() powers.PowerId
}
