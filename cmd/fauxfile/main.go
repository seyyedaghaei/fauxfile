package main

import (
	"log"
	"net/http"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	addr := pflag.String("addr", ":8080", "listen address")
	maxSize := pflag.String("max-size", "", "maximum download size (e.g. 1g, 100m); 0 or empty = no limit")
	maxUpload := pflag.String("max-upload", "", "maximum upload body size (e.g. 100m); 0 or empty = no limit")
	defaultHash := pflag.String("hash", "sha256", "default hash algorithm (sha256, sha512, sha1, md5)")
	defaultResponseType := pflag.String("response-type", "text", "default upload response type (text, json)")
	pflag.Parse()

	if a := os.Getenv("FAUXFILE_ADDR"); a != "" && !pflag.Lookup("addr").Changed {
		*addr = a
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/download", downloadHandler)
	mux.HandleFunc("/download/", downloadHandler)
	mux.HandleFunc("/upload", uploadHandler)

	log.Printf("fauxfile listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("listen: %v", err)
	}
	_ = maxSize
	_ = maxUpload
	_ = defaultHash
	_ = defaultResponseType
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNotImplemented)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNotImplemented)
}
