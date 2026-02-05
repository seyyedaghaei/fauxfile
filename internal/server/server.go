package server

import (
	"crypto/rand"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/seyyedaghaei/fauxfile/internal/parse"
)

// Server holds config for download/upload handlers.
type Server struct {
	MaxDownloadBytes int64
	MaxUploadBytes   int64
	DefaultHash      string
	DefaultRespType  string
}

// downloadSize returns size in bytes from request. Path wins over query.
// Path: /download/10mb.bin or /download/10m.bin (case insensitive).
// Query: ?size=10mb or ?size=1024.
func (s *Server) downloadSize(r *http.Request) (int64, error) {
	base := path.Base(r.URL.Path)
	lower := strings.ToLower(base)
	if strings.HasSuffix(lower, ".bin") {
		sizeStr := strings.TrimSuffix(lower, ".bin")
		return parse.Bytes(sizeStr)
	}
	if q := r.URL.Query().Get("size"); q != "" {
		return parse.Bytes(q)
	}
	return 0, nil
}

func (s *Server) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	size, err := s.downloadSize(r)
	if err != nil {
		http.Error(w, "invalid size: "+err.Error(), http.StatusBadRequest)
		return
	}
	if size == 0 {
		http.Error(w, "missing size (path e.g. /download/10mb.bin or query ?size=10mb)", http.StatusBadRequest)
		return
	}
	if s.MaxDownloadBytes > 0 && size > s.MaxDownloadBytes {
		http.Error(w, "size exceeds max allowed", http.StatusRequestEntityTooLarge)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.WriteHeader(http.StatusOK)

	const chunk = 64 * 1024
	written := int64(0)
	buf := make([]byte, chunk)
	for written < size {
		n := chunk
		if size-written < int64(n) {
			n = int(size - written)
		}
		if _, err := rand.Read(buf[:n]); err != nil {
			return
		}
		nw, err := w.Write(buf[:n])
		written += int64(nw)
		if err != nil || nw != n {
			return
		}
	}
}

func (s *Server) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, _ = io.Copy(io.Discard, r.Body)
	r.Body.Close()
	w.WriteHeader(http.StatusNotImplemented)
}
