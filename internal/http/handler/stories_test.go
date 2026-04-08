package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"novelai/internal/service"
)

type storiesClientStub struct{}

func (storiesClientStub) GetKeystore(ctx context.Context, token string) (map[string]any, error) {
	return map[string]any{"keystore": "abc"}, nil
}
func (storiesClientStub) PutKeystore(ctx context.Context, token string, payload map[string]any) (map[string]any, error) {
	return payload, nil
}
func (storiesClientStub) GetSubscription(ctx context.Context, token string) (map[string]any, error) {
	return map[string]any{"trainingStepsLeft": map[string]any{"fixedTrainingStepsLeft": 1}}, nil
}
func (storiesClientStub) ListObjects(ctx context.Context, token string, objectType string) (map[string]any, error) {
	return map[string]any{"objects": []any{}}, nil
}
func (storiesClientStub) PutObject(ctx context.Context, token string, objectType string, payload map[string]any) (map[string]any, error) {
	return payload, nil
}
func (storiesClientStub) PatchObject(ctx context.Context, token string, objectType string, id string, payload map[string]any) (map[string]any, error) {
	return payload, nil
}
func (storiesClientStub) DeleteObject(ctx context.Context, token string, objectType string, id string) error {
	return nil
}

func TestStoriesListObjects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/stories/objects/stories", nil)
	c.Params = append(c.Params, gin.Param{Key: "object_type", Value: "stories"})
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &StoriesHandler{Service: &service.StoriesService{Client: storiesClientStub{}}}
	h.ListObjects(c)

	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestStoriesPutObject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/stories/objects/stories", strings.NewReader(`{"data":"x","meta":"y"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = append(c.Params, gin.Param{Key: "object_type", Value: "stories"})
	c.Set("session", &service.Session{AuthToken: "token"})

	h := &StoriesHandler{Service: &service.StoriesService{Client: storiesClientStub{}}}
	h.PutObject(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v", body["code"])
	}
}
