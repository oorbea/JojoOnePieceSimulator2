package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// usageRefresher is the slice of fallback.PictureStorage the reconciler
// needs: re-seed its in-memory quota counters after correcting drift. Kept
// as its own interface (rather than importing the fallback package) so
// services doesn't depend on infrastructure/storage/fallback directly.
type usageRefresher interface {
	RefreshUsage(ctx context.Context) error
}

// StorageReconciler periodically corrects drift between what
// ports.IStorageLedger has on record for each ports.IStorageBackend and what that
// backend's bucket actually contains - the ledger's Record/Forget calls are
// best-effort (see the fallback chain's Upload/Delete), so this is the
// backstop that keeps the quota-check counters honest over time.
type StorageReconciler struct {
	backends []ports.IStorageBackend
	ledger   ports.IStorageLedger
	usage    usageRefresher
	interval time.Duration
}

// NewStorageReconciler builds a reconciler over backends, sharing ledger and
// usage with the fallback chain those same backends are wired into.
// interval <= 0 disables reconciliation (Start becomes a no-op).
func NewStorageReconciler(backends []ports.IStorageBackend, ledger ports.IStorageLedger, usage usageRefresher, interval time.Duration) *StorageReconciler {
	return &StorageReconciler{backends: backends, ledger: ledger, usage: usage, interval: interval}
}

// Start runs ReconcileOnce every interval until ctx is done. It blocks, so
// callers should run it in its own goroutine.
func (r *StorageReconciler) Start(ctx context.Context) {
	if r.interval <= 0 {
		return
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.ReconcileOnce(ctx); err != nil {
				log.Printf("storage reconciliation failed: %v", err)
			}
		}
	}
}

// ReconcileOnce walks every backend's bucket and replaces the ledger's
// inventory for it with what was actually found, logging the corrected
// delta per provider. It exists as its own method so callers (and tests)
// can trigger one reconciliation pass without waiting on Start's ticker.
func (r *StorageReconciler) ReconcileOnce(ctx context.Context) error {
	for _, backend := range r.backends {
		before, err := r.providerUsage(ctx, backend.Name())
		if err != nil {
			return err
		}

		var objects []ports.StorageObject
		var totalBytes int64
		if err := backend.Walk(ctx, func(key string, bytes int64) error {
			objects = append(objects, ports.StorageObject{Key: key, Provider: backend.Name(), Bytes: bytes})
			totalBytes += bytes
			return nil
		}); err != nil {
			return fmt.Errorf("walking %s: %w", backend.Name(), err)
		}

		if err := r.ledger.Replace(ctx, backend.Name(), objects); err != nil {
			return fmt.Errorf("replacing ledger inventory for %s: %w", backend.Name(), err)
		}

		log.Printf("storage reconciliation: %s corrected from %d bytes/%d objects (ledger) to %d bytes/%d objects (bucket)",
			backend.Name(), before.Bytes, before.Objects, totalBytes, len(objects))
	}

	if err := r.usage.RefreshUsage(ctx); err != nil {
		return fmt.Errorf("refreshing usage after reconciliation: %w", err)
	}
	return nil
}

func (r *StorageReconciler) providerUsage(ctx context.Context, provider string) (ports.StorageUsage, error) {
	usages, err := r.ledger.Usage(ctx)
	if err != nil {
		return ports.StorageUsage{}, fmt.Errorf("reading pre-reconciliation usage: %w", err)
	}
	for _, u := range usages {
		if u.Provider == provider {
			return u, nil
		}
	}
	return ports.StorageUsage{Provider: provider}, nil
}
