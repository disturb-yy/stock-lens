package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAccessLogMiddlewareWritesStructuredRequestLog(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware(), AccessLogMiddleware(logger))
	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(HeaderRequestID, "req_access")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}

	if entry["msg"] != "http_request" {
		t.Fatalf("msg = %v, want http_request", entry["msg"])
	}
	if entry["request_id"] != "req_access" {
		t.Fatalf("request_id = %v, want req_access", entry["request_id"])
	}
	if entry["method"] != http.MethodGet {
		t.Fatalf("method = %v, want %s", entry["method"], http.MethodGet)
	}
	if entry["path"] != "/healthz" {
		t.Fatalf("path = %v, want /healthz", entry["path"])
	}
	if entry["status"] != float64(http.StatusNoContent) {
		t.Fatalf("status = %v, want %d", entry["status"], http.StatusNoContent)
	}
}
