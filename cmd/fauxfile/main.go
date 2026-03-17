package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seyyedaghaei/fauxfile/internal/parse"
	"github.com/seyyedaghaei/fauxfile/internal/server"
	"github.com/spf13/pflag"
)

// Version is set at build time via -ldflags "-X main.Version=..."
var Version = "dev"

func main() {
	listen := pflag.StringP("listen", "l", ":8080", "listen address")
	maxSize := pflag.StringP("max-size", "s", "", "maximum download size (e.g. 1g, 100m); empty = no limit")
	maxUpload := pflag.StringP("max-upload", "u", "", "maximum upload body size (e.g. 100m); empty = no limit")
	defaultHash := pflag.StringP("hash", "H", "sha256", "default hash algorithm (sha256, sha512, sha1, md5)")
	defaultResponseType := pflag.StringP("response-type", "r", "text", "default upload response type (text, json)")
	tlsCert := pflag.StringP("tls-cert", "c", "", "path to TLS certificate file (enables HTTPS with --tls-key)")
	tlsKey := pflag.StringP("tls-key", "k", "", "path to TLS private key file (enables HTTPS with --tls-cert)")
	showVersion := pflag.BoolP("version", "v", false, "print version and exit")
	pflag.Parse()

	if *showVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	if a := os.Getenv("FAUXFILE_ADDR"); a != "" && !pflag.Lookup("listen").Changed {
		*listen = a
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
		Version:         Version,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/download", srv.Download)
	mux.HandleFunc("/download/", srv.Download)
	mux.HandleFunc("/upload", srv.Upload)
	mux.HandleFunc("/version", srv.ServeVersion)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Version != "" {
			w.Header().Set("X-Fauxfile-Version", Version)
		}
		mux.ServeHTTP(w, r)
	})

	if *tlsCert != "" || *tlsKey != "" {
		if *tlsCert == "" || *tlsKey == "" {
			log.Fatal("both --tls-cert and --tls-key must be set to enable HTTPS")
		}
	}

	httpSrv := &http.Server{Addr: *listen, Handler: handler}
	go func() {
		if *tlsCert != "" && *tlsKey != "" {
			log.Printf("fauxfile listening on %s (TLS)", *listen)
			if err := httpSrv.ListenAndServeTLS(*tlsCert, *tlsKey); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %v", err)
			}
		} else {
			log.Printf("fauxfile listening on %s", *listen)
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %v", err)
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Print("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Print("done")
}
