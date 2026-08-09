package services_test

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// fakeReconcilerBackend is an in-memory ports.IStorageBackend whose Walk
// reports whatever's currently in objects, regardless of what the ledger
// thinks - simulating drift between the two.
type fakeReconcilerBackend struct {
	name    string
	objects map[string]int64
}

func (b *fakeReconcilerBackend) Name() string { return b.name }

func (b *fakeReconcilerBackend) Put(context.Context, string, io.Reader, string, int64) error {
	return nil
}

func (b *fakeReconcilerBackend) PresignGet(context.Context, string) (string, error) { return "", nil }

func (b *fakeReconcilerBackend) Del(context.Context, string) error { return nil }

func (b *fakeReconcilerBackend) Walk(_ context.Context, fn func(key string, bytes int64) error) error {
	for key, size := range b.objects {
		if err := fn(key, size); err != nil {
			return err
		}
	}
	return nil
}

var _ ports.IStorageBackend = (*fakeReconcilerBackend)(nil)

// fakeReconcilerLedger is an in-memory ports.IStorageLedger recording only
// what Replace was called with, for reconciler tests.
type fakeReconcilerLedger struct {
	mu      sync.Mutex
	objects map[string]ports.StorageObject
}

func newFakeReconcilerLedger(initial ...ports.StorageObject) *fakeReconcilerLedger {
	l := &fakeReconcilerLedger{objects: make(map[string]ports.StorageObject)}
	for _, obj := range initial {
		l.objects[obj.Key] = obj
	}
	return l
}

func (l *fakeReconcilerLedger) Record(context.Context, ports.StorageObject) error { return nil }
func (l *fakeReconcilerLedger) Forget(context.Context, string) error              { return nil }

func (l *fakeReconcilerLedger) Get(_ context.Context, key string) (ports.StorageObject, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	obj, ok := l.objects[key]
	return obj, ok, nil
}

func (l *fakeReconcilerLedger) Usage(_ context.Context) ([]ports.StorageUsage, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	totals := make(map[string]*ports.StorageUsage)
	for _, obj := range l.objects {
		u, ok := totals[obj.Provider]
		if !ok {
			u = &ports.StorageUsage{Provider: obj.Provider}
			totals[obj.Provider] = u
		}
		u.Bytes += obj.Bytes
		u.Objects++
	}
	usages := make([]ports.StorageUsage, 0, len(totals))
	for _, u := range totals {
		usages = append(usages, *u)
	}
	return usages, nil
}

func (l *fakeReconcilerLedger) Replace(_ context.Context, provider string, objects []ports.StorageObject) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, obj := range l.objects {
		if obj.Provider == provider {
			delete(l.objects, key)
		}
	}
	for _, obj := range objects {
		l.objects[obj.Key] = obj
	}
	return nil
}

var _ ports.IStorageLedger = (*fakeReconcilerLedger)(nil)

// fakeUsageRefresher counts how many times RefreshUsage was called.
type fakeUsageRefresher struct {
	calls int
}

func (f *fakeUsageRefresher) RefreshUsage(context.Context) error {
	f.calls++
	return nil
}

func TestReconcileOnce_CorrectsDriftBothWays(t *testing.T) {
	// The bucket actually has "found-only" (missing from the ledger) and is
	// missing "ledger-only" (which the ledger wrongly still tracks).
	backend := &fakeReconcilerBackend{name: "r2", objects: map[string]int64{"found-only": 42}}
	ledger := newFakeReconcilerLedger(ports.StorageObject{Key: "ledger-only", Provider: "r2", Bytes: 7})
	usage := &fakeUsageRefresher{}

	reconciler := services.NewStorageReconciler([]ports.IStorageBackend{backend}, ledger, usage, 0)
	if err := reconciler.ReconcileOnce(context.Background()); err != nil {
		t.Fatalf("ReconcileOnce: %v", err)
	}

	if _, ok, _ := ledger.Get(context.Background(), "ledger-only"); ok {
		t.Error("ledger-only should have been dropped - it wasn't in the bucket")
	}
	obj, ok, _ := ledger.Get(context.Background(), "found-only")
	if !ok || obj.Bytes != 42 {
		t.Errorf("found-only = %+v, ok=%v, want it recorded with 42 bytes", obj, ok)
	}
	if usage.calls != 1 {
		t.Errorf("RefreshUsage calls = %d, want 1", usage.calls)
	}
}
