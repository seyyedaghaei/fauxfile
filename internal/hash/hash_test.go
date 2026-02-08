package hash

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestSupported(t *testing.T) {
	for _, algo := range []string{"sha256", "sha512", "sha1", "md5"} {
		if !Supported(algo) {
			t.Errorf("Supported(%q) = false", algo)
		}
		if !Supported(strings.ToUpper(algo)) {
			t.Errorf("Supported(%q) = false", strings.ToUpper(algo))
		}
	}
	if Supported("sha3") {
		t.Error("Supported(sha3) should be false")
	}
}

func TestSum(t *testing.T) {
	data := []byte("hello")
	got := Sum("sha256", data)
	h := sha256.Sum256(data)
	want := hex.EncodeToString(h[:])
	if got != want {
		t.Errorf("Sum(sha256, %q) = %q, want %q", data, got, want)
	}
}

func TestSumReader(t *testing.T) {
	data := []byte("hello")
	got, err := SumReader("sha256", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(data)
	want := hex.EncodeToString(h[:])
	if got != want {
		t.Errorf("SumReader(sha256, ...) = %q, want %q", got, want)
	}
}
