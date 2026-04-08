package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"novelai/internal/service"
)

func TestImageGenerate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/image/generate", strings.NewReader(`{"prompt":"cat"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &ImageHandler{Service: &service.ImageService{}}
	h.Generate(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v", body["code"])
	}
}

func TestDirectorTools(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/image/director-tools", strings.NewReader(`{"tool":"remove_bg","image":"abc"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &ImageHandler{Service: &service.ImageService{}}
	h.DirectorTools(c)

	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestEncodeVibe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/image/encode-vibe", strings.NewReader(`{"image":"abc"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &ImageHandler{Service: &service.ImageService{}}
	h.EncodeVibe(c)

	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestImageGenerateStream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/image/generate", strings.NewReader(`{"prompt":"cat","stream":true}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &ImageHandler{Service: &service.ImageService{}}
	h.Generate(c)

	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Fatalf("content-type = %q", got)
	}
	if !strings.Contains(w.Body.String(), "event:final") {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestImageGenerateAcceptsCustomLargeValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/image/generate", strings.NewReader(`{"prompt":"cat","width":2048,"height":2048,"steps":40,"n_samples":2}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &ImageHandler{Service: &service.ImageService{}}
	h.Generate(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v", body["code"])
	}
}
