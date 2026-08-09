package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// StorageLedger is the pgx/sqlc adapter for ports.IStorageLedger, backed by
// the storage_objects table.
type StorageLedger struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

var _ ports.IStorageLedger = (*StorageLedger)(nil)

func NewStorageLedger(pool *pgxpool.Pool) *StorageLedger {
	return &StorageLedger{pool: pool, queries: db.New(pool)}
}

// Record implements ports.IStorageLedger.
func (r *StorageLedger) Record(ctx context.Context, obj ports.StorageObject) error {
	if err := r.queries.RecordStorageObject(ctx, db.RecordStorageObjectParams{
		Key:      obj.Key,
		Provider: obj.Provider,
		Bytes:    obj.Bytes,
	}); err != nil {
		return fmt.Errorf("recording storage object %q: %w", obj.Key, err)
	}
	return nil
}

// Forget implements ports.IStorageLedger.
func (r *StorageLedger) Forget(ctx context.Context, key string) error {
	if err := r.queries.ForgetStorageObject(ctx, key); err != nil {
		return fmt.Errorf("forgetting storage object %q: %w", key, err)
	}
	return nil
}

// Get implements ports.IStorageLedger.
func (r *StorageLedger) Get(ctx context.Context, key string) (ports.StorageObject, bool, error) {
	row, err := r.queries.GetStorageObject(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.StorageObject{}, false, nil
		}
		return ports.StorageObject{}, false, fmt.Errorf("querying storage object %q: %w", key, err)
	}
	return ports.StorageObject{Key: row.Key, Provider: row.Provider, Bytes: row.Bytes}, true, nil
}

// Usage implements ports.IStorageLedger.
func (r *StorageLedger) Usage(ctx context.Context) ([]ports.StorageUsage, error) {
	rows, err := r.queries.SumStorageUsage(ctx)
	if err != nil {
		return nil, fmt.Errorf("summing storage usage: %w", err)
	}
	usages := make([]ports.StorageUsage, 0, len(rows))
	for _, row := range rows {
		usages = append(usages, ports.StorageUsage{Provider: row.Provider, Bytes: row.Bytes, Objects: row.Objects})
	}
	return usages, nil
}

// Replace implements ports.IStorageLedger, swapping provider's entire
// inventory for objects inside one transaction.
func (r *StorageLedger) Replace(ctx context.Context, provider string, objects []ports.StorageObject) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			fmt.Printf("rolling back transaction replacing storage objects for %q: %v\n", provider, rollbackErr)
		}
	}()

	q := r.queries.WithTx(tx)

	if err := q.DeleteStorageObjectsByProvider(ctx, provider); err != nil {
		return fmt.Errorf("clearing storage objects for %q: %w", provider, err)
	}
	for _, obj := range objects {
		if err := q.RecordStorageObject(ctx, db.RecordStorageObjectParams{
			Key:      obj.Key,
			Provider: provider,
			Bytes:    obj.Bytes,
		}); err != nil {
			return fmt.Errorf("recording storage object %q for %q: %w", obj.Key, provider, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing storage objects replace for %q: %w", provider, err)
	}
	return nil
}
