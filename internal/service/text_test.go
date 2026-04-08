package service

import (
	"context"
	"slices"
	"testing"

	"novelai/internal/model"
	"novelai/internal/novelai"
)

func TestTextService(t *testing.T) {
	svc := &TextService{Client: novelai.StaticTextClient{}}
	resp, err := svc.Complete(t.Context(), "token", model.CompletionRequest{Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != "hello" {
		t.Fatalf("text = %q", resp.Text)
	}
}

func TestMessagesToPromptConcatenatesRoles(t *testing.T) {
	got := MessagesToPrompt([]model.ChatMessage{{Role: "system", Content: "Be nice"}, {Role: "user", Content: "Hello"}})
	want := "System: Be nice\nUser: Hello\nAssistant:"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestProbeModel(t *testing.T) {
	svc := &TextService{Client: novelai.StaticTextClient{}}
	resp, err := svc.ProbeModel(t.Context(), "token", model.ModelProbeRequest{Model: "llama-3-erato-v1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.OAAvailable {
		t.Fatalf("oa_available = true")
	}
	if !resp.NativeRecognized {
		t.Fatalf("native_recognized = false")
	}
}

type captureTextClient struct {
	lastCompleteReq novelai.CompletionRequest
	lastStreamReq   novelai.CompletionRequest
}

func (c *captureTextClient) Complete(_ context.Context, _ string, req novelai.CompletionRequest) (*novelai.CompletionResult, error) {
	c.lastCompleteReq = req
	return &novelai.CompletionResult{Text: "ok"}, nil
}

func (c *captureTextClient) CompleteStream(_ context.Context, _ string, req novelai.CompletionRequest) ([]novelai.CompletionChunk, error) {
	c.lastStreamReq = req
	return []novelai.CompletionChunk{{Choices: []struct {
		Text string `json:"text"`
	}{{Text: "ok"}}}}, nil
}

func (c *captureTextClient) ListOpenAIModels(_ context.Context, _ string) ([]string, error) {
	return []string{"xialong-v1"}, nil
}

func (c *captureTextClient) ProbeNativeModel(_ context.Context, _ string, _ string) error {
	return nil
}

func TestChatPassesSamplingAndTokenParams(t *testing.T) {
	client := &captureTextClient{}
	svc := &TextService{Client: client}
	_, err := svc.Chat(t.Context(), "token", model.ChatCompletionRequest{
		Model:       "xialong-v1",
		MaxTokens:   512,
		Temperature: 0.7,
		TopP:        0.95,
		Messages:    []model.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if client.lastCompleteReq.MaxTokens != 512 {
		t.Fatalf("max_tokens = %d", client.lastCompleteReq.MaxTokens)
	}
	if client.lastCompleteReq.Temperature != 0.7 {
		t.Fatalf("temperature = %v", client.lastCompleteReq.Temperature)
	}
	if client.lastCompleteReq.TopP != 0.95 {
		t.Fatalf("top_p = %v", client.lastCompleteReq.TopP)
	}
	if !slices.Contains(client.lastCompleteReq.Stop, "\nUser:") || !slices.Contains(client.lastCompleteReq.Stop, "\nuser:") {
		t.Fatalf("stop = %#v", client.lastCompleteReq.Stop)
	}
}

func TestChatUsesDefaultMaxTokensWhenMissing(t *testing.T) {
	client := &captureTextClient{}
	svc := &TextService{Client: client}
	_, err := svc.Chat(t.Context(), "token", model.ChatCompletionRequest{
		Model:    "xialong-v1",
		Messages: []model.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if client.lastCompleteReq.MaxTokens != defaultChatMaxTokens {
		t.Fatalf("max_tokens = %d", client.lastCompleteReq.MaxTokens)
	}
}

func TestChatStreamPassesSamplingAndTokenParams(t *testing.T) {
	client := &captureTextClient{}
	svc := &TextService{Client: client}
	_, err := svc.ChatStream(t.Context(), "token", model.ChatCompletionRequest{
		Model:       "xialong-v1",
		MaxTokens:   256,
		Temperature: 0.2,
		TopP:        0.8,
		Messages:    []model.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if client.lastStreamReq.MaxTokens != 256 {
		t.Fatalf("max_tokens = %d", client.lastStreamReq.MaxTokens)
	}
	if client.lastStreamReq.Temperature != 0.2 {
		t.Fatalf("temperature = %v", client.lastStreamReq.Temperature)
	}
	if client.lastStreamReq.TopP != 0.8 {
		t.Fatalf("top_p = %v", client.lastStreamReq.TopP)
	}
	if !slices.Contains(client.lastStreamReq.Stop, "\nUser:") || !slices.Contains(client.lastStreamReq.Stop, "\nuser:") {
		t.Fatalf("stop = %#v", client.lastStreamReq.Stop)
	}
}

func TestChatTruncatesInjectedRoleContinuation(t *testing.T) {
	svc := &TextService{Client: &captureContinuationClient{}}
	resp, err := svc.Chat(t.Context(), "token", model.ChatCompletionRequest{
		Model:    "xialong-v1",
		Messages: []model.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != "first answer" {
		t.Fatalf("text = %q", resp.Text)
	}
}

func TestChatStreamTruncatesInjectedRoleContinuation(t *testing.T) {
	svc := &TextService{Client: &captureContinuationClient{}}
	chunks, err := svc.ChatStream(t.Context(), "token", model.ChatCompletionRequest{
		Model:    "xialong-v1",
		Messages: []model.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 1 {
		t.Fatalf("chunks = %d", len(chunks))
	}
	if chunks[0].Text() != "first answer" {
		t.Fatalf("text = %q", chunks[0].Text())
	}
}

func TestChatRespectsCustomStopSequences(t *testing.T) {
	client := &captureTextClient{}
	svc := &TextService{Client: client}
	_, err := svc.Chat(t.Context(), "token", model.ChatCompletionRequest{
		Model:    "xialong-v1",
		Stop:     []string{"END"},
		Messages: []model.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(client.lastCompleteReq.Stop, "END") || !slices.Contains(client.lastCompleteReq.Stop, "\nUser:") || !slices.Contains(client.lastCompleteReq.Stop, "\nuser:") {
		t.Fatalf("stop = %#v", client.lastCompleteReq.Stop)
	}
}

func TestTruncateDialogueContinuationHandlesRoleAtLineStart(t *testing.T) {
	got := truncateDialogueContinuation("Assistant: leaked prefix\nreal text")
	if got != "leaked prefix\nreal text" {
		t.Fatalf("got = %q", got)
	}
}

func TestTruncateDialogueContinuationHandlesCRLF(t *testing.T) {
	got := truncateDialogueContinuation("ok answer\r\nUser: leaked")
	if got != "ok answer" {
		t.Fatalf("got = %q", got)
	}
}

func TestTruncateDialogueContinuationHandlesRoleAfterPunctuation(t *testing.T) {
	got := truncateDialogueContinuation("这段是正文。User: leaked")
	if got != "这段是正文。" {
		t.Fatalf("got = %q", got)
	}
}

type captureContinuationClient struct{}

func (c *captureContinuationClient) Complete(_ context.Context, _ string, req novelai.CompletionRequest) (*novelai.CompletionResult, error) {
	return &novelai.CompletionResult{
		Text: "first answer\nUser: next turn\nAssistant: reply",
	}, nil
}

func (c *captureContinuationClient) CompleteStream(_ context.Context, _ string, req novelai.CompletionRequest) ([]novelai.CompletionChunk, error) {
	return []novelai.CompletionChunk{{
		Choices: []struct {
			Text string `json:"text"`
		}{{Text: "first answer\nUser: next turn"}},
	}}, nil
}

func (c *captureContinuationClient) ListOpenAIModels(_ context.Context, _ string) ([]string, error) {
	return []string{"xialong-v1"}, nil
}

func (c *captureContinuationClient) ProbeNativeModel(_ context.Context, _ string, _ string) error {
	return nil
}
