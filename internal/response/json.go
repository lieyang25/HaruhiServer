package response

import (
	"encoding/json"
	"net/http"

	"HaruhiServer/internal/apperr"
)

type Envelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, Envelope{
		Code:    string(apperr.CodeOK),
		Message: "ok",
		Data:    data,
	})
}

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
