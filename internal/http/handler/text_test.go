package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"novelai/internal/novelai"
	"novelai/internal/service"
)

type upstreamFailClient struct{}

func (upstreamFailClient) Complete(context.Context, string, novelai.CompletionRequest) (*novelai.CompletionResult, error) {
	return nil, &novelai.UpstreamError{StatusCode: 400, Body: []byte(`{"message":"bad prompt"}`)}
}
func (upstreamFailClient) CompleteStream(context.Context, string, novelai.CompletionRequest) ([]novelai.CompletionChunk, error) {
	return nil, &novelai.UpstreamError{StatusCode: 400, Body: []byte(`{"message":"bad prompt"}`)}
}
func (upstreamFailClient) ListOpenAIModels(context.Context, string) ([]string, error) { return nil, nil }
func (upstreamFailClient) ProbeNativeModel(context.Context, string, string) error      { return nil }

func TestText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/text/completions", strings.NewReader(`{"prompt":"hello"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &TextHandler{Service: &service.TextService{Client: novelai.StaticTextClient{}}}
	h.Completions(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v", body["code"])
	}
}

func TestChatCompletions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/text/chat/completions", strings.NewReader(`{"messages":[{"role":"user","content":"hello"}]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &TextHandler{Service: &service.TextService{Client: novelai.StaticTextClient{}}}
	h.ChatCompletions(c)

	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestTextStream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/text/completions", strings.NewReader(`{"prompt":"hello","stream":true}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &TextHandler{Service: &service.TextService{Client: novelai.StaticTextClient{}}}
	h.Completions(c)

	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Fatalf("content-type = %q", got)
	}
	if !strings.Contains(w.Body.String(), "event:delta") {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestProbeModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/text/models/probe", strings.NewReader(`{"model":"kayra-v1"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &TextHandler{Service: &service.TextService{Client: novelai.StaticTextClient{}}}
	h.ProbeModel(c)

	if w.Code != 200 {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"native_recognized":true`) {
		t.Fatalf("body = %q", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"oa_available":false`) {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestTextCompletionPassesUpstreamStatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/text/completions", strings.NewReader(`{"prompt":"hello"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &TextHandler{Service: &service.TextService{Client: upstreamFailClient{}}}
	h.Completions(c)

	if w.Code != 400 {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
}
