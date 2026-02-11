package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadPathSize(t *testing.T) {
	srv := &Server{DefaultHash: "sha256"}
	ts := httptest.NewServer(http.HandlerFunc(srv.Download))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/download/256.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != 256 {
		t.Errorf("len(body) = %d, want 256", len(body))
	}
	// Trailers are set by the real server; with httptest.Server we get a real response
	if resp.Trailer != nil {
		if h := resp.Trailer.Get("X-Content-Hash"); h != "" {
			sum := sha256.Sum256(body)
			if h != hex.EncodeToString(sum[:]) {
				t.Errorf("X-Content-Hash = %s, want %s", h, hex.EncodeToString(sum[:]))
			}
		}
	}
}

func TestDownloadQuerySize(t *testing.T) {
	srv := &Server{DefaultHash: "sha256"}
	ts := httptest.NewServer(http.HandlerFunc(srv.Download))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/download?size=100")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) != 100 {
		t.Errorf("len(body) = %d, want 100", len(body))
	}
}

func TestDownloadPathWinsOverQuery(t *testing.T) {
	srv := &Server{}
	ts := httptest.NewServer(http.HandlerFunc(srv.Download))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/download/50.bin?size=99999")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if len(body) != 50 {
		t.Errorf("path should win: len(body) = %d, want 50", len(body))
	}
}

func TestDownloadMaxSize(t *testing.T) {
	srv := &Server{MaxDownloadBytes: 100}
	ts := httptest.NewServer(http.HandlerFunc(srv.Download))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/download?size=200")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", resp.StatusCode)
	}
}

func TestUploadHash(t *testing.T) {
	srv := &Server{DefaultHash: "sha256"}
	ts := httptest.NewServer(http.HandlerFunc(srv.Upload))
	defer ts.Close()

	body := []byte("hello")
	resp, err := http.Post(ts.URL+"/upload", "application/octet-stream", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	sum := sha256.Sum256(body)
	wantHash := hex.EncodeToString(sum[:])
	if h := resp.Header.Get("X-Content-Hash"); h != wantHash {
		t.Errorf("X-Content-Hash = %s, want %s", h, wantHash)
	}
	respBody, _ := io.ReadAll(resp.Body)
	if string(respBody) != wantHash {
		t.Errorf("body = %s, want %s", respBody, wantHash)
	}
}

func TestUploadJSON(t *testing.T) {
	srv := &Server{DefaultHash: "sha256"}
	ts := httptest.NewServer(http.HandlerFunc(srv.Upload))
	defer ts.Close()

	body := []byte("test")
	resp, err := http.Post(ts.URL+"/upload?type=json", "application/octet-stream", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %s", ct)
	}
	respBody, _ := io.ReadAll(resp.Body)
	// Should be JSON like {"hash":"...","algorithm":"sha256"}
	if !bytes.Contains(respBody, []byte("hash")) || !bytes.Contains(respBody, []byte("algorithm")) {
		t.Errorf("body = %s", respBody)
	}
}
