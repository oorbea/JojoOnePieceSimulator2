package fallback_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	delErr  error
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
	if b.delErr != nil {
		return b.delErr
	}
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
	mu        sync.Mutex
	objects   map[string]ports.StorageObject
	recordErr error
	getErr    error
	forgetErr error
	usageErr  error
}

func newFakeLedger() *fakeLedger {
	return &fakeLedger{objects: make(map[string]ports.StorageObject)}
}

func (l *fakeLedger) Record(_ context.Context, obj ports.StorageObject) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.recordErr != nil {
		return l.recordErr
	}
	l.objects[obj.Key] = obj
	return nil
}

func (l *fakeLedger) Forget(_ context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.forgetErr != nil {
		return l.forgetErr
	}
	delete(l.objects, key)
	return nil
}

func (l *fakeLedger) Get(_ context.Context, key string) (ports.StorageObject, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.getErr != nil {
		return ports.StorageObject{}, false, l.getErr
	}
	obj, ok := l.objects[key]
	return obj, ok, nil
}

func (l *fakeLedger) Usage(_ context.Context) ([]ports.StorageUsage, error) {
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

// --- New: construction ---

func TestNew_RejectsEmptyTiers(t *testing.T) {
	if _, err := fallback.New(context.Background(), nil, newFakeLedger(), 95); err == nil {
		t.Fatal("New: want error for zero tiers, got nil")
	}
}

func TestNew_RejectsInvalidThresholdPct(t *testing.T) {
	r2 := newFakeBackend("r2")
	for _, pct := range []int{0, -1, 101} {
		t.Run(fmt.Sprintf("pct_%d", pct), func(t *testing.T) {
			_, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 1000}}, newFakeLedger(), pct)
			if err == nil {
				t.Fatalf("New: want error for thresholdPct=%d, got nil", pct)
			}
		})
	}
}

func TestNew_PropagatesLedgerUsageError(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	ledger.usageErr = errors.New("db down")

	if _, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 1000}}, ledger, 95); err == nil {
		t.Fatal("New: want error when the ledger's Usage fails to seed counters, got nil")
	}
}

// --- Upload: quota boundary ---

func TestUpload_ExactlyAtThresholdFits(t *testing.T) {
	r2 := newFakeBackend("r2")
	// threshold 95% of 100 = 95 bytes; a 95-byte upload lands exactly on the
	// boundary and must still fit.
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 100}}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	content := bytes.Repeat([]byte("x"), 95)
	stored, err := chain.Upload(context.Background(), "a", ports.Picture{Content: bytes.NewReader(content), ContentType: "image/webp", Size: 95})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if stored.Provider != "r2" {
		t.Errorf("Provider = %q, want %q (exactly-at-threshold upload should fit)", stored.Provider, "r2")
	}
}

func TestUpload_OneByteOverThresholdRejected(t *testing.T) {
	r2 := newFakeBackend("r2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 100}}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	content := bytes.Repeat([]byte("x"), 96)
	_, err = chain.Upload(context.Background(), "a", ports.Picture{Content: bytes.NewReader(content), ContentType: "image/webp", Size: 96})
	if !errors.Is(err, fallback.ErrStorageExhausted) {
		t.Fatalf("err = %v, want ErrStorageExhausted (96 bytes is 1 over the 95-byte threshold)", err)
	}
}

func TestUpload_ZeroQuotaMeansUnlimited(t *testing.T) {
	r2 := newFakeBackend("r2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 0}}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	huge := bytes.Repeat([]byte("x"), 10_000_000)
	stored, err := chain.Upload(context.Background(), "a", ports.Picture{Content: bytes.NewReader(huge), ContentType: "image/webp", Size: int64(len(huge))})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if stored.Provider != "r2" {
		t.Errorf("Provider = %q, want %q (0 quota should never reject on size)", stored.Provider, "r2")
	}
}

// --- Upload: ledger best-effort semantics ---

func TestUpload_SucceedsEvenWhenLedgerRecordFails(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	ledger.recordErr = errors.New("db unreachable")
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 1000}}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	stored, err := chain.Upload(context.Background(), "a", pic("hello"))
	if err != nil {
		t.Fatalf("Upload: %v (a Put success must not fail just because Record did)", err)
	}
	if stored.Provider != "r2" || stored.Key != "a" {
		t.Errorf("stored = %+v, want {r2 a}", stored)
	}
	if !r2.has("a") {
		t.Error("the object should still be on r2 despite the ledger write failing")
	}
}

func TestUpload_UsageCounterNotIncrementedWhenLedgerRecordFails(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	ledger.recordErr = errors.New("db unreachable")
	// Quota tight enough that a second upload only fits if the first's bytes
	// were (wrongly) counted.
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 6}}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	if _, err := chain.Upload(ctx, "a", pic("hello")); err != nil {
		t.Fatalf("first Upload: %v", err)
	}
	// If Record's failure had still bumped the in-memory counter, "hello"
	// (5 bytes) would already be at the 95%-of-6≈5-byte threshold and this
	// second upload would be rejected. It must succeed, proving the counter
	// stayed at 0 - accounting is deferred to the next reconciliation pass.
	if _, err := chain.Upload(ctx, "b", pic("hi")); err != nil {
		t.Fatalf("second Upload: %v (usage counter should not have advanced on a failed Record)", err)
	}
}

// --- Upload: PreferProvider edge cases ---

func TestUpload_UnknownPreferProviderFallsBackToDefaultOrder(t *testing.T) {
	r2, b2 := newFakeBackend("r2"), newFakeBackend("b2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 1000},
		{Backend: b2, QuotaBytes: 1000},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	p := pic("hello")
	p.PreferProvider = "supabase" // not in this chain
	stored, err := chain.Upload(context.Background(), "a", p)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if stored.Provider != "r2" {
		t.Errorf("Provider = %q, want %q (unknown PreferProvider should fall back to chain order)", stored.Provider, "r2")
	}
}

// --- Upload: malformed content ---

func TestUpload_NonSeekableContentShorterThanDeclaredSizeErrors(t *testing.T) {
	r2 := newFakeBackend("r2")
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 1000}}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	nonSeekable := bytes.NewBuffer([]byte("hi")) // 2 bytes, but Size claims 5
	_, err = chain.Upload(context.Background(), "a", ports.Picture{Content: nonSeekable, ContentType: "image/webp", Size: 5})
	if err == nil {
		t.Fatal("Upload: want error when declared Size exceeds actual content length, got nil")
	}
}

// --- Upload: full chain exhaustion with errors ---

func TestUpload_ThreeTierExhaustionJoinsAllPutErrors(t *testing.T) {
	r2, b2, sb := newFakeBackend("r2"), newFakeBackend("b2"), newFakeBackend("supabase")
	r2.putErr = errors.New("r2 down")
	b2.putErr = errors.New("b2 down")
	sb.putErr = errors.New("supabase down")
	chain, err := fallback.New(context.Background(), []fallback.Tier{
		{Backend: r2, QuotaBytes: 1000},
		{Backend: b2, QuotaBytes: 1000},
		{Backend: sb, QuotaBytes: 1000},
	}, newFakeLedger(), 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = chain.Upload(context.Background(), "a", pic("hello"))
	if !errors.Is(err, fallback.ErrStorageExhausted) {
		t.Fatalf("err = %v, want ErrStorageExhausted", err)
	}
	for _, msg := range []string{"r2 down", "b2 down", "supabase down"} {
		if !contains(err.Error(), msg) {
			t.Errorf("err = %q, want it to contain %q (every tier's failure should be joined in)", err.Error(), msg)
		}
	}
}

func contains(haystack, needle string) bool {
	return bytes.Contains([]byte(haystack), []byte(needle))
}

// --- Delete: error paths ---

func TestDelete_BackendErrorPropagatesAndSkipsForget(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 0}}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	if _, err := chain.Upload(ctx, "a", pic("hello")); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	r2.delErr = errors.New("network blip")
	if err := chain.Delete(ctx, "a"); err == nil {
		t.Fatal("Delete: want error when the backend's Del fails, got nil")
	}
	if _, ok, _ := ledger.Get(ctx, "a"); !ok {
		t.Error("ledger entry should still exist - Forget must not run after a failed Del")
	}
}

func TestDelete_PropagatesLedgerGetError(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 0}}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ledger.getErr = errors.New("db down")

	if err := chain.Delete(context.Background(), "a"); err == nil {
		t.Fatal("Delete: want error when the ledger's Get fails, got nil")
	}
}

func TestDelete_UnknownRecordedProviderErrors(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	// Simulate a ledger row pointing at a provider no longer in the chain
	// (e.g. removed from STORAGE_PROVIDERS after objects already landed there).
	if err := ledger.Record(context.Background(), ports.StorageObject{Key: "a", Provider: "ghost", Bytes: 5}); err != nil {
		t.Fatalf("seeding ledger: %v", err)
	}
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 0}}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := chain.Delete(context.Background(), "a"); err == nil {
		t.Fatal("Delete: want error for a key recorded on an unknown provider, got nil")
	}
}

func TestDelete_ForgetFailureIsSwallowedButUsageNotDecremented(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 6}}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	if _, err := chain.Upload(ctx, "a", pic("hello")); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	ledger.forgetErr = errors.New("db down")
	if err := chain.Delete(ctx, "a"); err != nil {
		t.Fatalf("Delete: %v (a Forget failure after a successful Del must not fail the call)", err)
	}
	if r2.has("a") {
		t.Error("the object should be gone from the backend regardless of the ledger's Forget failing")
	}

	// The in-memory usage counter was not decremented (Delete returns before
	// that point when Forget errors), so a same-size upload should still be
	// rejected as if "hello" were still counted against the 6-byte quota.
	if _, err := chain.Upload(ctx, "b", pic("hi")); err == nil {
		t.Error("Upload: want ErrStorageExhausted - usage should still reflect the undeleted-in-ledger bytes")
	}
}

// --- PresignGetURL: error paths ---

func TestPresignGetURL_PropagatesLedgerGetError(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 0}}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ledger.getErr = errors.New("db down")

	if _, err := chain.PresignGetURL(context.Background(), "a"); err == nil {
		t.Fatal("PresignGetURL: want error when the ledger's Get fails, got nil")
	}
}

func TestPresignGetURL_UnknownRecordedProviderErrors(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	if err := ledger.Record(context.Background(), ports.StorageObject{Key: "a", Provider: "ghost", Bytes: 5}); err != nil {
		t.Fatalf("seeding ledger: %v", err)
	}
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 0}}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := chain.PresignGetURL(context.Background(), "a"); err == nil {
		t.Fatal("PresignGetURL: want error for a key recorded on an unknown provider, got nil")
	}
}

// --- Concurrency ---

// TestUpload_ConcurrentUploadsAreRaceFree hammers Upload from many goroutines
// at once (run with -race) and checks the in-memory usage counter ends up
// exactly matching the bytes actually accepted - the atomic counter must
// never lose an update even though many goroutines read-then-add it
// concurrently.
func TestUpload_ConcurrentUploadsAreRaceFree(t *testing.T) {
	r2 := newFakeBackend("r2")
	ledger := newFakeLedger()
	chain, err := fallback.New(context.Background(), []fallback.Tier{{Backend: r2, QuotaBytes: 0}}, ledger, 95)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			if _, err := chain.Upload(ctx, key, pic("hello")); err != nil {
				t.Errorf("Upload(%s): %v", key, err)
			}
		}(i)
	}
	wg.Wait()

	usages, err := ledger.Usage(ctx)
	if err != nil {
		t.Fatalf("Usage: %v", err)
	}
	var total int64
	for _, u := range usages {
		total += u.Bytes
	}
	if want := int64(n * len("hello")); total != want {
		t.Errorf("ledger total bytes = %d, want %d (every concurrent upload should be recorded)", total, want)
	}
}
