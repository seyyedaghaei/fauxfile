package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
	"strings"
)

var algos = map[string]func() hash.Hash{
	"sha256": func() hash.Hash { return sha256.New() },
	"sha512": func() hash.Hash { return sha512.New() },
	"sha1":   func() hash.Hash { return sha1.New() },
	"md5":    func() hash.Hash { return md5.New() },
}

// Supported returns true if algo is supported.
func Supported(algo string) bool {
	_, ok := algos[strings.ToLower(algo)]
	return ok
}

// New returns a hash.Hash for the given algorithm, or nil if unsupported.
func New(algo string) hash.Hash {
	f, ok := algos[strings.ToLower(algo)]
	if !ok {
		return nil
	}
	return f()
}

// Sum returns the hex-encoded hash of data using the given algorithm.
func Sum(algo string, data []byte) string {
	h := New(algo)
	if h == nil {
		return ""
	}
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// SumReader hashes the content read from r and returns hex-encoded hash.
func SumReader(algo string, r io.Reader) (string, error) {
	h := New(algo)
	if h == nil {
		return "", nil
	}
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
