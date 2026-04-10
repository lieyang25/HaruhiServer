package config_test

import (
	"log/slog"
	"testing"

	"HaruhiServer/internal/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8080")
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
}

func TestLoad_OverridesAndTrim(t *testing.T) {
	t.Setenv("HS_HTTP_ADDR", "  :9090  ")
	t.Setenv("HS_LOG_LEVEL", "  DeBuG  ")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() err = %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":9090")
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelDebug)
	}
}

func TestLoad_EmptyEnvIgnored(t *testing.T) {
	t.Setenv("HS_HTTP_ADDR", "   ")
	t.Setenv("HS_LOG_LEVEL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() err = %v", err)
	}

	def := config.Default()
	if cfg.HTTPAddr != def.HTTPAddr || cfg.LogLevel != def.LogLevel {
		t.Fatalf("cfg = %#v, want default %#v", cfg, def)
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	t.Setenv("HS_LOG_LEVEL", "verbose")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() err = nil, want non-nil")
	}
}
