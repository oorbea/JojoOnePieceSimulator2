// Package cache decorates ports.IStandRepository and ports.IPictureStorage
// with a ports.ICache, so read-heavy calls can skip Postgres/R2 entirely on
// a hit. Both decorators are fail-open: any cache error is treated as a
// miss, never as a request failure.
package cache

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/powersnap"
)

// marshalStand/marshalStands/unmarshalStand/unmarshalStands are thin
// wrappers over infrastructure/powersnap, which holds the actual
// (de)serialization logic - promoted out of this package so it can also be
// shared by the game store, whose Loadout snapshots embed full powers (see
// entities/game.LoadoutSnapshot). Kept here so stand_repository.go's call
// sites and this package's own tests don't need to change.

func marshalStand(stand *powers.Stand) ([]byte, error) {
	return powersnap.MarshalStand(stand)
}

func unmarshalStand(data []byte) (*powers.Stand, error) {
	return powersnap.UnmarshalStand(data)
}

func marshalStands(stands []*powers.Stand) ([]byte, error) {
	return powersnap.MarshalStands(stands)
}

func unmarshalStands(data []byte) ([]*powers.Stand, error) {
	return powersnap.UnmarshalStands(data)
}
