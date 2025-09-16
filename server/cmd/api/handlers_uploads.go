package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var uploadDir = "/app/server/uploads" // absolute path inside the api container

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "bad multipart", http.StatusBadRequest)
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file missing", http.StatusBadRequest)
		return
	}
	defer file.Close()

	_ = os.MkdirAll(uploadDir, 0o755)

	ext := filepath.Ext(hdr.Filename)
	if ext == "" {
		ext = ".bin"
	}
	name := time.Now().UTC().Format("20060102-150405.000000000") + ext
	dstPath := filepath.Join(uploadDir, name)

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "cannot save", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err = io.Copy(dst, file); err != nil {
		http.Error(w, "write error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"url": "/uploads/" + name, // path we serve below
	})
}
