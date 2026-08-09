package services_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// fakeReconcilerBackend is an in-memory ports.IStorageBackend whose Walk
// reports whatever's currently in objects, regardless of what the ledger
// thinks - simulating drift between the two.
type fakeReconcilerBackend struct {
	name    string
	objects map[string]int64
	walkErr error
}

func (b *fakeReconcilerBackend) Name() string { return b.name }

func (b *fakeReconcilerBackend) Put(context.Context, string, io.Reader, string, int64) error {
	return nil
}

func (b *fakeReconcilerBackend) PresignGet(context.Context, string) (string, error) { return "", nil }

func (b *fakeReconcilerBackend) Del(context.Context, string) error { return nil }

func (b *fakeReconcilerBackend) Walk(_ context.Context, fn func(key string, bytes int64) error) error {
	if b.walkErr != nil {
		return b.walkErr
	}
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
	mu         sync.Mutex
	objects    map[string]ports.StorageObject
	usageErr   error
	replaceErr error
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
	if l.usageErr != nil {
		return nil, l.usageErr
	}
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
	if l.replaceErr != nil {
		return l.replaceErr
	}
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
	err   error
}

func (f *fakeUsageRefresher) RefreshUsage(context.Context) error {
	f.calls++
	return f.err
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

func TestReconcileOnce_PropagatesWalkError(t *testing.T) {
	backend := &fakeReconcilerBackend{name: "r2", walkErr: errors.New("bucket unreachable")}
	ledger := newFakeReconcilerLedger()
	usage := &fakeUsageRefresher{}

	reconciler := services.NewStorageReconciler([]ports.IStorageBackend{backend}, ledger, usage, 0)
	if err := reconciler.ReconcileOnce(context.Background()); err == nil {
		t.Fatal("ReconcileOnce: want error when Walk fails, got nil")
	}
	if usage.calls != 0 {
		t.Errorf("RefreshUsage calls = %d, want 0 - a failed pass must not refresh usage", usage.calls)
	}
}

func TestReconcileOnce_PropagatesReplaceError(t *testing.T) {
	backend := &fakeReconcilerBackend{name: "r2", objects: map[string]int64{"a": 1}}
	ledger := newFakeReconcilerLedger()
	ledger.replaceErr = errors.New("db down")
	usage := &fakeUsageRefresher{}

	reconciler := services.NewStorageReconciler([]ports.IStorageBackend{backend}, ledger, usage, 0)
	if err := reconciler.ReconcileOnce(context.Background()); err == nil {
		t.Fatal("ReconcileOnce: want error when Replace fails, got nil")
	}
}

func TestReconcileOnce_PropagatesUsageRefreshError(t *testing.T) {
	backend := &fakeReconcilerBackend{name: "r2", objects: map[string]int64{"a": 1}}
	ledger := newFakeReconcilerLedger()
	usage := &fakeUsageRefresher{err: errors.New("refresh failed")}

	reconciler := services.NewStorageReconciler([]ports.IStorageBackend{backend}, ledger, usage, 0)
	if err := reconciler.ReconcileOnce(context.Background()); err == nil {
		t.Fatal("ReconcileOnce: want error when RefreshUsage fails, got nil")
	}
}

func TestReconcileOnce_PropagatesPreReconciliationUsageError(t *testing.T) {
	backend := &fakeReconcilerBackend{name: "r2", objects: map[string]int64{"a": 1}}
	ledger := newFakeReconcilerLedger()
	ledger.usageErr = errors.New("db down")
	usage := &fakeUsageRefresher{}

	reconciler := services.NewStorageReconciler([]ports.IStorageBackend{backend}, ledger, usage, 0)
	if err := reconciler.ReconcileOnce(context.Background()); err == nil {
		t.Fatal("ReconcileOnce: want error when the pre-reconciliation Usage read fails, got nil")
	}
}

func TestReconcileOnce_MultipleBackendsEachReplacedIndependently(t *testing.T) {
	r2 := &fakeReconcilerBackend{name: "r2", objects: map[string]int64{"r2-key": 5}}
	b2 := &fakeReconcilerBackend{name: "b2", objects: map[string]int64{"b2-key": 9}}
	ledger := newFakeReconcilerLedger(
		ports.StorageObject{Key: "stale-r2", Provider: "r2", Bytes: 1},
		ports.StorageObject{Key: "stale-b2", Provider: "b2", Bytes: 2},
	)
	usage := &fakeUsageRefresher{}

	reconciler := services.NewStorageReconciler([]ports.IStorageBackend{r2, b2}, ledger, usage, 0)
	if err := reconciler.ReconcileOnce(context.Background()); err != nil {
		t.Fatalf("ReconcileOnce: %v", err)
	}

	if _, ok, _ := ledger.Get(context.Background(), "stale-r2"); ok {
		t.Error("stale-r2 should have been dropped")
	}
	if _, ok, _ := ledger.Get(context.Background(), "stale-b2"); ok {
		t.Error("stale-b2 should have been dropped")
	}
	if obj, ok, _ := ledger.Get(context.Background(), "r2-key"); !ok || obj.Bytes != 5 {
		t.Errorf("r2-key = %+v, ok=%v, want it recorded with 5 bytes on r2", obj, ok)
	}
	if obj, ok, _ := ledger.Get(context.Background(), "b2-key"); !ok || obj.Bytes != 9 {
		t.Errorf("b2-key = %+v, ok=%v, want it recorded with 9 bytes on b2", obj, ok)
	}
}

func TestStart_DisabledWhenIntervalZero(t *testing.T) {
	backend := &fakeReconcilerBackend{name: "r2", objects: map[string]int64{"a": 1}}
	ledger := newFakeReconcilerLedger()
	usage := &fakeUsageRefresher{}
	reconciler := services.NewStorageReconciler([]ports.IStorageBackend{backend}, ledger, usage, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	reconciler.Start(ctx) // must return immediately, not block until ctx times out

	if usage.calls != 0 {
		t.Errorf("RefreshUsage calls = %d, want 0 - interval<=0 must disable Start entirely", usage.calls)
	}
}

func TestStart_RunsPeriodicallyAndStopsOnContextCancel(t *testing.T) {
	backend := &fakeReconcilerBackend{name: "r2", objects: map[string]int64{"a": 1}}
	ledger := newFakeReconcilerLedger()
	usage := &fakeUsageRefresher{}
	reconciler := services.NewStorageReconciler([]ports.IStorageBackend{backend}, ledger, usage, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		reconciler.Start(ctx)
		close(done)
	}()

	time.Sleep(35 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Start did not return after ctx was cancelled")
	}
	if usage.calls == 0 {
		t.Error("RefreshUsage calls = 0, want at least 1 tick to have fired before cancellation")
	}
}
