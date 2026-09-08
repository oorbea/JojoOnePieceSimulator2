package endpoints_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
)

// mediaFakeStorage is a minimal in-memory ports.IPictureStorage - only
// Download is exercised by the media proxy; the rest just need to compile.
type mediaFakeStorage struct {
	objects map[string][]byte
}

func newMediaFakeStorage() *mediaFakeStorage {
	return &mediaFakeStorage{objects: make(map[string][]byte)}
}

func (s *mediaFakeStorage) Upload(context.Context, string, ports.Picture) (ports.StoredPicture, error) {
	return ports.StoredPicture{}, nil
}

func (s *mediaFakeStorage) PresignGetURL(_ context.Context, key string) (string, error) {
	return "https://r2.test/" + key, nil
}

func (s *mediaFakeStorage) Delete(context.Context, string) error { return nil }

func (s *mediaFakeStorage) Download(_ context.Context, key string) (io.ReadCloser, ports.ObjectInfo, error) {
	data, ok := s.objects[key]
	if !ok {
		return nil, ports.ObjectInfo{}, ports.ErrObjectNotFound
	}
	return io.NopCloser(byteReader(data)), ports.ObjectInfo{ContentType: "image/webp", Size: int64(len(data))}, nil
}

var _ ports.IPictureStorage = (*mediaFakeStorage)(nil)

// mediaFakeRepo is a minimal in-memory ports.IMediaRepository.
type mediaFakeRepo struct {
	byGroup map[string][]ports.MediaObject
}

func newMediaFakeRepo() *mediaFakeRepo {
	return &mediaFakeRepo{byGroup: make(map[string][]ports.MediaObject)}
}

func (r *mediaFakeRepo) PutMediaObjects(_ context.Context, objects []ports.MediaObject) error {
	for _, o := range objects {
		r.byGroup[o.GroupID] = append(r.byGroup[o.GroupID], o)
	}
	return nil
}

func (r *mediaFakeRepo) GetMediaObjectsByGroup(_ context.Context, groupID string) ([]ports.MediaObject, error) {
	return r.byGroup[groupID], nil
}

var _ ports.IMediaRepository = (*mediaFakeRepo)(nil)

func byteReader(b []byte) *bytesReaderCloser { return &bytesReaderCloser{data: b} }

type bytesReaderCloser struct {
	data []byte
	pos  int
}

func (b *bytesReaderCloser) Read(p []byte) (int, error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

const testGroupID = "0123456789abcdef0123456789abcdef"

func newMediaTestServer(t *testing.T) (*httptest.Server, *mediaFakeStorage, *mediaFakeRepo, dto.MediaURLBuilder) {
	t.Helper()
	storage := newMediaFakeStorage()
	repo := newMediaFakeRepo()
	urls := dto.NewMediaURLBuilder("/api/v1/media", "01234567890123456789012345678901", time.Hour)
	mediaEndpoints := endpoints.NewMediaEndpoints(storage, repo, urls, endpoints.MediaConfig{Mode: "proxy"})

	authEndpoints := endpoints.NewAuthEndpoints(nil, endpoints.CookieConfig{})
	standEndpoints := endpoints.NewStandEndpoints(nil)
	r := chiTestRouter(authEndpoints, standEndpoints, mediaEndpoints)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv, storage, repo, urls
}

// chiTestRouter builds the smallest NewRouter call that exercises /media -
// every other endpoints arg is nil/bare, same convention as the rest of this
// package's *_test.go files.
func chiTestRouter(authEndpoints *endpoints.AuthEndpoints, standEndpoints *endpoints.StandEndpoints, mediaEndpoints *endpoints.MediaEndpoints) http.Handler {
	return endpoints.NewRouter(authEndpoints, standEndpoints, endpoints.NewDevilFruitEndpoints(nil), endpoints.NewUserEndpoints(nil),
		endpoints.NewEventsEndpoints(nil, fakeTokenIssuer{}, nil, context.Background()),
		endpoints.NewGameEndpoints(nil, nil, nil, nil, nil, nil, fakeTokenIssuer{}, nil, context.Background(), endpoints.GameWSConfig{}),
		endpoints.NewStageEndpoints(nil), mediaEndpoints, fakeTokenIssuer{}, endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{}, 0)
}

func TestMedia_Public_UnknownGroup_Returns404(t *testing.T) {
	srv, _, _, _ := newMediaTestServer(t)

	resp, err := http.Get(srv.URL + "/api/v1/media/" + testGroupID + "/main.webp")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestMedia_Public_InvalidGroupFormat_Returns400_BeforeTouchingStorage(t *testing.T) {
	srv, storage, _, _ := newMediaTestServer(t)

	resp, err := http.Get(srv.URL + "/api/v1/media/not-hex/main.webp")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	if len(storage.objects) != 0 {
		t.Error("storage should never have been touched for an invalid group id")
	}
}

func TestMedia_Public_InvalidVariant_Returns400(t *testing.T) {
	srv, _, _, _ := newMediaTestServer(t)

	resp, err := http.Get(srv.URL + "/api/v1/media/" + testGroupID + "/huge.webp")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestMedia_Public_KnownGroup_Serves200WithImmutableHeaders(t *testing.T) {
	srv, storage, repo, _ := newMediaTestServer(t)
	storage.objects["stands/x/main.webp"] = []byte("main-bytes")
	_ = repo.PutMediaObjects(context.Background(), []ports.MediaObject{
		{GroupID: testGroupID, Variant: "main", StorageKey: "stands/x/main.webp", ContentType: "image/webp", Bytes: 10, Scope: "public"},
	})

	resp, err := http.Get(srv.URL + "/api/v1/media/" + testGroupID + "/main.webp")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "main-bytes" {
		t.Errorf("body = %q, want %q", body, "main-bytes")
	}
	if got := resp.Header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control = %q", got)
	}
	if got := resp.Header.Get("Content-Type"); got != "image/webp" {
		t.Errorf("Content-Type = %q", got)
	}
	if resp.Header.Get("ETag") == "" {
		t.Error("ETag missing")
	}
}

func TestMedia_Public_CardFallsBackToThumbThenMain(t *testing.T) {
	srv, storage, repo, _ := newMediaTestServer(t)
	storage.objects["stands/x/thumb.webp"] = []byte("thumb-bytes")
	_ = repo.PutMediaObjects(context.Background(), []ports.MediaObject{
		{GroupID: testGroupID, Variant: "thumb", StorageKey: "stands/x/thumb.webp", ContentType: "image/webp", Bytes: 11, Scope: "public"},
	})

	// card requested, only thumb backfilled - must serve thumb instead of 404.
	resp, err := http.Get(srv.URL + "/api/v1/media/" + testGroupID + "/card.webp")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (fallback ladder)", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "thumb-bytes" {
		t.Errorf("body = %q, want thumb bytes", body)
	}
}

func TestMedia_Public_IfNoneMatch_Returns304NoBody(t *testing.T) {
	srv, storage, repo, _ := newMediaTestServer(t)
	storage.objects["stands/x/main.webp"] = []byte("main-bytes")
	_ = repo.PutMediaObjects(context.Background(), []ports.MediaObject{
		{GroupID: testGroupID, Variant: "main", StorageKey: "stands/x/main.webp", ContentType: "image/webp", Bytes: 10, Scope: "public"},
	})

	first, err := http.Get(srv.URL + "/api/v1/media/" + testGroupID + "/main.webp")
	if err != nil {
		t.Fatalf("first GET: %v", err)
	}
	etag := first.Header.Get("ETag")
	_ = first.Body.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/media/"+testGroupID+"/main.webp", nil)
	req.Header.Set("If-None-Match", etag)
	second, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("second GET: %v", err)
	}
	defer func() { _ = second.Body.Close() }()

	if second.StatusCode != http.StatusNotModified {
		t.Errorf("status = %d, want 304", second.StatusCode)
	}
	body, _ := io.ReadAll(second.Body)
	if len(body) != 0 {
		t.Errorf("304 body = %q, want empty", body)
	}
}

func TestMedia_Public_NeverRequiresAuthorization(t *testing.T) {
	srv, storage, repo, _ := newMediaTestServer(t)
	storage.objects["stands/x/main.webp"] = []byte("main-bytes")
	_ = repo.PutMediaObjects(context.Background(), []ports.MediaObject{
		{GroupID: testGroupID, Variant: "main", StorageKey: "stands/x/main.webp", ContentType: "image/webp", Bytes: 10, Scope: "public"},
	})

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/media/"+testGroupID+"/main.webp", nil)
	// Deliberately no Authorization header.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 without Authorization", resp.StatusCode)
	}
}

func TestMedia_Private_ValidSignature_Serves200(t *testing.T) {
	srv, storage, repo, urls := newMediaTestServer(t)
	storage.objects["users/x/main.webp"] = []byte("avatar-bytes")
	_ = repo.PutMediaObjects(context.Background(), []ports.MediaObject{
		{GroupID: testGroupID, Variant: "main", StorageKey: "users/x/main.webp", ContentType: "image/webp", Bytes: 12, Scope: "private"},
	})

	url := urls.Private(testGroupID, "main", time.Now())
	resp, err := http.Get(srv.URL + url)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if got := resp.Header.Get("Vary"); got != "" {
		t.Errorf("Vary = %q, want unset (authority lives in the URL, not a header)", got)
	}
	cc := resp.Header.Get("Cache-Control")
	if cc == "" || cc == "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control = %q, want a private/immutable directive", cc)
	}
}

func TestMedia_Private_ExpiredSignature_Returns401(t *testing.T) {
	srv, storage, repo, urls := newMediaTestServer(t)
	storage.objects["users/x/main.webp"] = []byte("avatar-bytes")
	_ = repo.PutMediaObjects(context.Background(), []ports.MediaObject{
		{GroupID: testGroupID, Variant: "main", StorageKey: "users/x/main.webp", ContentType: "image/webp", Bytes: 12, Scope: "private"},
	})

	past := time.Now().Add(-2 * time.Hour)
	url := urls.Private(testGroupID, "main", past)
	resp, err := http.Get(srv.URL + url)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestMedia_Private_TamperedSignature_Returns401(t *testing.T) {
	srv, storage, repo, urls := newMediaTestServer(t)
	storage.objects["users/x/main.webp"] = []byte("avatar-bytes")
	_ = repo.PutMediaObjects(context.Background(), []ports.MediaObject{
		{GroupID: testGroupID, Variant: "main", StorageKey: "users/x/main.webp", ContentType: "image/webp", Bytes: 12, Scope: "private"},
	})

	url := urls.Private(testGroupID, "main", time.Now())
	// url is ".../p/{exp}/{sig}/{group}/main.webp" - flip the last char of
	// sig specifically, not the variant/group segments after it.
	parts := strings.Split(url, "/")
	sigIdx := len(parts) - 3
	parts[sigIdx] = parts[sigIdx][:len(parts[sigIdx])-1] + "_"
	tampered := strings.Join(parts, "/")
	resp, err := http.Get(srv.URL + tampered)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestMedia_Public_HEAD_ReturnsHeadersNoBody(t *testing.T) {
	srv, storage, repo, _ := newMediaTestServer(t)
	storage.objects["stands/x/main.webp"] = []byte("main-bytes")
	_ = repo.PutMediaObjects(context.Background(), []ports.MediaObject{
		{GroupID: testGroupID, Variant: "main", StorageKey: "stands/x/main.webp", ContentType: "image/webp", Bytes: 10, Scope: "public"},
	})

	req, _ := http.NewRequest(http.MethodHead, srv.URL+"/api/v1/media/"+testGroupID+"/main.webp", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HEAD: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) != 0 {
		t.Errorf("HEAD body = %q, want empty", body)
	}
	if resp.Header.Get("Content-Length") == "" {
		t.Error("Content-Length missing on HEAD")
	}
}
