package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"changes/internal/library"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	lib, err := library.Default()
	if err != nil {
		t.Fatalf("library.Default: %v", err)
	}
	return New(lib, "../../web").Handler()
}

func TestHandleList(t *testing.T) {
	h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/standards", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got []library.Summary
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) < 5 {
		t.Errorf("got %d standards, want >=5", len(got))
	}
}

func TestHandleStandardTranspose(t *testing.T) {
	h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/standards/blue-bossa?key=G&roman=1", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var r library.Rendered
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.Key != "G" {
		t.Errorf("key = %q, want G", r.Key)
	}
	first := r.Sections[0].Bars[0][0]
	if first.Symbol != "Gm7" || first.Roman != "i7" {
		t.Errorf("first chord = %q (%s), want Gm7 (i7)", first.Symbol, first.Roman)
	}
}

func TestHandleStandardNotFound(t *testing.T) {
	h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/standards/does-not-exist", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHandleStandardBadKey(t *testing.T) {
	h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/standards/blue-bossa?key=H", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleHealth(t *testing.T) {
	h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Status    string `json:"status"`
		Standards int    `json:"standards"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ok" || body.Standards < 5 {
		t.Errorf("health = %+v, want ok with >=5 standards", body)
	}
}

func TestServesIndex(t *testing.T) {
	h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<title>changes") {
		t.Errorf("index.html not served")
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("missing no-store cache header")
	}
}
