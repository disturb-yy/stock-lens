package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterHealthz(t *testing.T) {
	router := NewRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRouterAddsRequestIDHeader(t *testing.T) {
	router := NewRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get(HeaderRequestID); got == "" {
		t.Fatalf("%s is empty", HeaderRequestID)
	}
}

func TestRouterReadyz(t *testing.T) {
	tests := []struct {
		name   string
		ready  ReadinessCheck
		status int
	}{
		{
			name: "ready",
			ready: func() error {
				return nil
			},
			status: http.StatusOK,
		},
		{
			name: "not ready",
			ready: func() error {
				return errors.New("database unavailable")
			},
			status: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter(tt.ready)

			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
		})
	}
}
