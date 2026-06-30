package server

import "github.com/gin-gonic/gin"

const (
	CodeOK           = "OK"
	CodeUnauthorized = "UNAUTHORIZED"

	MessageOK           = "ok"
	MessageUnauthorized = "missing or invalid admin token"
)

type Response struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
}

func WriteSuccess(c *gin.Context, status int, requestID string, data any) {
	c.JSON(status, Response{
		Code:      CodeOK,
		Message:   MessageOK,
		RequestID: requestID,
		Data:      data,
	})
}

func WriteError(c *gin.Context, status int, code string, message string, requestID string) {
	WriteErrorData(c, status, code, message, requestID, nil)
}

func WriteErrorData(c *gin.Context, status int, code string, message string, requestID string, data any) {
	c.JSON(status, Response{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		Data:      data,
	})
}

func WriteErrorCode(c *gin.Context, code string, requestID string, data any) {
	spec := ErrorSpecForCode(code)
	WriteErrorData(c, spec.HTTPStatus, code, spec.Message, requestID, data)
}
