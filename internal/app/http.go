package app

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

const maximumRequestBodySize = 1 << 20

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	body := http.MaxBytesReader(w, r.Body, maximumRequestBodySize)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(destination)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func newSPAHandler(root string) http.Handler {
	files := http.Dir(root)
	fileServer := http.FileServer(files)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath := filepath.Join(root, filepath.Clean(r.URL.Path))
		fileInfo, err := os.Stat(requestedPath)
		if err == nil && !fileInfo.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		index, err := fs.ReadFile(os.DirFS(root), "index.html")
		if err != nil {
			http.Error(w, "Frontend ainda não compilado. Execute npm run build em web/.", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(index)
	})
}
