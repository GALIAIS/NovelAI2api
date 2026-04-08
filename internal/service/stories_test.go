package service

import (
	"context"
	"testing"
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

func TestStoriesServiceListObjects(t *testing.T) {
	svc := &StoriesService{Client: storiesClientStub{}}
	resp, err := svc.ListObjects(t.Context(), "token", "stories")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := resp["objects"]; !ok {
		t.Fatalf("resp = %#v", resp)
	}
}

func TestStoriesServiceRejectsInvalidObjectType(t *testing.T) {
	svc := &StoriesService{Client: storiesClientStub{}}
	if _, err := svc.ListObjects(t.Context(), "token", "../stories"); err == nil {
		t.Fatal("expected error")
	}
}

