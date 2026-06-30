package server

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const authorizationBearerPrefix = "Bearer "

func AdminTokenMiddleware(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(adminToken)) != 1 {
			WriteError(c, http.StatusUnauthorized, CodeUnauthorized, MessageUnauthorized, RequestID(c))
			c.Abort()
			return
		}

		c.Next()
	}
}

func bearerToken(header string) string {
	if !strings.HasPrefix(header, authorizationBearerPrefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, authorizationBearerPrefix))
}
