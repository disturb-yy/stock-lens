package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorSpecForCode(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "invalid argument",
			code:        CodeInvalidArgument,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid argument",
		},
		{
			name:        "unauthorized",
			code:        CodeUnauthorized,
			wantStatus:  http.StatusUnauthorized,
			wantMessage: MessageUnauthorized,
		},
		{
			name:        "stock not found",
			code:        CodeMarketStockNotFound,
			wantStatus:  http.StatusNotFound,
			wantMessage: "stock not found",
		},
		{
			name:        "sync task conflict",
			code:        CodeMarketSyncTaskConflict,
			wantStatus:  http.StatusConflict,
			wantMessage: "sync task already running",
		},
		{
			name:        "provider error",
			code:        CodeMarketProviderError,
			wantStatus:  http.StatusBadGateway,
			wantMessage: "market provider error",
		},
		{
			name:        "unknown code",
			code:        "UNKNOWN",
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorSpecForCode(tt.code)
			if got.HTTPStatus != tt.wantStatus {
				t.Fatalf("HTTPStatus = %d, want %d", got.HTTPStatus, tt.wantStatus)
			}
			if got.Message != tt.wantMessage {
				t.Fatalf("Message = %q, want %q", got.Message, tt.wantMessage)
			}
		})
	}
}

func TestWriteErrorCodeUsesMappedSpecAndData(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/conflict", func(c *gin.Context) {
		WriteErrorCode(c, CodeMarketSyncTaskConflict, "req_conflict", gin.H{
			"task_uid": "01J2Z3ABCDEF123456789XYZAB",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/conflict", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"] != CodeMarketSyncTaskConflict {
		t.Fatalf("code = %v, want %s", body["code"], CodeMarketSyncTaskConflict)
	}
	if body["message"] != "sync task already running" {
		t.Fatalf("message = %v", body["message"])
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data = %T, want object", body["data"])
	}
	if data["task_uid"] != "01J2Z3ABCDEF123456789XYZAB" {
		t.Fatalf("task_uid = %v", data["task_uid"])
	}
}
