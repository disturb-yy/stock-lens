package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddlewareUsesIncomingRequestID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/ok", func(c *gin.Context) {
		WriteSuccess(c, http.StatusOK, RequestID(c), gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set(HeaderRequestID, "req_client")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get(HeaderRequestID); got != "req_client" {
		t.Fatalf("%s = %q, want %q", HeaderRequestID, got, "req_client")
	}
	assertResponseRequestID(t, rec, "req_client")
}

func TestRequestIDMiddlewareGeneratesRequestID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/ok", func(c *gin.Context) {
		WriteSuccess(c, http.StatusOK, RequestID(c), gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	requestID := rec.Header().Get(HeaderRequestID)
	if !strings.HasPrefix(requestID, "req_") {
		t.Fatalf("%s = %q, want req_ prefix", HeaderRequestID, requestID)
	}
	assertResponseRequestID(t, rec, requestID)
}

func assertResponseRequestID(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["request_id"] != want {
		t.Fatalf("request_id = %v, want %s", body["request_id"], want)
	}
}
