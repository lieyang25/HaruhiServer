package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	// HTTPAddr is the listening address, for example ":8080".
	HTTPAddr string
	// LogLevel controls the minimum level written by slog.
	LogLevel slog.Level
}

// Default returns baseline config used when env vars are absent.
func Default() Config {
	return Config{
		HTTPAddr: ":8080",
		LogLevel: slog.LevelInfo,
	}
}

// Load builds config from defaults and optional env overrides.
func Load() (Config, error) {
	cfg := Default()

	if v, ok := lookupEnv("HS_HTTP_ADDR"); ok {
		cfg.HTTPAddr = v
	}

	if v, ok := lookupEnv("HS_LOG_LEVEL"); ok {
		level, err := parseLogLevel(v)
		if err != nil {
			return Config{}, fmt.Errorf("parse HS_LOG_LEVEL: %w", err)
		}

		cfg.LogLevel = level
	}

	return cfg, nil
}

// lookupEnv reads a non-empty, trimmed env value.
func lookupEnv(key string) (string, bool) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}

	v = strings.TrimSpace(v)

	if v == "" {
		return "", false
	}

	return v, true
}

// parseLogLevel maps env text into slog levels.
func parseLogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid level %q", s)
	}
}
