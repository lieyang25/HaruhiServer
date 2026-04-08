package transporthttp

import (
	"net/http"

	"HaruhiServer/internal/apperr"
	"HaruhiServer/internal/response"
)

func (h *Handler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
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
