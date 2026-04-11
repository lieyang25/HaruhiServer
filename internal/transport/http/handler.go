package transporthttp

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"HaruhiServer/internal/repository/memory"
	"HaruhiServer/internal/service"
)

type Handler struct {
	logger     *slog.Logger
	systemInfo SystemInfo
	services   *service.Services
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
	svcs, err := service.NewServices(memory.NewRepositories(), service.RandomIDGenerator{}, time.Now)
	if err != nil {
		panic(fmt.Sprintf("build services: %v", err))
	}
	return NewHandlerWithServices(logger, info, svcs)
}

func NewHandlerWithServices(logger *slog.Logger, info SystemInfo, svcs *service.Services) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if svcs == nil {
		panic("services is required")
	}

	info = normalizeSystemInfo(info)

	h := &Handler{
		logger:     logger,
		systemInfo: info,
		services:   svcs,
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
