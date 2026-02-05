package main

import (
	"log"
	"net/http"
	"os"

	"github.com/seyyedaghaei/fauxfile/internal/parse"
	"github.com/seyyedaghaei/fauxfile/internal/server"
	"github.com/spf13/pflag"
)

func main() {
	addr := pflag.String("addr", ":8080", "listen address")
	maxSize := pflag.String("max-size", "", "maximum download size (e.g. 1g, 100m); empty = no limit")
	maxUpload := pflag.String("max-upload", "", "maximum upload body size (e.g. 100m); empty = no limit")
	defaultHash := pflag.String("hash", "sha256", "default hash algorithm (sha256, sha512, sha1, md5)")
	defaultResponseType := pflag.String("response-type", "text", "default upload response type (text, json)")
	pflag.Parse()

	if a := os.Getenv("FAUXFILE_ADDR"); a != "" && !pflag.Lookup("addr").Changed {
		*addr = a
	}

	var maxDownload, maxUploadBytes int64
	if *maxSize != "" {
		n, err := parse.Bytes(*maxSize)
		if err != nil {
			log.Fatalf("invalid -max-size: %v", err)
		}
		maxDownload = n
	}
	if *maxUpload != "" {
		n, err := parse.Bytes(*maxUpload)
		if err != nil {
			log.Fatalf("invalid -max-upload: %v", err)
		}
		maxUploadBytes = n
	}

	srv := &server.Server{
		MaxDownloadBytes: maxDownload,
		MaxUploadBytes:   maxUploadBytes,
		DefaultHash:     *defaultHash,
		DefaultRespType: *defaultResponseType,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/download", srv.Download)
	mux.HandleFunc("/download/", srv.Download)
	mux.HandleFunc("/upload", srv.Upload)

	log.Printf("fauxfile listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
