package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"HaruhiServer/internal/apperr"
	"HaruhiServer/internal/config"
	"HaruhiServer/internal/response"
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
			response.Error(w, apperr.New(
				apperr.CodeMethodNotAllowed,
				"method not allowed",
			))
			return
		}

		response.OK(w, map[string]any{
			"status": "ok",
		})
	}
}
