package model

type CompletionRequest struct {
	Prompt      string   `json:"prompt" binding:"required"`
	Model       string   `json:"model,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
}

type ChatCompletionRequest struct {
	Model       string        `json:"model,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
	Messages    []ChatMessage `json:"messages" binding:"required"`
}

type ChatMessage struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type ModelProbeRequest struct {
	Model string `json:"model" binding:"required"`
}
