package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"novelai/internal/model"
	"novelai/internal/novelai"
)

type TextService struct {
	Client novelai.TextClient
}

const defaultChatMaxTokens = 1024

var defaultChatStopSequences = []string{"\nUser:", "\nuser:", "\nSystem:", "\nsystem:"}
var roleLineMarkerPattern = regexp.MustCompile(`(?im)(?:^|[\r\n]|[。！？.!?"'”’])[ \t]*(user|assistant|system)[ \t]*[:：]`)

func (s *TextService) ListOpenAIModels(ctx context.Context, token string) ([]string, error) {
	return s.Client.ListOpenAIModels(ctx, token)
}

func (s *TextService) Complete(ctx context.Context, token string, req model.CompletionRequest) (*model.CompletionResponse, error) {
	result, err := s.Client.Complete(ctx, token, novelai.CompletionRequest{
		Prompt:      req.Prompt,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        normalizeStopSequences(req.Stop),
		Stream:      req.Stream,
	})
	if err != nil {
		return nil, err
	}
	return &model.CompletionResponse{Text: result.Text}, nil
}

func (s *TextService) CompleteStream(ctx context.Context, token string, req model.CompletionRequest) ([]novelai.CompletionChunk, error) {
	return s.Client.CompleteStream(ctx, token, novelai.CompletionRequest{
		Prompt:      req.Prompt,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        normalizeStopSequences(req.Stop),
		Stream:      true,
	})
}

func (s *TextService) Chat(ctx context.Context, token string, req model.ChatCompletionRequest) (*model.CompletionResponse, error) {
	prompt := MessagesToPrompt(req.Messages)
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultChatMaxTokens
	}
	result, err := s.Client.Complete(ctx, token, novelai.CompletionRequest{
		Prompt:      prompt,
		Model:       req.Model,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        mergeStopSequences(req.Stop, defaultChatStopSequences),
		Stream:      req.Stream,
	})
	if err != nil {
		return nil, err
	}
	return &model.CompletionResponse{Text: truncateDialogueContinuation(result.Text)}, nil
}

func (s *TextService) ChatStream(ctx context.Context, token string, req model.ChatCompletionRequest) ([]novelai.CompletionChunk, error) {
	prompt := MessagesToPrompt(req.Messages)
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultChatMaxTokens
	}
	chunks, err := s.Client.CompleteStream(ctx, token, novelai.CompletionRequest{
		Prompt:      prompt,
		Model:       req.Model,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        mergeStopSequences(req.Stop, defaultChatStopSequences),
		Stream:      true,
	})
	if err != nil {
		return nil, err
	}
	return truncateDialogueContinuationInChunks(chunks), nil
}

func MessagesToPrompt(messages []model.ChatMessage) string {
	var b strings.Builder
	for _, message := range messages {
		switch message.Role {
		case "system":
			b.WriteString("System: ")
		case "user":
			b.WriteString("User: ")
		case "assistant":
			b.WriteString("Assistant: ")
		default:
			b.WriteString("User: ")
		}
		b.WriteString(message.Content)
		b.WriteString("\n")
	}
	b.WriteString("Assistant:")
	return b.String()
}

func (s *TextService) ProbeModel(ctx context.Context, token string, req model.ModelProbeRequest) (*model.ModelProbeResponse, error) {
	targetModel := strings.TrimSpace(req.Model)
	resp := &model.ModelProbeResponse{Model: targetModel}

	oaModels, err := s.Client.ListOpenAIModels(ctx, token)
	if err != nil {
		return nil, err
	}
	for _, item := range oaModels {
		if item == targetModel {
			resp.OAAvailable = true
			break
		}
	}

	nativeErr := s.Client.ProbeNativeModel(ctx, token, targetModel)
	if nativeErr == nil {
		resp.NativeRecognized = true
		resp.NativeStatusCode = 200
		return resp, nil
	}
	upstream, ok := nativeErr.(*novelai.UpstreamError)
	if !ok {
		return nil, nativeErr
	}

	resp.NativeStatusCode = upstream.StatusCode
	resp.NativeMessage = extractMessage(upstream.Body)
	if upstream.StatusCode == 400 && strings.Contains(strings.ToLower(resp.NativeMessage), "invalid packed input") {
		resp.NativeRecognized = true
	}
	return resp, nil
}

func extractMessage(body []byte) string {
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Message != "" {
		return payload.Message
	}
	return strings.TrimSpace(string(body))
}

func normalizeStopSequences(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	out := make([]string, 0, len(input))
	seen := map[string]struct{}{}
	for _, item := range input {
		if strings.TrimSpace(item) == "" {
			continue
		}
		value := item
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func mergeStopSequences(primary []string, defaults []string) []string {
	out := normalizeStopSequences(primary)
	seen := make(map[string]struct{}, len(out))
	for _, item := range out {
		seen[item] = struct{}{}
	}
	for _, item := range defaults {
		if strings.TrimSpace(item) == "" {
			continue
		}
		value := item
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func truncateDialogueContinuation(text string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	first := roleLineMarkerPattern.FindStringSubmatchIndex(normalized)
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
	next := roleLineMarkerPattern.FindStringSubmatchIndex(rest)
	if next != nil && next[2] > 0 {
		return strings.TrimSpace(rest[:next[2]])
	}
	return strings.TrimSpace(rest)
}

func truncateDialogueContinuationInChunks(chunks []novelai.CompletionChunk) []novelai.CompletionChunk {
	if len(chunks) == 0 {
		return chunks
	}
	var merged strings.Builder
	for _, chunk := range chunks {
		merged.WriteString(chunk.Text())
	}
	text := truncateDialogueContinuation(merged.String())
	if text == "" {
		return []novelai.CompletionChunk{}
	}
	return []novelai.CompletionChunk{{
		Choices: []struct {
			Text string `json:"text"`
		}{
			{Text: text},
		},
	}}
}
