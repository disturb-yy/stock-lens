package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWriteSuccessResponseShape(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/ok", func(c *gin.Context) {
		WriteSuccess(c, http.StatusAccepted, "req_123", gin.H{
			"task_uid": "01J2Z3ABCDEF123456789XYZAB",
			"status":   "PENDING",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["code"] != "OK" {
		t.Fatalf("code = %v, want OK", body["code"])
	}
	if body["message"] != "ok" {
		t.Fatalf("message = %v, want ok", body["message"])
	}
	if body["request_id"] != "req_123" {
		t.Fatalf("request_id = %v, want req_123", body["request_id"])
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data = %T, want object", body["data"])
	}
	if data["task_uid"] != "01J2Z3ABCDEF123456789XYZAB" {
		t.Fatalf("task_uid = %v", data["task_uid"])
	}
	if data["status"] != "PENDING" {
		t.Fatalf("status = %v", data["status"])
	}
}

func TestWriteErrorResponseShape(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/error", func(c *gin.Context) {
		WriteError(c, http.StatusUnauthorized, "UNAUTHORIZED", "admin token is required", "req_456")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["code"] != "UNAUTHORIZED" {
		t.Fatalf("code = %v, want UNAUTHORIZED", body["code"])
	}
	if body["message"] != "admin token is required" {
		t.Fatalf("message = %v", body["message"])
	}
	if body["request_id"] != "req_456" {
		t.Fatalf("request_id = %v, want req_456", body["request_id"])
	}
	if body["data"] != nil {
		t.Fatalf("data = %v, want nil", body["data"])
	}
}
