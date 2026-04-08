package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"novelai/internal/novelai"
	"novelai/internal/service"
)

type upstreamFailTextClient struct{}

func (upstreamFailTextClient) Complete(context.Context, string, novelai.CompletionRequest) (*novelai.CompletionResult, error) {
	return nil, &novelai.UpstreamError{StatusCode: 400, Body: []byte(`{"message":"bad input"}`)}
}
func (upstreamFailTextClient) CompleteStream(context.Context, string, novelai.CompletionRequest) ([]novelai.CompletionChunk, error) {
	return nil, &novelai.UpstreamError{StatusCode: 400, Body: []byte(`{"message":"bad input"}`)}
}
func (upstreamFailTextClient) ListOpenAIModels(context.Context, string) ([]string, error) {
	return nil, &novelai.UpstreamError{StatusCode: 400, Body: []byte(`{"message":"bad input"}`)}
}
func (upstreamFailTextClient) ProbeNativeModel(context.Context, string, string) error {
	return &novelai.UpstreamError{StatusCode: 400, Body: []byte(`{"message":"bad input"}`)}
}

type captureCompletionClient struct {
	lastCompleteReq novelai.CompletionRequest
}

func (c *captureCompletionClient) Complete(_ context.Context, _ string, req novelai.CompletionRequest) (*novelai.CompletionResult, error) {
	c.lastCompleteReq = req
	return &novelai.CompletionResult{Text: "hello\nUser: leaked"}, nil
}

func (c *captureCompletionClient) CompleteStream(_ context.Context, _ string, req novelai.CompletionRequest) ([]novelai.CompletionChunk, error) {
	c.lastCompleteReq = req
	return []novelai.CompletionChunk{{
		Choices: []struct {
			Text string `json:"text"`
		}{
			{Text: "hello\nAssistant: leaked"},
		},
	}}, nil
}

func (c *captureCompletionClient) ListOpenAIModels(_ context.Context, _ string) ([]string, error) {
	return []string{"xialong-v1"}, nil
}

func (c *captureCompletionClient) ProbeNativeModel(_ context.Context, _ string, _ string) error {
	return nil
}

func TestOpenAIChatCompletions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(`{"model":"xialong-v1","messages":[{"role":"user","content":"hello"}]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &OpenAIHandler{
		TextService:  &service.TextService{Client: novelai.StaticTextClient{}},
		ImageService: &service.ImageService{},
	}
	h.ChatCompletions(c)

	if w.Code != 200 {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["object"] != "chat.completion" {
		t.Fatalf("object = %v", body["object"])
	}
}

func TestParseSize(t *testing.T) {
	w, h := parseSize("512x768")
	if w != 512 || h != 768 {
		t.Fatalf("size = %dx%d", w, h)
	}
}

func TestOpenAIErrorStatusPassThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/v1/models", nil)
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &OpenAIHandler{
		TextService:  &service.TextService{Client: upstreamFailTextClient{}},
		ImageService: &service.ImageService{},
	}
	h.ListModels(c)

	if w.Code != 400 {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestOpenAIResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/responses", strings.NewReader(`{"model":"xialong-v1","input":"hello"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &OpenAIHandler{
		TextService:  &service.TextService{Client: novelai.StaticTextClient{}},
		ImageService: &service.ImageService{},
	}
	h.Responses(c)

	if w.Code != 200 {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["object"] != "response" {
		t.Fatalf("object = %v", body["object"])
	}
}

func TestStopSequencesUnmarshalSupportsStringAndArray(t *testing.T) {
	var single stopSequences
	if err := json.Unmarshal([]byte(`"END"`), &single); err != nil {
		t.Fatal(err)
	}
	if len(single) != 1 || single[0] != "END" {
		t.Fatalf("single = %#v", single)
	}

	var multi stopSequences
	if err := json.Unmarshal([]byte(`["A","B"]`), &multi); err != nil {
		t.Fatal(err)
	}
	if len(multi) != 2 || multi[0] != "A" || multi[1] != "B" {
		t.Fatalf("multi = %#v", multi)
	}
}

func TestOpenAICompletionsSillyTavernBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/completions", strings.NewReader(`{
		"model":"xialong-v1",
		"prompt":["hello","world"],
		"user_name":"Alice",
		"char_name":"Lilith"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	client := &captureCompletionClient{}
	h := &OpenAIHandler{
		TextService:  &service.TextService{Client: client},
		ImageService: &service.ImageService{},
	}
	h.Completions(c)

	if w.Code != 200 {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	if client.lastCompleteReq.Prompt != "hello\nworld" {
		t.Fatalf("prompt = %q", client.lastCompleteReq.Prompt)
	}
	if !slices.Contains(client.lastCompleteReq.Stop, "\nUser:") || !slices.Contains(client.lastCompleteReq.Stop, "\nuser:") {
		t.Fatalf("stop = %#v", client.lastCompleteReq.Stop)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	choices, _ := body["choices"].([]any)
	if len(choices) != 1 {
		t.Fatalf("choices = %#v", body["choices"])
	}
	choice, _ := choices[0].(map[string]any)
	if choice["text"] != "hello" {
		t.Fatalf("text = %v", choice["text"])
	}
}

func TestOpenAIChatCompletionsSillyTavernBranchAddsStops(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(`{
		"model":"xialong-v1",
		"messages":[{"role":"user","content":"hello"}],
		"user_name":"Alice",
		"char_name":"Lilith"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	client := &captureCompletionClient{}
	h := &OpenAIHandler{
		TextService:  &service.TextService{Client: client},
		ImageService: &service.ImageService{},
	}
	h.ChatCompletions(c)

	if w.Code != 200 {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	if !slices.Contains(client.lastCompleteReq.Stop, "\nUser:") || !slices.Contains(client.lastCompleteReq.Stop, "\nassistant:") {
		t.Fatalf("stop = %#v", client.lastCompleteReq.Stop)
	}
}

func TestIsLikelySillyTavernPayloadBySystemPrompt(t *testing.T) {
	payload := map[string]any{
		"messages": []any{
			map[string]any{"role": "system", "content": "[Start a new chat]"},
			map[string]any{"role": "user", "content": "hello"},
		},
	}
	if !isLikelySillyTavernPayload(payload) {
		t.Fatalf("expected sillytavern payload")
	}
}

func TestOpenAICompletionsTrimByPromptPatternWithoutSTMarkers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/completions", strings.NewReader(`{
		"model":"xialong-v1",
		"prompt":"User: hi\\nAssistant:"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	client := &captureCompletionClient{}
	h := &OpenAIHandler{
		TextService:  &service.TextService{Client: client},
		ImageService: &service.ImageService{},
	}
	h.Completions(c)

	if w.Code != 200 {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	choices, _ := body["choices"].([]any)
	choice, _ := choices[0].(map[string]any)
	if choice["text"] != "hello" {
		t.Fatalf("text = %v", choice["text"])
	}
}

func TestTrimSillyTavernCompletionTextHandlesLowercaseRoles(t *testing.T) {
	got := trimSillyTavernCompletionText("assistant: ok line\nuser: leaked")
	if got != "ok line" {
		t.Fatalf("got = %q", got)
	}
}

func TestTrimSillyTavernCompletionTextHandlesRoleAfterPunctuation(t *testing.T) {
	got := trimSillyTavernCompletionText("正文。User: leaked")
	if got != "正文。" {
		t.Fatalf("got = %q", got)
	}
}
