package fallback_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/storage/fallback"
)

// fakeBackend is an in-memory ports.IStorageBackend, so the fallback chain
// can be tested without a real S3-compatible bucket.
type fakeBackend struct {
	mu      sync.Mutex
	name    string
	objects map[string][]byte
	putErr  error
}

func newFakeBackend(name string) *fakeBackend {
	return &fakeBackend{name: name, objects: make(map[string][]byte)}
}

func (b *fakeBackend) Name() string { return b.name }

func (b *fakeBackend) Put(_ context.Context, key string, content io.Reader, _ string, _ int64) error {
	if b.putErr != nil {
		return b.putErr
	}
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.objects[key] = data
	return nil
}

func (b *fakeBackend) PresignGet(_ context.Context, key string) (string, error) {
	return "https://" + b.name + ".test/" + key, nil
}

func (b *fakeBackend) Del(_ context.Context, key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.objects, key)
	return nil
}

func (b *fakeBackend) Walk(_ context.Context, fn func(key string, bytes int64) error) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for key, data := range b.objects {
		if err := fn(key, int64(len(data))); err != nil {
			return err
		}
	}
	return nil
}

func (b *fakeBackend) has(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.objects[key]
	return ok
}

var _ ports.IStorageBackend = (*fakeBackend)(nil)

// fakeLedger is an in-memory ports.IStorageLedger.
type fakeLedger struct {
	mu      sync.Mutex
	objects map[string]ports.StorageObject
}

func newFakeLedger() *fakeLedger {
	return &fakeLedger{objects: make(map[string]ports.StorageObject)}
}

func (l *fakeLedger) Record(_ context.Context, obj ports.StorageObject) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.objects[obj.Key] = obj
	return nil
}

func (l *fakeLedger) Forget(_ context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.objects, key)
	return nil
}

func (l *fakeLedger) Get(_ context.Context, key string) (ports.StorageObject, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	obj, ok := l.objects[key]
	return obj, ok, nil
}

func (l *fakeLedger) Usage(_ context.Context) ([]ports.StorageUsage, error) {
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

func (l *fakeLedger) Replace(_ context.Context, provider string, objects []ports.StorageObject) error {
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

var _ ports.IStorageLedger = (*fakeLedger)(nil)

func pic(content string) ports.Picture {
	return ports.Picture{Content: bytes.NewReader([]byte(content)), ContentType: "image/webp", Size: int64(len(content))}
}

func TestUpload_LandsOnFirstTierWithRoom(t *testing.T) {
	r2, b2 := newFakeBackend("r2"), newFakeBackend("b2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 1000},
		{Backend: b2, QuotaBytes: 1000},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	stored, err := chain.Upload(context.Background(), "stands/1/main.webp", pic("hello"))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if stored.Provider != "r2" {
		t.Errorf("Provider = %q, want %q", stored.Provider, "r2")
	}
	if !r2.has("stands/1/main.webp") {
		t.Error("r2 should have the object")
	}
}

func TestUpload_SkipsTierOverQuota(t *testing.T) {
	r2, b2 := newFakeBackend("r2"), newFakeBackend("b2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 10}, // threshold at 95% of 10 = 9 bytes, "hello" is 5 - fits once
		{Backend: b2, QuotaBytes: 1000},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	if _, err := chain.Upload(ctx, "a", pic("hello")); err != nil {
		t.Fatalf("first Upload: %v", err)
	}
	// r2 now has 5/10 bytes; a second 5-byte upload would push it to 10,
	// over the 9-byte (95%) threshold, so it must fall through to b2.
	stored, err := chain.Upload(ctx, "b", pic("world"))
	if err != nil {
		t.Fatalf("second Upload: %v", err)
	}
	if stored.Provider != "b2" {
		t.Errorf("Provider = %q, want %q (r2 should be over quota)", stored.Provider, "b2")
	}
}

func TestUpload_FallsThroughOnPutError(t *testing.T) {
	r2, b2 := newFakeBackend("r2"), newFakeBackend("b2")
	r2.putErr = errors.New("network blip")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 1000},
		{Backend: b2, QuotaBytes: 1000},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	stored, err := chain.Upload(context.Background(), "a", pic("hello"))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if stored.Provider != "b2" {
		t.Errorf("Provider = %q, want %q (r2's Put should have failed)", stored.Provider, "b2")
	}
	if !b2.has("a") {
		t.Error("b2 should have the object")
	}
}

func TestUpload_AllTiersExhaustedReturnsErrStorageExhausted(t *testing.T) {
	r2 := newFakeBackend("r2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 1},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = chain.Upload(context.Background(), "a", pic("hello"))
	if !errors.Is(err, fallback.ErrStorageExhausted) {
		t.Fatalf("err = %v, want ErrStorageExhausted", err)
	}
}

func TestUpload_PreferProviderOverridesOrder(t *testing.T) {
	r2, b2 := newFakeBackend("r2"), newFakeBackend("b2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 1000},
		{Backend: b2, QuotaBytes: 1000},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	p := pic("hello")
	p.PreferProvider = "b2"
	stored, err := chain.Upload(context.Background(), "thumb", p)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if stored.Provider != "b2" {
		t.Errorf("Provider = %q, want %q (PreferProvider should win)", stored.Provider, "b2")
	}
	if !b2.has("thumb") {
		t.Error("b2 should have the object despite being second in the chain")
	}
}

func TestDelete_UsesRecordedProvider(t *testing.T) {
	r2, b2 := newFakeBackend("r2"), newFakeBackend("b2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 0},
		{Backend: b2, QuotaBytes: 0},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	p := pic("hello")
	p.PreferProvider = "b2"
	if _, err := chain.Upload(ctx, "a", p); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := chain.Delete(ctx, "a"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if b2.has("a") {
		t.Error("b2 should no longer have the object")
	}
}

func TestPresignGetURL_UnknownKeyDefaultsToFirstTier(t *testing.T) {
	r2, b2 := newFakeBackend("r2"), newFakeBackend("b2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 0},
		{Backend: b2, QuotaBytes: 0},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	url, err := chain.PresignGetURL(context.Background(), "stands/legacy/main.webp")
	if err != nil {
		t.Fatalf("PresignGetURL: %v", err)
	}
	if url != "https://r2.test/stands/legacy/main.webp" {
		t.Errorf("url = %q, want the first tier (r2) to serve an unrecorded key", url)
	}
}

func TestUpload_NonSeekableContentIsBufferedAndRetried(t *testing.T) {
	r2, b2 := newFakeBackend("r2"), newFakeBackend("b2")
	r2.putErr = errors.New("boom")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 1000},
		{Backend: b2, QuotaBytes: 1000},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// bytes.NewBuffer's *Buffer only implements io.Reader (not io.ReadSeeker
	// - it drains as it's read), forcing the chain's buffering path.
	nonSeekable := bytes.NewBuffer([]byte("hello"))
	stored, err := chain.Upload(context.Background(), "a", ports.Picture{
		Content: nonSeekable, ContentType: "image/webp", Size: 5,
	})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if stored.Provider != "b2" {
		t.Fatalf("Provider = %q, want %q", stored.Provider, "b2")
	}
	if got := string(b2.objects["a"]); got != "hello" {
		t.Errorf("stored content = %q, want %q (full body should have reached the second tier)", got, "hello")
	}
}

func TestRefreshUsage_ReflectsLedgerAfterReplace(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 10},
	}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	// Simulate drift: the reconciler found more on the bucket than the
	// ledger had on record, via Replace.
	if err := ledger.Replace(ctx, "r2", []ports.StorageObject{{Key: "x", Provider: "r2", Bytes: 9}}); err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if err := chain.RefreshUsage(ctx); err != nil {
		t.Fatalf("RefreshUsage: %v", err)
	}

	// r2 is now at 9/10 bytes (over the 95%-of-10=9 threshold once this
	// upload's 1 byte is added), so the next upload must fall through - but
	// there's nowhere to fall through to, so it must be rejected.
	_, err = chain.Upload(ctx, "y", pic("z"))
	if !errors.Is(err, fallback.ErrStorageExhausted) {
		t.Fatalf("err = %v, want ErrStorageExhausted (RefreshUsage should have picked up the drift)", err)
	}
}
