package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
)

// IInventory would resolve a player's owned powers for Versus's
// enums.Inventory AbilitySource. There is no persisted inventory yet (no
// schema, no unlock/gacha flow) - game.Config.NewConfig rejects
// enums.Inventory with game.ErrInventoryNotSupported until an adapter for
// this port exists.
type IInventory interface {
	OwnedStands(ctx context.Context, id user.UserID) ([]*powers.Stand, error)
	OwnedDevilFruits(ctx context.Context, id user.UserID) ([]*powers.DevilFruit, error)
}
