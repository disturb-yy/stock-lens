package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminTokenMiddlewareRejectsMissingToken(t *testing.T) {
	router := newAdminTokenTestRouter("secret")

	req := httptest.NewRequest(http.MethodPost, "/sync", nil)
	req.Header.Set(HeaderRequestID, "req_admin")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"] != CodeUnauthorized {
		t.Fatalf("code = %v, want %s", body["code"], CodeUnauthorized)
	}
	if body["request_id"] != "req_admin" {
		t.Fatalf("request_id = %v, want req_admin", body["request_id"])
	}
}

func TestAdminTokenMiddlewareAllowsValidBearerToken(t *testing.T) {
	router := newAdminTokenTestRouter("secret")

	req := httptest.NewRequest(http.MethodPost, "/sync", nil)
	req.Header.Set(HeaderRequestID, "req_admin")
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
}

func newAdminTokenTestRouter(adminToken string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware(), AdminTokenMiddleware(adminToken))
	router.POST("/sync", func(c *gin.Context) {
		WriteSuccess(c, http.StatusAccepted, RequestID(c), gin.H{"status": "PENDING"})
	})
	return router
}
