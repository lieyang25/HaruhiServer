package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"HaruhiServer/internal/config"
)

const appName = "haruhiserver"

func main() {
	cfg, err := config.Load()

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load config failded: %v\n", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	server := newHTTPServer(cfg.HTTPAddr, logger)

	logger.Info(
		"server starting",
		"app", appName,
		"addr", cfg.HTTPAddr,
		"log_level", cfg.LogLevel.String(),
	)

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func newLogger(level slog.Level) *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}

func newHTTPServer(addr string, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	registerRoutes(mux, logger)

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func registerRoutes(mux *http.ServeMux, logger *slog.Logger) {
	mux.HandleFunc("/healthz", healthzHandler(logger))
}

func healthzHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info(
			"http request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}
