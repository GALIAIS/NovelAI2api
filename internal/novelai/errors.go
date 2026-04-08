package novelai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type UpstreamError struct {
	StatusCode int
	Body       []byte
}

func (e *UpstreamError) Error() string {
	if e == nil {
		return "novelai upstream error"
	}
	message := extractUpstreamMessage(e.Body)
	if message == "" {
		return fmt.Sprintf("novelai upstream error (status=%d)", e.StatusCode)
	}
	return fmt.Sprintf("novelai upstream error (status=%d): %s", e.StatusCode, message)
}

func extractUpstreamMessage(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return ""
	}
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Message != "" {
		return payload.Message
	}
	if len(trimmed) > 240 {
		return trimmed[:240]
	}
	return trimmed
}
