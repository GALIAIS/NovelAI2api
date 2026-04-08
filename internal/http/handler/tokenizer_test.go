package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"novelai/internal/novelai"
	"novelai/internal/service"
)

func TestTokenizerEncode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/tokenizer/encode", strings.NewReader(`{"text":"hi","tokenizer":"gpt2"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &TokenizerHandler{
		Service: &service.TokenizerService{Tokenizer: novelai.NewLocalTokenizer()},
	}
	h.Encode(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v", body["code"])
	}
}

func TestTokenizerEncodeSupportsFrontendAlias(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/tokenizer/encode", strings.NewReader(`{"text":"hi","tokenizer":"pile-nai"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &TokenizerHandler{
		Service: &service.TokenizerService{Tokenizer: novelai.NewLocalTokenizer()},
	}
	h.Encode(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v", body["code"])
	}
}
