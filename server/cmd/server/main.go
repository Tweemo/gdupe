// Command server runs the media-merge HTTP API.
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/timl/media-merge/server/internal/api"
	"github.com/timl/media-merge/server/internal/session"
)

func main() {
	addr := envOr("MEDIA_MERGE_ADDR", ":8080")

	root := envOr("MEDIA_MERGE_WORKDIR", filepath.Join(os.TempDir(), "media-merge"))
	if err := os.MkdirAll(root, 0o755); err != nil {
		log.Fatalf("could not create work dir %q: %v", root, err)
	}

	store := session.NewStore(root)
	srv := api.New(store)
	srv.StaticDir = os.Getenv("MEDIA_MERGE_STATIC_DIR")

	log.Printf("media-merge server listening on %s (workdir %s)", addr, root)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
