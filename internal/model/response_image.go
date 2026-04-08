package model

type ImageResponse struct {
	Images []ImagePayload   `json:"images"`
	Meta   map[string]int64 `json:"meta,omitempty"`
}

type ImagePayload struct {
	MIMEType string `json:"mime_type"`
	Base64   string `json:"base64"`
}
