package novelai

import (
	"context"
	"encoding/json"
	"io"
	"strings"
)

type TextClient interface {
	Complete(ctx context.Context, token string, req CompletionRequest) (*CompletionResult, error)
	CompleteStream(ctx context.Context, token string, req CompletionRequest) ([]CompletionChunk, error)
	ListOpenAIModels(ctx context.Context, token string) ([]string, error)
	ProbeNativeModel(ctx context.Context, token string, model string) error
}

type CompletionRequest struct {
	Prompt      string
	Model       string
	MaxTokens   int
	Temperature float64
	TopP        float64
	Stop        []string
	Stream      bool
}

type CompletionResult struct {
	Text string
}

type openAICompletionResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

type openAIModelListResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type StaticTextClient struct{}

func (StaticTextClient) Complete(_ context.Context, _ string, req CompletionRequest) (*CompletionResult, error) {
	return &CompletionResult{Text: strings.TrimSpace(req.Prompt)}, nil
}

func (StaticTextClient) CompleteStream(_ context.Context, _ string, req CompletionRequest) ([]CompletionChunk, error) {
	return []CompletionChunk{{Choices: []struct {
		Text string `json:"text"`
	}{{Text: strings.TrimSpace(req.Prompt)}}}}, nil
}

func (StaticTextClient) ListOpenAIModels(_ context.Context, _ string) ([]string, error) {
	return []string{"glm-4-6", "glm-4-5", "xialong-v1"}, nil
}

func (StaticTextClient) ProbeNativeModel(_ context.Context, _ string, model string) error {
	switch model {
	case "llama-3-erato-v1", "kayra-v1":
		return &UpstreamError{StatusCode: 400, Body: []byte(`{"message":"Invalid request: Invalid packed input size"}`)}
	case "cassandra":
		return &UpstreamError{StatusCode: 403, Body: []byte(`{"message":"model 'cassandra' doesn't exist"}`)}
	default:
		return nil
	}
}

func (c *Client) Complete(ctx context.Context, token string, req CompletionRequest) (*CompletionResult, error) {
	payload := map[string]any{
		"prompt": req.Prompt,
		"stream": false,
	}
	if req.Model != "" {
		payload["model"] = req.Model
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	if req.Temperature != 0 {
		payload["temperature"] = req.Temperature
	}
	if req.TopP != 0 {
		payload["top_p"] = req.TopP
	}
	if len(req.Stop) > 0 {
		payload["stop"] = req.Stop
	}
	var out openAICompletionResponse
	if err := c.doJSON(ctx, "POST", c.TextBase+"/oa/v1/completions", token, payload, &out); err != nil {
		return nil, err
	}
	var b strings.Builder
	for _, choice := range out.Choices {
		b.WriteString(choice.Text)
	}
	return &CompletionResult{Text: b.String()}, nil
}

func (c *Client) CompleteStream(ctx context.Context, token string, req CompletionRequest) ([]CompletionChunk, error) {
	payload := map[string]any{
		"prompt": req.Prompt,
		"stream": true,
	}
	if req.Model != "" {
		payload["model"] = req.Model
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	if req.Temperature != 0 {
		payload["temperature"] = req.Temperature
	}
	if req.TopP != 0 {
		payload["top_p"] = req.TopP
	}
	if len(req.Stop) > 0 {
		payload["stop"] = req.Stop
	}
	resp, err := c.do(ctx, "POST", c.TextBase+"/oa/v1/completions", token, "application/json", strings.NewReader(string(mustJSON(payload))))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, &UpstreamError{StatusCode: resp.StatusCode, Body: buf}
	}
	return ParseCompletionStream(resp.Body)
}

func (c *Client) ListOpenAIModels(ctx context.Context, token string) ([]string, error) {
	var out openAIModelListResponse
	if err := c.doJSON(ctx, "GET", c.TextBase+"/oa/v1/models", token, nil, &out); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(out.Data))
	for _, item := range out.Data {
		if item.ID != "" {
			models = append(models, item.ID)
		}
	}
	return models, nil
}

func (c *Client) ProbeNativeModel(ctx context.Context, token string, model string) error {
	payload := map[string]any{
		"input": "aGVsbG8=", // intentionally invalid packed input for zero-cost model-recognition probe
		"model": model,
		"parameters": map[string]any{
			"max_length":               16,
			"min_length":               1,
			"temperature":              0.7,
			"top_k":                    40,
			"top_p":                    0.9,
			"tail_free_sampling":       1,
			"repetition_penalty":       1.05,
			"repetition_penalty_range": 2048,
			"repetition_penalty_slope": 0.09,
			"bad_words_ids":            []any{},
			"stop_sequences":           []any{},
			"generate_until_sentence":  false,
			"use_cache":                false,
			"return_full_text":         false,
		},
	}
	raw, _ := json.Marshal(payload)
	resp, err := c.do(ctx, "POST", c.TextBase+"/ai/generate", token, "application/json", strings.NewReader(string(raw)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		return &UpstreamError{StatusCode: resp.StatusCode, Body: buf}
	}
	return nil
}
