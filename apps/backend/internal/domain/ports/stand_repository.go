package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type StandFilters struct {
	Rarity      *enums.PowerRarity
	AttackPower *enums.StandStat
	Speed       *enums.StandStat
	AttackRange *enums.StandStat
	Endurance   *enums.StandStat
	Precision   *enums.StandStat
	Potential   *enums.StandStat
	EvolvesFrom *string
}

type IStandRepository interface {
	Save(ctx context.Context, stand *powers.Stand) error
	FindByID(ctx context.Context, id powers.PowerID) (*powers.Stand, error)
	FindByName(ctx context.Context, name string) (*powers.Stand, error)
	GetAll(ctx context.Context) ([]*powers.Stand, error)
	Filter(ctx context.Context, filters StandFilters) ([]*powers.Stand, error)
	Delete(ctx context.Context, id powers.PowerID) error
	// UpdatePicture updates only a stand's picture renditions and pipeline
	// status. A nil main or thumb leaves that column untouched.
	UpdatePicture(ctx context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error
}
