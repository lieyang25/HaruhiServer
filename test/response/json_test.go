package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"HaruhiServer/internal/apperr"
	"HaruhiServer/internal/response"
)

func decodeEnvelope(t *testing.T, body string) response.Envelope {
	t.Helper()
	var got response.Envelope
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", body, err)
	}
	return got
}

func TestOK(t *testing.T) {
	rr := httptest.NewRecorder()
	response.OK(rr, map[string]any{"k": "v"})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if !strings.HasSuffix(rr.Body.String(), "\n") {
		t.Fatalf("body should end with newline: %q", rr.Body.String())
	}

	env := decodeEnvelope(t, rr.Body.String())
	if env.Code != string(apperr.CodeOK) || env.Message != "ok" {
		t.Fatalf("env = %#v", env)
	}
	m, ok := env.Data.(map[string]any)
	if !ok || m["k"] != "v" {
		t.Fatalf("env.Data = %#v", env.Data)
	}
}

func TestError_AppError(t *testing.T) {
	rr := httptest.NewRecorder()
	response.Error(rr, apperr.New(apperr.CodeNotFound, "missing"))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	env := decodeEnvelope(t, rr.Body.String())
	if env.Code != string(apperr.CodeNotFound) || env.Message != "missing" {
		t.Fatalf("env = %#v", env)
	}
	if env.Data != nil {
		t.Fatalf("env.Data = %#v, want nil", env.Data)
	}
}

func TestError_PlainErrorFallsBackToInternal(t *testing.T) {
	rr := httptest.NewRecorder()
	response.Error(rr, errors.New("x"))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	env := decodeEnvelope(t, rr.Body.String())
	if env.Code != string(apperr.CodeInternal) || env.Message != "internal error" {
		t.Fatalf("env = %#v", env)
	}
}
