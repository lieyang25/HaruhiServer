package transporthttp

import (
	"net/http"
	"time"

	"HaruhiServer/internal/apperr"
	"HaruhiServer/internal/response"
)

func (h *Handler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/readyz", h.readyz)
	mux.HandleFunc("/api/v1/system/info", h.systemInfoGet)
	mux.HandleFunc("/api/v1/projects", h.projects)
	mux.HandleFunc("/api/v1/projects/", h.projectByID)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperr.New(
			apperr.CodeMethodNotAllowed,
			"method not allowed",
		))
		return
	}

	h.logger.Debug(
		"healthz handled",
		"request_id", RequestIDFromContext(r.Context()),
	)

	response.OK(w, map[string]any{
		"status": "ok",
	})
}

func (h *Handler) readyz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperr.New(
			apperr.CodeMethodNotAllowed,
			"method not allowed",
		))
		return
	}

	h.logger.Debug(
		"readyz handled",
		"request_id", RequestIDFromContext(r.Context()),
	)

	response.OK(w, map[string]any{
		"status": "ready",
	})
}

func (h *Handler) systemInfoGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperr.New(
			apperr.CodeMethodNotAllowed,
			"method not allowed",
		))
		return
	}

	startedAt := h.systemInfo.StartedAt.UTC()
	uptime := time.Since(startedAt).Seconds()
	if uptime < 0 {
		uptime = 0
	}

	response.OK(w, map[string]any{
		"service":        h.systemInfo.Name,
		"version":        h.systemInfo.Version,
		"started_at":     startedAt.Format(time.RFC3339),
		"uptime_seconds": int64(uptime),
	})
}
