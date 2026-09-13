package workerproxy_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/storage/workerproxy"
)

const testSecret = "test-shared-secret"

func newBackend(t *testing.T, url string) *workerproxy.Backend {
	t.Helper()
	b, err := workerproxy.New(workerproxy.Config{
		Name: "r2", BaseURL: url, Secret: testSecret, PresignTTL: 15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return b
}

func TestName(t *testing.T) {
	b := newBackend(t, "https://example.workers.dev")
	if got := b.Name(); got != "r2" {
		t.Errorf("Name() = %q, want %q", got, "r2")
	}
}

func TestPut_SendsAuthMethodPathBodyAndContentType(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotContentType string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	b := newBackend(t, srv.URL)
	content := []byte("hello world")
	if err := b.Put(context.Background(), "stands/1/main.webp", bytes.NewReader(content), "image/webp", int64(len(content))); err != nil {
		t.Fatalf("Put: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if gotPath != "/objects/stands/1/main.webp" {
		t.Errorf("path = %q, want %q", gotPath, "/objects/stands/1/main.webp")
	}
	if gotAuth != "Bearer "+testSecret {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer "+testSecret)
	}
	if gotContentType != "image/webp" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "image/webp")
	}
	if string(gotBody) != "hello world" {
		t.Errorf("body = %q, want %q", gotBody, "hello world")
	}
}

func TestPut_NonOKStatusIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()

	b := newBackend(t, srv.URL)
	err := b.Put(context.Background(), "a", bytes.NewReader([]byte("x")), "text/plain", 1)
	if err == nil {
		t.Fatal("Put: want error, got nil")
	}
}

func TestGet_ReturnsBodyAndContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/objects/a/b" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/objects/a/b")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+testSecret {
			t.Errorf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pngbytes"))
	}))
	defer srv.Close()

	b := newBackend(t, srv.URL)
	rc, info, err := b.Get(context.Background(), "a/b")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	if string(data) != "pngbytes" {
		t.Errorf("body = %q, want %q", data, "pngbytes")
	}
	if info.ContentType != "image/png" {
		t.Errorf("ContentType = %q, want %q", info.ContentType, "image/png")
	}
	if info.Size != int64(len("pngbytes")) {
		t.Errorf("Size = %d, want %d", info.Size, len("pngbytes"))
	}
}

func TestGet_404MapsToErrObjectNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	b := newBackend(t, srv.URL)
	_, _, err := b.Get(context.Background(), "missing")
	if !errors.Is(err, ports.ErrObjectNotFound) {
		t.Fatalf("err = %v, want ErrObjectNotFound", err)
	}
}

func TestGet_ServerErrorIsGenericError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	b := newBackend(t, srv.URL)
	_, _, err := b.Get(context.Background(), "a")
	if err == nil {
		t.Fatal("Get: want error, got nil")
	}
	if errors.Is(err, ports.ErrObjectNotFound) {
		t.Fatal("a 500 should not be reported as ErrObjectNotFound")
	}
}

func TestDel_SendsMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	b := newBackend(t, srv.URL)
	if err := b.Del(context.Background(), "x/y.webp"); err != nil {
		t.Fatalf("Del: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/objects/x/y.webp" {
		t.Errorf("path = %q, want %q", gotPath, "/objects/x/y.webp")
	}
}

func TestDel_MissingKeyIsNotAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	b := newBackend(t, srv.URL)
	if err := b.Del(context.Background(), "missing"); err != nil {
		t.Fatalf("Del of a missing key should not error, got: %v", err)
	}
}

func TestWalk_PagesThroughCursor(t *testing.T) {
	pages := []string{
		`{"keys":[{"key":"a","size":1},{"key":"b","size":2}],"cursor":"page2"}`,
		`{"keys":[{"key":"c","size":3}],"cursor":""}`,
	}
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/objects" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/objects")
		}
		cursor := r.URL.Query().Get("cursor")
		if call == 0 && cursor != "" {
			t.Errorf("first call cursor = %q, want empty", cursor)
		}
		if call == 1 && cursor != "page2" {
			t.Errorf("second call cursor = %q, want %q", cursor, "page2")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pages[call]))
		call++
	}))
	defer srv.Close()

	b := newBackend(t, srv.URL)
	var got []string
	var gotSizes []int64
	err := b.Walk(context.Background(), func(key string, size int64) error {
		got = append(got, key)
		gotSizes = append(gotSizes, size)
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if call != 2 {
		t.Fatalf("expected 2 requests (one per page), got %d", call)
	}
	if fmt.Sprint(got) != fmt.Sprint([]string{"a", "b", "c"}) {
		t.Errorf("keys = %v, want [a b c]", got)
	}
	if fmt.Sprint(gotSizes) != fmt.Sprint([]int64{1, 2, 3}) {
		t.Errorf("sizes = %v, want [1 2 3]", gotSizes)
	}
}

func TestWalk_PropagatesCallbackError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"keys":[{"key":"a","size":1}],"cursor":""}`))
	}))
	defer srv.Close()

	b := newBackend(t, srv.URL)
	stop := errors.New("stop walking")
	err := b.Walk(context.Background(), func(key string, size int64) error { return stop })
	if !errors.Is(err, stop) {
		t.Fatalf("err = %v, want %v", err, stop)
	}
}

func TestPresignGet_URLValidatesAgainstWorkerSideSignature(t *testing.T) {
	b := newBackend(t, "https://example.workers.dev")
	presignedURL, err := b.PresignGet(context.Background(), "stands/1/main.webp")
	if err != nil {
		t.Fatalf("PresignGet: %v", err)
	}

	// Parse out exp/sig the way the worker would, and recompute the
	// signature independently to prove the client and worker sides agree
	// on the scheme (HMAC-SHA256 of "key|exp" with the shared secret).
	u, err := url.Parse(presignedURL)
	if err != nil {
		t.Fatalf("parsing presigned url %q: %v", presignedURL, err)
	}
	if u.Path != "/objects/stands/1/main.webp" {
		t.Errorf("path = %q, want %q", u.Path, "/objects/stands/1/main.webp")
	}
	expStr := u.Query().Get("exp")
	sig := u.Query().Get("sig")
	if expStr == "" || sig == "" {
		t.Fatalf("expected exp and sig query params, got url %q", presignedURL)
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		t.Fatalf("parsing exp: %v", err)
	}
	if time.Until(time.Unix(exp, 0)) > 16*time.Minute || time.Until(time.Unix(exp, 0)) < 14*time.Minute {
		t.Errorf("exp = %v, want ~15 minutes from now (PresignTTL)", time.Unix(exp, 0))
	}

	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte(fmt.Sprintf("stands/1/main.webp|%d", exp)))
	want := hex.EncodeToString(mac.Sum(nil))
	if sig != want {
		t.Errorf("sig = %q, want %q (worker must recompute the same way)", sig, want)
	}
}

func TestNew_RejectsMissingBaseURLOrSecret(t *testing.T) {
	if _, err := workerproxy.New(workerproxy.Config{Name: "r2", BaseURL: "", Secret: "s"}); err == nil {
		t.Error("New: want error for empty BaseURL, got nil")
	}
	if _, err := workerproxy.New(workerproxy.Config{Name: "r2", BaseURL: "https://x", Secret: ""}); err == nil {
		t.Error("New: want error for empty Secret, got nil")
	}
}
