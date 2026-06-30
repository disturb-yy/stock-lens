package server

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const (
	HeaderRequestID  = "X-Request-ID"
	contextRequestID = "request_id"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(contextRequestID, requestID)
		c.Header(HeaderRequestID, requestID)
		c.Next()
	}
}

func RequestID(c *gin.Context) string {
	value, ok := c.Get(contextRequestID)
	if !ok {
		return ""
	}
	requestID, _ := value.(string)
	return requestID
}

func newRequestID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		// 随机源异常时仍返回可关联的进程内请求 ID，避免响应缺失 request_id。
		return "req_unavailable"
	}
	return "req_" + hex.EncodeToString(bytes[:])
}
