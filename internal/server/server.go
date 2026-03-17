package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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
	Version          string
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

func (s *Server) responseType(r *http.Request) string {
	if q := r.URL.Query().Get("type"); q != "" {
		switch strings.ToLower(q) {
		case "json", "text":
			return strings.ToLower(q)
		}
	}
	if s.DefaultRespType != "" {
		switch strings.ToLower(s.DefaultRespType) {
		case "json", "text":
			return strings.ToLower(s.DefaultRespType)
		}
	}
	return "text"
}

func (s *Server) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	algo := s.hashAlgo(r)
	hasher := hash.New(algo)
	if hasher == nil {
		http.Error(w, "unsupported hash algorithm", http.StatusBadRequest)
		return
	}

	body := r.Body
	if s.MaxUploadBytes > 0 {
		body = io.NopCloser(io.LimitReader(r.Body, s.MaxUploadBytes))
	}
	hexHash, err := hash.SumReader(algo, body)
	_ = r.Body.Close()
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Content-Hash", hexHash)
	w.Header().Set("X-Hash-Algorithm", algo)

	typ := s.responseType(r)
	switch typ {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		_ = enc.Encode(map[string]string{"hash": hexHash, "algorithm": algo})
	default:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(hexHash))
	}
}

// ServeVersion responds to GET /version with the server version (plain text).
func (s *Server) ServeVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	version := s.Version
	if version == "" {
		version = "dev"
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(version))
}
