package transporthttp

import (
	"log/slog"
	"net/http"
)

type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) http.Handler {
	h := &Handler{
		logger: logger,
	}

	mux := http.NewServeMux()
	h.registerRoutes(mux)

	return Chain(
		mux,
		CORSMiddleware(),
		RecoverMiddleware(logger),
		LoggingMiddleware(logger),
		RequestIDMiddleware(),
	)
}
