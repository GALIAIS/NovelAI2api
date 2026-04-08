package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"novelai/internal/model"
	"novelai/internal/novelai"
	"novelai/internal/service"
)

type OpenAIHandler struct {
	TextService  *service.TextService
	ImageService *service.ImageService
}

type openAICompletionRequest struct {
	Model       string        `json:"model"`
	Prompt      string        `json:"prompt"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stop        stopSequences `json:"stop,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type openAIChatCompletionRequest struct {
	Model       string                 `json:"model"`
	Messages    []model.ChatMessage    `json:"messages"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	Stop        stopSequences          `json:"stop,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Extra       map[string]interface{} `json:"-"`
}

type openAIResponsesRequest struct {
	Model           string        `json:"model"`
	Input           any           `json:"input"`
	Stream          bool          `json:"stream,omitempty"`
	MaxOutputTokens int           `json:"max_output_tokens,omitempty"`
	Temperature     float64       `json:"temperature,omitempty"`
	TopP            float64       `json:"top_p,omitempty"`
	Stop            stopSequences `json:"stop,omitempty"`
}

type openAIImageGenerationRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model,omitempty"`
	Size           string `json:"size,omitempty"` // e.g. 1024x1024
	N              int    `json:"n,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"` // b64_json/url
}

func (h *OpenAIHandler) ListModels(c *gin.Context) {
	session := c.MustGet("session").(*service.Session)
	models, err := h.TextService.ListOpenAIModels(c.Request.Context(), session.AuthToken)
	if err != nil {
		writeOpenAIUpstreamAwareError(c, err)
		return
	}
	data := make([]gin.H, 0, len(models))
	for _, id := range models {
		data = append(data, gin.H{"id": id, "object": "model", "created": 0, "owned_by": "novelai"})
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": data})
}

func (h *OpenAIHandler) Completions(c *gin.Context) {
	payload, err := decodeJSONMap(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]any{"message": "invalid request", "type": "invalid_request_error"}})
		return
	}
	isSillyTavern := isLikelySillyTavernPayload(payload)
	req, err := parseOpenAICompletionRequest(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]any{"message": "invalid request", "type": "invalid_request_error"}})
		return
	}
	stop := req.Stop.Values()
	if isSillyTavern {
		stop = mergeStringSlices(stop, defaultSillyTavernStopSequences)
	}
	session := c.MustGet("session").(*service.Session)
	serviceReq := model.CompletionRequest{
		Prompt:      req.Prompt,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        stop,
		Stream:      req.Stream,
	}
	if req.Stream {
		chunks, err := h.TextService.CompleteStream(c.Request.Context(), session.AuthToken, serviceReq)
		if err != nil {
			writeOpenAIUpstreamAwareError(c, err)
			return
		}
		if shouldTrimCompletionOutput(isSillyTavern, req.Prompt, stop) {
			chunks = normalizeSillyTavernCompletionChunks(chunks)
		}
		writeSSEHeaders(c)
		c.Status(http.StatusOK)
		id := fmt.Sprintf("cmpl-%d", time.Now().UnixNano())
		for _, chunk := range chunks {
			writeOpenAIStreamData(c, gin.H{
				"id":      id,
				"object":  "text_completion",
				"created": time.Now().Unix(),
				"model":   req.Model,
				"choices": []gin.H{{"text": chunk.Text(), "index": 0, "finish_reason": nil}},
			})
			c.Writer.Flush()
		}
		writeOpenAIStreamData(c, gin.H{
			"id":      id,
			"object":  "text_completion",
			"created": time.Now().Unix(),
			"model":   req.Model,
			"choices": []gin.H{{"text": "", "index": 0, "finish_reason": "stop"}},
		})
		c.Writer.Flush()
		_, _ = c.Writer.WriteString("data: [DONE]\n\n")
		c.Writer.Flush()
		return
	}

	resp, err := h.TextService.Complete(c.Request.Context(), session.AuthToken, serviceReq)
	if err != nil {
		writeOpenAIUpstreamAwareError(c, err)
		return
	}
	text := resp.Text
	if shouldTrimCompletionOutput(isSillyTavern, req.Prompt, stop) {
		text = trimSillyTavernCompletionText(text)
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      fmt.Sprintf("cmpl-%d", time.Now().UnixNano()),
		"object":  "text_completion",
		"created": time.Now().Unix(),
		"model":   req.Model,
		"choices": []gin.H{{"text": text, "index": 0, "finish_reason": "stop"}},
	})
}

func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	payload, err := decodeJSONMap(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]any{"message": "invalid request", "type": "invalid_request_error"}})
		return
	}
	req, err := parseOpenAIChatCompletionRequest(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]any{"message": "invalid request", "type": "invalid_request_error"}})
		return
	}
	stop := req.Stop.Values()
	if isLikelySillyTavernPayload(payload) {
		stop = mergeStringSlices(stop, defaultSillyTavernStopSequences)
	}
	session := c.MustGet("session").(*service.Session)
	serviceReq := model.ChatCompletionRequest{
		Model:       req.Model,
		Stream:      req.Stream,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        stop,
		Messages:    req.Messages,
	}
	if req.Stream {
		chunks, err := h.TextService.ChatStream(c.Request.Context(), session.AuthToken, serviceReq)
		if err != nil {
			writeOpenAIUpstreamAwareError(c, err)
			return
		}
		writeSSEHeaders(c)
		c.Status(http.StatusOK)
		id := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
		writeOpenAIStreamData(c, gin.H{
			"id":      id,
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   req.Model,
			"choices": []gin.H{{
				"index": 0,
				"delta": gin.H{"role": "assistant"},
			}},
		})
		c.Writer.Flush()
		for _, chunk := range chunks {
			writeOpenAIStreamData(c, gin.H{
				"id":      id,
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   req.Model,
				"choices": []gin.H{{
					"index": 0,
					"delta": gin.H{"content": chunk.Text()},
				}},
			})
			c.Writer.Flush()
		}
		writeOpenAIStreamData(c, gin.H{
			"id":      id,
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   req.Model,
			"choices": []gin.H{{
				"index":         0,
				"delta":         gin.H{},
				"finish_reason": "stop",
			}},
		})
		c.Writer.Flush()
		_, _ = c.Writer.WriteString("data: [DONE]\n\n")
		c.Writer.Flush()
		return
	}
	resp, err := h.TextService.Chat(c.Request.Context(), session.AuthToken, serviceReq)
	if err != nil {
		writeOpenAIUpstreamAwareError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   req.Model,
		"choices": []gin.H{{
			"index": 0,
			"message": gin.H{
				"role":    "assistant",
				"content": resp.Text,
			},
			"finish_reason": "stop",
		}},
	})
}

func (h *OpenAIHandler) Responses(c *gin.Context) {
	var req openAIResponsesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]any{"message": "invalid request", "type": "invalid_request_error"}})
		return
	}
	messages := normalizeResponseInput(req.Input)
	if len(messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]any{"message": "input is required", "type": "invalid_request_error"}})
		return
	}
	session := c.MustGet("session").(*service.Session)
	serviceReq := model.ChatCompletionRequest{
		Model:       req.Model,
		Stream:      req.Stream,
		MaxTokens:   req.MaxOutputTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop.Values(),
		Messages:    messages,
	}
	if req.Stream {
		chunks, err := h.TextService.ChatStream(c.Request.Context(), session.AuthToken, serviceReq)
		if err != nil {
			writeOpenAIUpstreamAwareError(c, err)
			return
		}
		responseID := fmt.Sprintf("resp-%d", time.Now().UnixNano())
		writeSSEHeaders(c)
		c.Status(http.StatusOK)
		for _, chunk := range chunks {
			writeOpenAIStreamData(c, gin.H{
				"type":        "response.output_text.delta",
				"response_id": responseID,
				"delta":       chunk.Text(),
			})
			c.Writer.Flush()
		}
		writeOpenAIStreamData(c, gin.H{"type": "response.completed", "response_id": responseID})
		c.Writer.Flush()
		return
	}
	resp, err := h.TextService.Chat(c.Request.Context(), session.AuthToken, serviceReq)
	if err != nil {
		writeOpenAIUpstreamAwareError(c, err)
		return
	}
	responseID := fmt.Sprintf("resp-%d", time.Now().UnixNano())
	c.JSON(http.StatusOK, gin.H{
		"id":      responseID,
		"object":  "response",
		"created": time.Now().Unix(),
		"model":   req.Model,
		"output": []gin.H{{
			"type": "message",
			"role": "assistant",
			"content": []gin.H{{
				"type": "output_text",
				"text": resp.Text,
			}},
		}},
	})
}

func (h *OpenAIHandler) ImageGenerations(c *gin.Context) {
	var req openAIImageGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Prompt) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]any{"message": "invalid request", "type": "invalid_request_error"}})
		return
	}
	format := strings.TrimSpace(req.ResponseFormat)
	if format == "" {
		format = "b64_json"
	}
	if format != "b64_json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]any{"message": "only b64_json response_format is supported", "type": "invalid_request_error"}})
		return
	}
	width, height := parseSize(req.Size)
	n := req.N
	if n <= 0 {
		n = 1
	}
	session := c.MustGet("session").(*service.Session)
	result, err := h.ImageService.Generate(c.Request.Context(), session.AuthToken, model.ImageGenerateRequest{
		Prompt:   req.Prompt,
		Model:    req.Model,
		Action:   "generate",
		Width:    width,
		Height:   height,
		NSamples: n,
	})
	if err != nil {
		writeOpenAIUpstreamAwareError(c, err)
		return
	}
	data := make([]gin.H, 0, len(result.Images))
	for _, image := range result.Images {
		data = append(data, gin.H{"b64_json": image.Base64})
	}
	c.JSON(http.StatusOK, gin.H{
		"created": time.Now().Unix(),
		"data":    data,
	})
}

func normalizeResponseInput(input any) []model.ChatMessage {
	switch typed := input.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		return []model.ChatMessage{{Role: "user", Content: typed}}
	case []any:
		var out []model.ChatMessage
		for _, item := range typed {
			switch obj := item.(type) {
			case map[string]any:
				role, _ := obj["role"].(string)
				content, _ := obj["content"].(string)
				if role == "" {
					role = "user"
				}
				if content == "" {
					continue
				}
				out = append(out, model.ChatMessage{Role: role, Content: content})
			}
		}
		return out
	default:
		return nil
	}
}

func writeOpenAIStreamData(c *gin.Context, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = c.Writer.WriteString("data: " + string(raw) + "\n\n")
}

type stopSequences []string

func (s *stopSequences) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if strings.TrimSpace(single) == "" {
			*s = nil
			return nil
		}
		*s = []string{single}
		return nil
	}
	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil {
		*s = multiple
		return nil
	}
	return fmt.Errorf("stop must be a string or string array")
}

func (s stopSequences) Values() []string {
	if len(s) == 0 {
		return nil
	}
	out := make([]string, 0, len(s))
	for _, item := range s {
		if strings.TrimSpace(item) == "" {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseSize(size string) (int, int) {
	if strings.TrimSpace(size) == "" {
		return 1024, 1024
	}
	parts := strings.Split(strings.ToLower(size), "x")
	if len(parts) != 2 {
		return 1024, 1024
	}
	w, errW := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, errH := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errW != nil || errH != nil || w <= 0 || h <= 0 {
		return 1024, 1024
	}
	return w, h
}

var defaultSillyTavernStopSequences = []string{"\nUser:", "\nuser:", "\nAssistant:", "\nassistant:", "\nSystem:", "\nsystem:"}
var roleLineCutPattern = regexp.MustCompile(`(?im)(?:^|[\r\n]|[。！？.!?"'”’])[ \t]*(user|assistant|system)[ \t]*[:：]`)

func decodeJSONMap(c *gin.Context) (map[string]any, error) {
	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		return nil, err
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

func parseOpenAICompletionRequest(payload map[string]any) (openAICompletionRequest, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return openAICompletionRequest{}, err
	}
	var aux struct {
		Model       string        `json:"model"`
		MaxTokens   int           `json:"max_tokens,omitempty"`
		Temperature float64       `json:"temperature,omitempty"`
		TopP        float64       `json:"top_p,omitempty"`
		Stop        stopSequences `json:"stop,omitempty"`
		Stream      bool          `json:"stream,omitempty"`
	}
	if err := json.Unmarshal(raw, &aux); err != nil {
		return openAICompletionRequest{}, err
	}
	req := openAICompletionRequest{
		Model:       aux.Model,
		MaxTokens:   aux.MaxTokens,
		Temperature: aux.Temperature,
		TopP:        aux.TopP,
		Stop:        aux.Stop,
		Stream:      aux.Stream,
	}
	req.Prompt = flattenCompletionPrompt(payload["prompt"])
	if strings.TrimSpace(req.Prompt) == "" {
		return openAICompletionRequest{}, fmt.Errorf("prompt is required")
	}
	return req, nil
}

func parseOpenAIChatCompletionRequest(payload map[string]any) (openAIChatCompletionRequest, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return openAIChatCompletionRequest{}, err
	}
	var req openAIChatCompletionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return openAIChatCompletionRequest{}, err
	}
	if len(req.Messages) == 0 {
		return openAIChatCompletionRequest{}, fmt.Errorf("messages is required")
	}
	return req, nil
}

func flattenCompletionPrompt(prompt any) string {
	switch typed := prompt.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(flattenCompletionPromptItem(item))
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func flattenCompletionPromptItem(item any) string {
	switch typed := item.(type) {
	case string:
		return typed
	case map[string]any:
		if text := strings.TrimSpace(stringValue(typed["content"])); text != "" {
			return text
		}
		if text := strings.TrimSpace(stringValue(typed["text"])); text != "" {
			return text
		}
		return strings.TrimSpace(fmt.Sprint(typed))
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func isLikelySillyTavernPayload(payload map[string]any) bool {
	if payload == nil {
		return false
	}
	if strings.TrimSpace(stringValue(payload["type"])) != "" {
		return true
	}
	if strings.TrimSpace(stringValue(payload["user_name"])) != "" && strings.TrimSpace(stringValue(payload["char_name"])) != "" {
		return true
	}
	if len(stringSliceValue(payload["group_names"])) > 0 {
		return true
	}
	if _, ok := payload["continue_prefill"]; ok {
		return true
	}
	if _, ok := payload["show_thoughts"]; ok {
		return true
	}
	for _, prompt := range collectSillyTavernSystemPrompts(payload) {
		lower := strings.ToLower(collapseWhitespace(prompt))
		if strings.Contains(lower, "fictional chat between") ||
			strings.Contains(lower, "[start a new chat]") ||
			strings.Contains(lower, "[continue your last message without repeating its original content.]") {
			return true
		}
	}
	return false
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func stringSliceValue(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if clean := strings.TrimSpace(stringValue(item)); clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func collectSillyTavernSystemPrompts(payload map[string]any) []string {
	items, ok := payload["messages"].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, raw := range items {
		msg, ok := raw.(map[string]any)
		if !ok || strings.TrimSpace(strings.ToLower(stringValue(msg["role"]))) != "system" {
			continue
		}
		text := collapseWhitespace(stringValue(msg["content"]))
		if text != "" {
			out = append(out, text)
		}
	}
	return out
}

func collapseWhitespace(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func mergeStringSlices(primary []string, defaults []string) []string {
	out := make([]string, 0, len(primary)+len(defaults))
	seen := map[string]struct{}{}
	appendValue := func(item string) {
		if strings.TrimSpace(item) == "" {
			return
		}
		if _, ok := seen[item]; ok {
			return
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	for _, item := range primary {
		appendValue(item)
	}
	for _, item := range defaults {
		appendValue(item)
	}
	return out
}

func trimSillyTavernCompletionText(text string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	first := roleLineCutPattern.FindStringSubmatchIndex(normalized)
	if first == nil {
		return strings.TrimSpace(normalized)
	}
	roleStart, roleEnd, markerEnd := first[2], first[3], first[1]
	prefix := strings.TrimSpace(normalized[:roleStart])
	if prefix != "" {
		return prefix
	}
	role := strings.ToLower(normalized[roleStart:roleEnd])
	rest := strings.TrimSpace(normalized[markerEnd:])
	if role != "assistant" {
		return ""
	}
	next := roleLineCutPattern.FindStringSubmatchIndex(rest)
	if next != nil && next[2] > 0 {
		return strings.TrimSpace(rest[:next[2]])
	}
	return strings.TrimSpace(rest)
}

func normalizeSillyTavernCompletionChunks(chunks []novelai.CompletionChunk) []novelai.CompletionChunk {
	if len(chunks) == 0 {
		return chunks
	}
	var full strings.Builder
	for _, chunk := range chunks {
		full.WriteString(chunk.Text())
	}
	trimmed := trimSillyTavernCompletionText(full.String())
	if trimmed == "" {
		return []novelai.CompletionChunk{}
	}
	return []novelai.CompletionChunk{{
		Choices: []struct {
			Text string `json:"text"`
		}{
			{Text: trimmed},
		},
	}}
}

func shouldTrimCompletionOutput(isSillyTavern bool, prompt string, stop []string) bool {
	if isSillyTavern {
		return true
	}
	lowerPrompt := strings.ToLower(prompt)
	if strings.Contains(lowerPrompt, "user:") || strings.Contains(lowerPrompt, "assistant:") || strings.Contains(lowerPrompt, "system:") ||
		strings.Contains(lowerPrompt, "user：") || strings.Contains(lowerPrompt, "assistant：") || strings.Contains(lowerPrompt, "system：") {
		return true
	}
	for _, marker := range stop {
		lower := strings.ToLower(strings.TrimSpace(marker))
		if strings.Contains(lower, "user:") || strings.Contains(lower, "assistant:") || strings.Contains(lower, "system:") {
			return true
		}
	}
	return false
}
