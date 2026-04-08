package novelai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oa/v1/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["prompt"] != "hello" {
			t.Fatalf("prompt = %v", body["prompt"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"text":"world"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	got, err := client.Complete(t.Context(), "token", CompletionRequest{Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Text != "world" {
		t.Fatalf("text = %q", got.Text)
	}
}

func TestClientCompleteStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oa/v1/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"text\":\"hello\"}]}\n\ndata: [DONE]\n\n"))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	chunks, err := client.CompleteStream(t.Context(), "token", CompletionRequest{Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 1 || chunks[0].Text() != "hello" {
		t.Fatalf("unexpected chunks: %#v", chunks)
	}
}

func TestClientListOpenAIModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oa/v1/models" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"glm-4-6"},{"id":"xialong-v1"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	models, err := client.ListOpenAIModels(t.Context(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0] != "glm-4-6" || models[1] != "xialong-v1" {
		t.Fatalf("models = %#v", models)
	}
}
