package server

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/seyyedaghaei/fauxfile/internal/hash"
	"github.com/seyyedaghaei/fauxfile/internal/parse"
)

// trailerWriter is satisfied by *http.response when using chunked encoding.
type trailerWriter interface {
	http.ResponseWriter
	Trailer() http.Header
}

// Server holds config for download/upload handlers.
type Server struct {
	MaxDownloadBytes int64
	MaxUploadBytes   int64
	DefaultHash      string
	DefaultRespType  string
}

func (s *Server) hashAlgo(r *http.Request) string {
	if q := r.URL.Query().Get("hash"); q != "" && hash.Supported(q) {
		return strings.ToLower(q)
	}
	if h := r.Header.Get("X-Hash-Algorithm"); h != "" && hash.Supported(h) {
		return strings.ToLower(h)
	}
	if s.DefaultHash != "" && hash.Supported(s.DefaultHash) {
		return strings.ToLower(s.DefaultHash)
	}
	return "sha256"
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

	algo := s.hashAlgo(r)
	hasher := hash.New(algo)
	if hasher == nil {
		http.Error(w, "unsupported hash algorithm", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Trailer", "X-Content-Hash, X-Hash-Algorithm")
	// Omit Content-Length so we use chunked encoding and can send trailers
	w.WriteHeader(http.StatusOK)

	multi := io.MultiWriter(hasher, w)
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
		nw, err := multi.Write(buf[:n])
		written += int64(nw)
		if err != nil || nw != n {
			return
		}
	}

	if tw, ok := w.(trailerWriter); ok {
		tw.Trailer().Set("X-Content-Hash", hex.EncodeToString(hasher.Sum(nil)))
		tw.Trailer().Set("X-Hash-Algorithm", algo)
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
