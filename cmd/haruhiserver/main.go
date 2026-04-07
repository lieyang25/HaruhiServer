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

// main wires config, logging, routes, and starts the HTTP server.
func main() {
	// Load runtime config from environment; stop early on invalid values.
	cfg, err := config.Load()

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load config failded: %v\n", err)
		os.Exit(1)
	}

	// Set process default logger so internal packages can use slog directly.
	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	// Build server with routes and basic timeouts.
	server := newHTTPServer(cfg.HTTPAddr, logger)

	logger.Info(
		"server starting",
		"app", appName,
		"addr", cfg.HTTPAddr,
		"log_level", cfg.LogLevel.String(),
	)

	// ErrServerClosed is expected when server is shut down intentionally.
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

// newLogger creates a text logger with the configured minimum level.
func newLogger(level slog.Level) *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}

// newHTTPServer creates the HTTP server and mounts all routes.
func newHTTPServer(addr string, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	registerRoutes(mux, logger)

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

// registerRoutes centralizes endpoint registration.
func registerRoutes(mux *http.ServeMux, logger *slog.Logger) {
	mux.HandleFunc("/healthz", healthzHandler(logger))
}

// healthzHandler serves a basic liveness endpoint.
func healthzHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info(
			"http request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		// Keep the endpoint contract strict: only GET is allowed.
		if r.Method != http.MethodGet {
			response.Error(w, apperr.New(
				apperr.CodeMethodNotAllowed,
				"method not allowed",
			))
			return
		}

		// Return unified JSON envelope for successful health checks.
		response.OK(w, map[string]any{
			"status": "ok",
		})
	}
}
