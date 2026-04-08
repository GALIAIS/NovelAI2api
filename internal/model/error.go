package model

type APIError struct {
	Type    string         `json:"type"`
	Details map[string]any `json:"details,omitempty"`
}

type Envelope struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}
