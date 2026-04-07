package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr string
	LogLevel slog.Level
}

func Default() Config {
	return Config{
		HTTPAddr: ":8080",
		LogLevel: slog.LevelInfo,
	}
}

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
