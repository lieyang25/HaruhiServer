package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"HaruhiServer/internal/config"
	transporthttp "HaruhiServer/internal/transport/http"
)

const appName = "haruhiserver"
const appVersion = "dev"

// main wires config, logging, routes, and starts the HTTP server.
func main() {
	startedAt := time.Now().UTC()

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
	server := newHTTPServer(cfg.HTTPAddr, logger, transporthttp.SystemInfo{
		Name:      appName,
		Version:   appVersion,
		StartedAt: startedAt,
	})

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
func newHTTPServer(addr string, logger *slog.Logger, info transporthttp.SystemInfo) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           transporthttp.NewHandlerWithSystemInfo(logger, info),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
