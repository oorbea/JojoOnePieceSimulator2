//go:build integration

package repositories_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

// newTestStorageLedger returns a StorageLedger plus the underlying pool, so
// tests can clean up rows directly - IStorageLedger intentionally has no
// bulk-delete method outside of Replace (per-provider).
func newTestStorageLedger(t *testing.T) (*repositories.StorageLedger, *pgxpool.Pool) {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("connecting to database: %v", err)
	}
	t.Cleanup(pool.Close)
	return repositories.NewStorageLedger(pool), pool
}

// cleanupKeys registers a cleanup that deletes the given keys directly,
// regardless of what Delete/Forget did during the test.
func cleanupKeys(t *testing.T, pool *pgxpool.Pool, keys ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, key := range keys {
			if _, err := pool.Exec(context.Background(), "DELETE FROM storage_objects WHERE key = $1", key); err != nil {
				t.Errorf("cleanup delete storage_objects %q: %v", key, err)
			}
		}
	})
}

func TestStorageLedger_RecordAndGet(t *testing.T) {
	ledger, pool := newTestStorageLedger(t)
	ctx := context.Background()
	key := "stands/" + t.Name() + "/main.webp"
	cleanupKeys(t, pool, key)

	if err := ledger.Record(ctx, ports.StorageObject{Key: key, Provider: "r2", Bytes: 12345}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	got, ok, err := ledger.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("Get: ok = false, want true")
	}
	if got.Provider != "r2" || got.Bytes != 12345 {
		t.Errorf("Get = %+v, want {Provider: r2, Bytes: 12345}", got)
	}
}

func TestStorageLedger_Get_NotFoundReturnsOkFalseNoError(t *testing.T) {
	ledger, _ := newTestStorageLedger(t)

	_, ok, err := ledger.Get(context.Background(), "nonexistent-key-"+t.Name())
	if err != nil {
		t.Fatalf("Get: %v, want nil error for a missing key", err)
	}
	if ok {
		t.Error("Get: ok = true, want false for a missing key")
	}
}

func TestStorageLedger_Record_UpsertsOnConflictingKey(t *testing.T) {
	ledger, pool := newTestStorageLedger(t)
	ctx := context.Background()
	key := "stands/" + t.Name() + "/main.webp"
	cleanupKeys(t, pool, key)

	if err := ledger.Record(ctx, ports.StorageObject{Key: key, Provider: "r2", Bytes: 100}); err != nil {
		t.Fatalf("Record (1st): %v", err)
	}
	if err := ledger.Record(ctx, ports.StorageObject{Key: key, Provider: "b2", Bytes: 200}); err != nil {
		t.Fatalf("Record (2nd): %v", err)
	}

	got, ok, err := ledger.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok || got.Provider != "b2" || got.Bytes != 200 {
		t.Errorf("Get = %+v, ok=%v, want the 2nd Record to have overwritten the 1st", got, ok)
	}
}

func TestStorageLedger_Forget_RemovesTheRow(t *testing.T) {
	ledger, pool := newTestStorageLedger(t)
	ctx := context.Background()
	key := "stands/" + t.Name() + "/main.webp"
	cleanupKeys(t, pool, key)

	if err := ledger.Record(ctx, ports.StorageObject{Key: key, Provider: "r2", Bytes: 50}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := ledger.Forget(ctx, key); err != nil {
		t.Fatalf("Forget: %v", err)
	}

	_, ok, err := ledger.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ok {
		t.Error("Get: ok = true, want false after Forget")
	}
}

func TestStorageLedger_Forget_UnknownKeyIsANoOp(t *testing.T) {
	ledger, _ := newTestStorageLedger(t)

	if err := ledger.Forget(context.Background(), "nonexistent-key-"+t.Name()); err != nil {
		t.Fatalf("Forget: %v, want nil error for a key that was never recorded", err)
	}
}

func TestStorageLedger_Usage_AggregatesBytesAndCountsPerProvider(t *testing.T) {
	ledger, pool := newTestStorageLedger(t)
	ctx := context.Background()
	prefix := "stands/" + t.Name() + "/"
	keys := []string{prefix + "a", prefix + "b", prefix + "c"}
	cleanupKeys(t, pool, keys...)

	objs := []ports.StorageObject{
		{Key: keys[0], Provider: "r2", Bytes: 100},
		{Key: keys[1], Provider: "r2", Bytes: 250},
		{Key: keys[2], Provider: "b2", Bytes: 999},
	}
	for _, obj := range objs {
		if err := ledger.Record(ctx, obj); err != nil {
			t.Fatalf("Record(%q): %v", obj.Key, err)
		}
	}

	usages, err := ledger.Usage(ctx)
	if err != nil {
		t.Fatalf("Usage: %v", err)
	}
	byProvider := make(map[string]ports.StorageUsage, len(usages))
	for _, u := range usages {
		byProvider[u.Provider] = u
	}

	// Usage sums over the WHOLE table (shared with other tests/rows that may
	// exist), so only assert this test's own contribution is present via
	// >= rather than an exact total.
	if r2 := byProvider["r2"]; r2.Bytes < 350 || r2.Objects < 2 {
		t.Errorf("r2 usage = %+v, want at least {Bytes: 350, Objects: 2}", r2)
	}
	if b2 := byProvider["b2"]; b2.Bytes < 999 || b2.Objects < 1 {
		t.Errorf("b2 usage = %+v, want at least {Bytes: 999, Objects: 1}", b2)
	}
}

func TestStorageLedger_Replace_SwapsOnlyTheGivenProvidersInventory(t *testing.T) {
	ledger, pool := newTestStorageLedger(t)
	ctx := context.Background()
	prefix := "stands/" + t.Name() + "/"
	r2Old, r2New, b2Untouched := prefix+"r2-old", prefix+"r2-new", prefix+"b2-kept"
	cleanupKeys(t, pool, r2Old, r2New, b2Untouched)

	if err := ledger.Record(ctx, ports.StorageObject{Key: r2Old, Provider: "r2", Bytes: 1}); err != nil {
		t.Fatalf("seed r2Old: %v", err)
	}
	if err := ledger.Record(ctx, ports.StorageObject{Key: b2Untouched, Provider: "b2", Bytes: 2}); err != nil {
		t.Fatalf("seed b2Untouched: %v", err)
	}

	if err := ledger.Replace(ctx, "r2", []ports.StorageObject{{Key: r2New, Provider: "r2", Bytes: 777}}); err != nil {
		t.Fatalf("Replace: %v", err)
	}

	if _, ok, _ := ledger.Get(ctx, r2Old); ok {
		t.Error("r2Old should have been dropped by Replace(r2, ...)")
	}
	if got, ok, _ := ledger.Get(ctx, r2New); !ok || got.Bytes != 777 {
		t.Errorf("r2New = %+v, ok=%v, want it recorded with 777 bytes", got, ok)
	}
	if got, ok, _ := ledger.Get(ctx, b2Untouched); !ok || got.Provider != "b2" {
		t.Errorf("b2Untouched = %+v, ok=%v, want it left alone by a replace scoped to r2", got, ok)
	}
}

func TestStorageLedger_Replace_WithEmptyObjectsClearsTheProvider(t *testing.T) {
	ledger, pool := newTestStorageLedger(t)
	ctx := context.Background()
	key := "stands/" + t.Name() + "/main.webp"
	cleanupKeys(t, pool, key)

	if err := ledger.Record(ctx, ports.StorageObject{Key: key, Provider: "r2", Bytes: 1}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := ledger.Replace(ctx, "r2", nil); err != nil {
		t.Fatalf("Replace: %v", err)
	}

	if _, ok, _ := ledger.Get(ctx, key); ok {
		t.Error("Get: ok = true, want false - Replace with no objects should empty the provider's inventory")
	}
}
