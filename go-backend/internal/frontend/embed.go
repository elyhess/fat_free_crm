package frontend

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var distFS embed.FS

// Handler returns an http.Handler that serves the React SPA.
// Static files are served from the embedded dist directory.
// Non-file paths (SPA routes) fall back to index.html.
func Handler() http.Handler {
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("frontend: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(subFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// API and health routes are not handled here
		if strings.HasPrefix(path, "/api/") || path == "/health" {
			http.NotFound(w, r)
			return
		}

		// Try to serve the exact file
		if path != "/" {
			clean := strings.TrimPrefix(path, "/")
			if f, err := subFS.Open(clean); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// SPA fallback: serve index.html for all other paths
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
