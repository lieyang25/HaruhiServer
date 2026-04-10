package transporthttp

import (
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	logger     *slog.Logger
	systemInfo SystemInfo
}

type SystemInfo struct {
	Name      string
	Version   string
	StartedAt time.Time
}

func NewHandler(logger *slog.Logger) http.Handler {
	return NewHandlerWithSystemInfo(logger, SystemInfo{})
}

func NewHandlerWithSystemInfo(logger *slog.Logger, info SystemInfo) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	info = normalizeSystemInfo(info)

	h := &Handler{
		logger:     logger,
		systemInfo: info,
	}

	mux := http.NewServeMux()
	h.registerRoutes(mux)

	return Chain(
		mux,
		RequestIDMiddleware(),
		LoggingMiddleware(logger),
		RecoverMiddleware(logger),
		CORSMiddleware(),
	)
}

func normalizeSystemInfo(info SystemInfo) SystemInfo {
	if info.Name == "" {
		info.Name = "haruhiserver"
	}
	if info.Version == "" {
		info.Version = "dev"
	}
	if info.StartedAt.IsZero() {
		info.StartedAt = time.Now().UTC()
	}
	return info
}
