package ports

import "context"

// StorageObject is one object tracked by an IStorageLedger: which key, on
// which provider, occupying how many bytes.
type StorageObject struct {
	Key      string
	Provider string
	Bytes    int64
}

// StorageUsage is the aggregate bytes/object count an IStorageLedger has on
// record for one provider - what the storage fallback chain checks against
// each provider's quota before deciding where an upload lands.
type StorageUsage struct {
	Provider string
	Bytes    int64
	Objects  int64
}

// IStorageLedger is the single source of truth for which provider a given
// object-storage key lives on, and how many bytes each provider currently
// holds. It exists because the fallback chain lets objects stay put once
// written (no migration on quota change), so PresignGetURL/Delete need a way
// to find an already-stored key's provider without guessing.
type IStorageLedger interface {
	// Record notes that obj.Key lives on obj.Provider occupying obj.Bytes,
	// replacing whatever was recorded for that key before.
	Record(ctx context.Context, obj StorageObject) error
	// Forget removes key from the ledger. Forgetting a key that isn't
	// tracked is not an error.
	Forget(ctx context.Context, key string) error
	// Get returns the StorageObject last Recorded under key. ok is false if
	// key isn't tracked (e.g. it predates the ledger).
	Get(ctx context.Context, key string) (obj StorageObject, ok bool, err error)
	// Usage returns the current bytes/object count per provider that has at
	// least one tracked object.
	Usage(ctx context.Context) ([]StorageUsage, error)
	// Replace atomically swaps everything tracked for provider with objects,
	// used by the reconciler to correct drift against what the bucket
	// actually contains.
	Replace(ctx context.Context, provider string, objects []StorageObject) error
}
