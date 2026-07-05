// Package api exposes the HTTP endpoints that drive the media-merge workflow:
// create a session, upload files, analyse for duplicates and similar groups,
// fetch thumbnails, then export and download the merged archive.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/timl/media-merge/server/internal/session"
)

// Server holds the dependencies shared by all handlers.
type Server struct {
	store *session.Store

	// StaticDir, when non-empty, is a directory of static frontend assets
	// served for any non-/api request. Left empty (e.g. in tests, or local
	// `go run`) the server is API-only.
	StaticDir string
}

// New constructs a Server backed by the given session store.
func New(store *session.Store) *Server {
	return &Server{store: store}
}

// Handler returns the HTTP router for the API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/sessions", s.handleCreateSession)
	mux.HandleFunc("POST /api/sessions/{id}/upload", s.handleUpload)
	mux.HandleFunc("POST /api/sessions/{id}/analyze", s.handleAnalyze)
	mux.HandleFunc("GET /api/sessions/{id}/thumbnail/{fileId}", s.handleThumbnail)
	mux.HandleFunc("POST /api/sessions/{id}/export", s.handleExport)
	mux.HandleFunc("GET /api/sessions/{id}/download", s.handleDownload)
	mux.HandleFunc("DELETE /api/sessions/{id}", s.handleDelete)

	// Serve the static frontend as the catch-all. The specific "/api/..."
	// patterns above take precedence over "GET /" in Go's ServeMux.
	if s.StaticDir != "" {
		mux.Handle("GET /", http.FileServer(http.Dir(s.StaticDir)))
	}

	return withCORS(mux)
}

// session looks up the session named in the {id} path value, writing a 404 if
// it is missing.
func (s *Server) session(w http.ResponseWriter, r *http.Request) (*session.Session, bool) {
	sess, ok := s.store.Get(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "session not found")
		return nil, false
	}
	return sess, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// withCORS allows the Next.js dev server (a different origin) to call the API.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
