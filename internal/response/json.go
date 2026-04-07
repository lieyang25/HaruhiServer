package response

import (
	"encoding/json"
	"net/http"

	"HaruhiServer/internal/apperr"
)

type Envelope struct {
	// Code is the stable business result code.
	Code    string `json:"code"`
	// Message is a short human-readable description.
	Message string `json:"message"`
	// Data carries successful payload and is omitted on nil.
	Data    any    `json:"data,omitempty"`
}

// OK writes a standard success envelope with HTTP 200.
func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, Envelope{
		Code:    string(apperr.CodeOK),
		Message: "ok",
		Data:    data,
	})
}

// Error converts any error into a standard error envelope.
func Error(w http.ResponseWriter, err error) {
	appErr := apperr.As(err)
	if appErr == nil {
		appErr = apperr.Wrap(apperr.CodeInternal, "internal error", err)
	}

	writeJSON(w, apperr.HTTPStatus(appErr), Envelope{
		Code:    string(appErr.Code),
		Message: appErr.Message,
	})
}

// writeJSON marshals envelope and writes status/body consistently.
func writeJSON(w http.ResponseWriter, status int, v Envelope) {
	b, err := json.Marshal(v)

	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(append(b, '\n'))
}
