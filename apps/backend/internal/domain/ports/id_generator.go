package ports

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"

type IIdGenerator interface {
	GeneratePowerId() powers.PowerId
}
