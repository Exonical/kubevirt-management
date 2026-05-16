// Package static serves the embedded React SPA build.
//
// During development, the web/dist directory is typically empty until the
// frontend is built; we ship a minimal placeholder so the binary still
// compiles. Production builds copy the real assets into web/dist via the
// Dockerfile.
package static

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Handler returns an http.Handler that serves the SPA. Unknown paths fall
// through to index.html so client-side routing works.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean(r.URL.Path)
		if clean == "/" {
			clean = "/index.html"
		}
		if strings.HasPrefix(clean, "/api/") {
			http.NotFound(w, r)
			return
		}
		// If the requested file exists, serve it; otherwise serve index.html.
		f, err := sub.Open(strings.TrimPrefix(clean, "/"))
		if err != nil {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/index.html"
			fileServer.ServeHTTP(w, r2)
			return
		}
		_ = f.Close()
		fileServer.ServeHTTP(w, r)
	})
}
