package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// MediaRepository is the pgx/sqlc adapter for ports.IMediaRepository.
type MediaRepository struct {
	queries *db.Queries
}

var _ ports.IMediaRepository = (*MediaRepository)(nil)

// NewMediaRepository builds a MediaRepository over pool - same construction
// shape as every other repository in this package, even though this one
// never needs a transaction (every write here is a single-row upsert with
// no cross-table invariant to keep atomic).
func NewMediaRepository(pool *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{queries: db.New(pool)}
}

// PutMediaObjects implements ports.IMediaRepository.
func (r *MediaRepository) PutMediaObjects(ctx context.Context, objects []ports.MediaObject) error {
	for _, o := range objects {
		if err := r.queries.UpsertMediaObject(ctx, db.UpsertMediaObjectParams{
			GroupID:     o.GroupID,
			Variant:     o.Variant,
			StorageKey:  o.StorageKey,
			ContentType: o.ContentType,
			Bytes:       o.Bytes,
			Scope:       o.Scope,
		}); err != nil {
			return fmt.Errorf("upserting media object %s/%s: %w", o.GroupID, o.Variant, err)
		}
	}
	return nil
}

// GetMediaObjectsByGroup implements ports.IMediaRepository.
func (r *MediaRepository) GetMediaObjectsByGroup(ctx context.Context, groupID string) ([]ports.MediaObject, error) {
	rows, err := r.queries.GetMediaObjectsByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("querying media objects for group %s: %w", groupID, err)
	}
	objects := make([]ports.MediaObject, len(rows))
	for i, row := range rows {
		objects[i] = ports.MediaObject{
			GroupID: row.GroupID, Variant: row.Variant, StorageKey: row.StorageKey,
			ContentType: row.ContentType, Bytes: row.Bytes, Scope: row.Scope,
		}
	}
	return objects, nil
}
