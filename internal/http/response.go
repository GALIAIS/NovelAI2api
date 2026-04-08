package http

import (
	"github.com/gin-gonic/gin"
	"novelai/internal/model"
)

func WriteSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, model.Envelope{Code: 0, Message: "ok", Data: data})
}

func WriteError(c *gin.Context, status int, code int, message string, errType string, details map[string]any) {
	c.JSON(status, model.Envelope{
		Code:    code,
		Message: message,
		Error: &model.APIError{
			Type:    errType,
			Details: details,
		},
	})
}
