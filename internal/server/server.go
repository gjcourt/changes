// Package server is the HTTP driving adapter: it serves the static web/
// frontend and a small JSON API over the standards library.
package server

import (
	"encoding/json"
	"net/http"
	"path"

	"changes/internal/library"
)

// Server wires the library to HTTP handlers.
type Server struct {
	lib    *library.Library
	webDir string
}

// New constructs a Server serving the given library and web directory.
func New(lib *library.Library, webDir string) *Server {
	return &Server{lib: lib, webDir: webDir}
}

// Handler returns the root http.Handler with routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/standards", s.handleList)
	mux.HandleFunc("GET /api/standards/{id}", s.handleStandard)
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.Handle("GET /", s.static())
	return withNoCache(mux)
}

// handleHealth is a liveness/readiness probe: 200 once the corpus is loaded.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"standards": len(s.lib.List()),
	})
}

// handleList returns the corpus metadata, sorted by title.
func (s *Server) handleList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.lib.List())
}

// handleStandard returns one standard transposed to ?key= (default original),
// with Roman-numeral analysis when ?roman=1|true.
func (s *Server) handleStandard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	key := r.URL.Query().Get("key")
	roman := isTrue(r.URL.Query().Get("roman"))

	rendered, err := s.lib.Render(id, key, roman)
	if err != nil {
		// Unknown id is a 404; a bad ?key= is a 400.
		if _, ok := s.lib.Get(id); !ok {
			http.Error(w, "standard not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, rendered)
}

// static serves files from webDir, falling back to index.html for "/".
func (s *Server) static() http.Handler {
	fileServer := http.FileServer(http.Dir(s.webDir))
	index := path.Join(s.webDir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path.Clean(r.URL.Path) == "/" {
			http.ServeFile(w, r, index)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func isTrue(s string) bool {
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

func withNoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
