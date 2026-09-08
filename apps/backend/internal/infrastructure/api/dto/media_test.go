package dto

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func testBuilder() MediaURLBuilder {
	return NewMediaURLBuilder("/api/v1/media", "0123456789012345678901234567890123456789", time.Hour)
}

func TestMediaURLBuilder_Public_EmptyGroupID_ReturnsEmpty(t *testing.T) {
	b := testBuilder()
	if got := b.Public("", "main"); got != "" {
		t.Errorf("Public(\"\", ...) = %q, want empty", got)
	}
}

func TestMediaURLBuilder_Public_BuildsExpectedPath(t *testing.T) {
	b := testBuilder()
	got := b.Public("abc123", "card")
	want := "/api/v1/media/abc123/card.webp"
	if got != want {
		t.Errorf("Public = %q, want %q", got, want)
	}
}

// TestMediaURLBuilder_Private_QuantizedExp_IsDeterministicWithinWindow is the
// property the whole avatar-cacheability story depends on: two calls to
// Private within the same quantization window must return the
// byte-identical URL, or the browser cache/response ETag over a list
// containing many avatar URLs never actually hits.
func TestMediaURLBuilder_Private_QuantizedExp_IsDeterministicWithinWindow(t *testing.T) {
	b := testBuilder() // PrivateTTL = 1h, window = 30m
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	url1 := b.Private("group1", "main", base)
	url2 := b.Private("group1", "main", base.Add(10*time.Minute))
	if url1 != url2 {
		t.Errorf("Private URLs within the same window differ:\n  %s\n  %s", url1, url2)
	}
}

func TestMediaURLBuilder_Private_DifferentWindows_DifferentURLs(t *testing.T) {
	b := testBuilder()
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	url1 := b.Private("group1", "main", base)
	url2 := b.Private("group1", "main", base.Add(45*time.Minute))
	if url1 == url2 {
		t.Error("Private URLs across different windows must differ (each carries its own exp+sig)")
	}
}

func TestMediaURLBuilder_Private_EmptyGroupID_ReturnsEmpty(t *testing.T) {
	b := testBuilder()
	if got := b.Private("", "main", time.Now()); got != "" {
		t.Errorf("Private(\"\", ...) = %q, want empty", got)
	}
}

func TestMediaURLBuilder_VerifyPrivate_ValidSignature_Passes(t *testing.T) {
	b := testBuilder()
	now := time.Now()
	url := b.Private("group1", "main", now)

	exp, sig := parseTestPrivateURL(t, url)
	if !b.VerifyPrivate(exp, "group1", "main", sig, now) {
		t.Error("VerifyPrivate rejected a signature it just minted")
	}
}

func TestMediaURLBuilder_VerifyPrivate_Expired_Fails(t *testing.T) {
	b := testBuilder()
	now := time.Now()
	url := b.Private("group1", "main", now)
	exp, sig := parseTestPrivateURL(t, url)

	future := time.Unix(exp+1, 0)
	if b.VerifyPrivate(exp, "group1", "main", sig, future) {
		t.Error("VerifyPrivate accepted a signature past its exp")
	}
}

func TestMediaURLBuilder_VerifyPrivate_TamperedGroup_Fails(t *testing.T) {
	b := testBuilder()
	now := time.Now()
	url := b.Private("group1", "main", now)
	exp, sig := parseTestPrivateURL(t, url)

	if b.VerifyPrivate(exp, "group-other", "main", sig, now) {
		t.Error("VerifyPrivate accepted a signature minted for a different group")
	}
}

func TestMediaURLBuilder_VerifyPrivate_TamperedVariant_Fails(t *testing.T) {
	b := testBuilder()
	now := time.Now()
	url := b.Private("group1", "main", now)
	exp, sig := parseTestPrivateURL(t, url)

	if b.VerifyPrivate(exp, "group1", "thumb", sig, now) {
		t.Error("VerifyPrivate accepted a signature minted for a different variant")
	}
}

func TestMediaURLBuilder_VerifyPrivate_WrongSecret_Fails(t *testing.T) {
	b := testBuilder()
	other := NewMediaURLBuilder("/api/v1/media", "99999999999999999999999999999999999999", time.Hour)
	now := time.Now()
	url := b.Private("group1", "main", now)
	exp, sig := parseTestPrivateURL(t, url)

	if other.VerifyPrivate(exp, "group1", "main", sig, now) {
		t.Error("VerifyPrivate accepted a signature minted under a different secret")
	}
}

// parseTestPrivateURL extracts exp/sig from a
// "/api/v1/media/p/{exp}/{sig}/{group}/{variant}.webp" URL.
func parseTestPrivateURL(t *testing.T, url string) (exp int64, sig string) {
	t.Helper()
	parts := strings.Split(strings.TrimPrefix(url, "/api/v1/media/p/"), "/")
	if len(parts) < 2 {
		t.Fatalf("parsing private URL %q: unexpected shape", url)
	}
	parsed, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		t.Fatalf("parsing exp from %q: %v", url, err)
	}
	return parsed, parts[1]
}
