package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type DevilFruitFilters struct {
	Rarity    *enums.PowerRarity
	FruitType *enums.FruitType
}

type IDevilFruitRepository interface {
	Save(ctx context.Context, fruit *powers.DevilFruit) error
	FindByID(ctx context.Context, id powers.PowerID) (*powers.DevilFruit, error)
	FindByName(ctx context.Context, name string) (*powers.DevilFruit, error)
	GetAll(ctx context.Context) ([]*powers.DevilFruit, error)
	Filter(ctx context.Context, filters DevilFruitFilters) ([]*powers.DevilFruit, error)
	Delete(ctx context.Context, id powers.PowerID) error
	// UpdatePicture updates only a devil fruit's picture renditions and
	// pipeline status. A nil main or thumb leaves that column untouched.
	UpdatePicture(ctx context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error
}
