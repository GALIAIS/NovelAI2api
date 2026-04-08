package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"novelai/internal/novelai"
	httpapi "novelai/internal/httpapi"
)

func writeOpenAIUpstreamAwareError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := err.Error()
	if upstream, ok := asUpstreamError(err); ok {
		status = normalizeStatus(upstream.StatusCode)
		if extracted := extractUpstreamMessage(upstream.Body); extracted != "" {
			message = extracted
		}
	}
	c.JSON(status, gin.H{
		"error": gin.H{
			"message": message,
			"type":    errorTypeFromStatus(status),
		},
	})
}

func writeEnvelopeUpstreamAwareError(c *gin.Context, defaultStatus int, code int, message string, defaultType string, err error) {
	status := defaultStatus
	errType := defaultType
	errMessage := err.Error()
	if upstream, ok := asUpstreamError(err); ok {
		status = normalizeStatus(upstream.StatusCode)
		if status >= 400 && status < 500 {
			errType = "bad_request"
		} else if status >= 500 {
			errType = "internal_error"
		}
		if extracted := extractUpstreamMessage(upstream.Body); extracted != "" {
			errMessage = extracted
		}
	}
	httpapi.WriteError(c, status, code, message, errType, map[string]any{"error": errMessage})
}

func asUpstreamError(err error) (*novelai.UpstreamError, bool) {
	if err == nil {
		return nil, false
	}
	var upstream *novelai.UpstreamError
	if !errors.As(err, &upstream) || upstream == nil {
		return nil, false
	}
	return upstream, true
}

func normalizeStatus(code int) int {
	if code >= 400 && code <= 599 {
		return code
	}
	return http.StatusInternalServerError
}

func errorTypeFromStatus(status int) string {
	if status >= 400 && status < 500 {
		return "invalid_request_error"
	}
	return "server_error"
}

func extractUpstreamMessage(body []byte) string {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return ""
	}
	var payload struct {
		Message string `json:"message"`
		Error   any    `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		if payload.Message != "" {
			return payload.Message
		}
		if payload.Error != nil {
			b, _ := json.Marshal(payload.Error)
			return strings.TrimSpace(string(b))
		}
	}
	if len(raw) > 500 {
		return raw[:500]
	}
	return raw
}
