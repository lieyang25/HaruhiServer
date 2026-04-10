package apperr_test

import (
	"errors"
	"net/http"
	"testing"

	"HaruhiServer/internal/apperr"
)

func TestNewWrapAsAndUnwrap(t *testing.T) {
	base := errors.New("boom")
	wrapped := apperr.Wrap(apperr.CodeConflict, "conflict", base)

	if got := wrapped.Error(); got != "CONFLICT: conflict" {
		t.Fatalf("Error() = %q", got)
	}
	if !errors.Is(wrapped, base) {
		t.Fatal("errors.Is(wrapped, base) = false, want true")
	}

	appErr := apperr.As(wrapped)
	if appErr == nil {
		t.Fatal("As(wrapped) = nil")
	}
	if appErr.Code != apperr.CodeConflict || appErr.Message != "conflict" {
		t.Fatalf("As(wrapped) = %#v", appErr)
	}
}

func TestAsNilAndNonAppError(t *testing.T) {
	if apperr.As(nil) != nil {
		t.Fatal("As(nil) != nil")
	}
	if apperr.As(errors.New("x")) != nil {
		t.Fatal("As(non-app-error) != nil")
	}
}

func TestHTTPStatus(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{apperr.New(apperr.CodeInvalidArgument, "bad request"), http.StatusBadRequest},
		{apperr.New(apperr.CodeNotFound, "not found"), http.StatusNotFound},
		{apperr.New(apperr.CodeMethodNotAllowed, "nope"), http.StatusMethodNotAllowed},
		{apperr.New(apperr.CodeConflict, "conflict"), http.StatusConflict},
		{apperr.New(apperr.CodeInternal, "internal"), http.StatusInternalServerError},
		{errors.New("plain"), http.StatusInternalServerError},
		{nil, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		got := apperr.HTTPStatus(tc.err)
		if got != tc.want {
			t.Fatalf("HTTPStatus(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}
