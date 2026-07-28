package endpoints_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetStands_ETag_MatchingIfNoneMatchReturns304(t *testing.T) {
	h := newTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Star Platinum"))

	first := doRequest(t, h, http.MethodGet, "/api/v1/stands", nil)
	if first.Code != http.StatusOK {
		t.Fatalf("first GET: status = %d, want %d", first.Code, http.StatusOK)
	}
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("first GET: no ETag header set")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stands", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	req.Header.Set("If-None-Match", etag)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotModified {
		t.Fatalf("second GET with If-None-Match: status = %d, want %d", rec.Code, http.StatusNotModified)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("304 response body = %q, want empty", rec.Body.String())
	}
	if got := rec.Header().Get("ETag"); got != etag {
		t.Errorf("304 ETag = %q, want %q", got, etag)
	}
}

func TestGetStands_ETag_ChangesAfterWrite(t *testing.T) {
	h := newTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Silver Chariot"))

	first := doRequest(t, h, http.MethodGet, "/api/v1/stands", nil)
	firstETag := first.Header().Get("ETag")
	if firstETag == "" {
		t.Fatal("first GET: no ETag header set")
	}

	doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Star Platinum"))

	second := doRequest(t, h, http.MethodGet, "/api/v1/stands", nil)
	secondETag := second.Header().Get("ETag")
	if secondETag == "" {
		t.Fatal("second GET: no ETag header set")
	}
	if secondETag == firstETag {
		t.Error("ETag unchanged after a write that altered the collection")
	}
}

func TestGetStand_CacheControl_Present(t *testing.T) {
	h := newTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("The World"))
	id := mustExtractID(t, createRec)

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stands/"+id, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	cc := rec.Header().Get("Cache-Control")
	if cc == "" {
		t.Fatal("Cache-Control header not set")
	}
	if !strings.Contains(cc, "private") {
		t.Errorf("Cache-Control = %q, want it to contain %q", cc, "private")
	}
	if got := rec.Header().Get("Vary"); got != "Authorization" {
		t.Errorf("Vary = %q, want %q", got, "Authorization")
	}
}

// TestWriteEndpoints_NoETag proves the cache-headers middleware is only
// mounted on the two GET routes: a write's response must carry no ETag.
func TestWriteEndpoints_NoETag(t *testing.T) {
	h := newTestServer()
	rec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Killer Queen"))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get("ETag"); got != "" {
		t.Errorf("POST response ETag = %q, want none", got)
	}
}

func mustExtractID(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	id, ok := created["id"].(string)
	if !ok {
		t.Fatalf("response has no string id: %v", created)
	}
	return id
}
