package main

import (
	"log"
	"net/http"
	"time"
)

const (
	defaultAddr = ":8080"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allwoed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              defaultAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server starting on %s", defaultAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrBodyReadAfterClose {
		log.Fatalf("server failed: %v", err)
	}
}
