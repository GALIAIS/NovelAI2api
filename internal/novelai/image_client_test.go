package novelai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"novelai/internal/testutil"
)

func TestClientGenerate(t *testing.T) {
	stream := testutil.BuildImageStreamFixture(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ai/generate-image-stream" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("content-type = %q", r.Header.Get("Content-Type"))
		}
		if _, err := io.Copy(w, stream); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	got, err := client.Generate(t.Context(), "token", ImageGenerateRequest{Prompt: "cat", Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Images) != 1 {
		t.Fatalf("len = %d", len(got.Images))
	}
}

func TestClientEncodeVibe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ai/encode-vibe" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("vibe"))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	got, err := client.EncodeVibe(t.Context(), "token", EncodeVibeRequest{Image: []byte("img")})
	if err != nil {
		t.Fatal(err)
	}
	if got.VibeCode == "" {
		t.Fatal("expected vibe code")
	}
}

func TestClientGenerateStream(t *testing.T) {
	stream := testutil.BuildImageStreamFixture(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ai/generate-image-stream" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if _, err := io.Copy(w, stream); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	events, err := client.GenerateStream(t.Context(), "token", ImageGenerateRequest{Prompt: "cat"})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].EventType != "final" {
		t.Fatalf("unexpected events: %#v", events)
	}
}

func TestClientGenerateStreamMultipartWithFiles(t *testing.T) {
	stream := testutil.BuildImageStreamFixture(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ai/generate-image-stream" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("content-type = %q", r.Header.Get("Content-Type"))
		}
		if _, err := io.Copy(w, stream); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	events, err := client.GenerateStream(t.Context(), "token", ImageGenerateRequest{
		Prompt: "cat",
		Model:  "nai-diffusion-4-5-curated",
		Files:  map[string][]byte{"image": []byte("png")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("len = %d", len(events))
	}
}

func TestBuildImageGeneratePayloadRespectsParameters(t *testing.T) {
	flag := false
	payload := buildImageGeneratePayload(ImageGenerateRequest{
		Prompt: "cat",
		Model:  "nai-diffusion-4-5-curated",
		Action: "img2img",
		Parameters: map[string]any{
			"width":  1536,
			"height": 1536,
		},
		UseNewSharedTrial: &flag,
	})
	if payload["action"] != "img2img" {
		t.Fatalf("action = %v", payload["action"])
	}
	if payload["use_new_shared_trial"] != false {
		t.Fatalf("trial = %v", payload["use_new_shared_trial"])
	}
	params := payload["parameters"].(map[string]any)
	if params["width"] != 1536 {
		t.Fatalf("width = %v", params["width"])
	}
	if params["stream"] != "msgpack" {
		t.Fatalf("stream = %v", params["stream"])
	}
	if params["params_version"] != 3 {
		t.Fatalf("params_version = %v", params["params_version"])
	}
}

func TestBuildImageGeneratePayloadHonorsRawRequest(t *testing.T) {
	payload := buildImageGeneratePayload(ImageGenerateRequest{
		Prompt: "cat",
		Model:  "nai-diffusion-4-5-curated",
		Parameters: map[string]any{
			"width": 1536,
		},
		RawRequest: map[string]any{
			"input": "raw-input",
			"parameters": map[string]any{
				"width":  2048,
				"height": 1024,
			},
			"use_new_shared_trial": false,
		},
	})
	if payload["input"] != "raw-input" {
		t.Fatalf("input = %v", payload["input"])
	}
	if payload["use_new_shared_trial"] != false {
		t.Fatalf("trial = %v", payload["use_new_shared_trial"])
	}
	params := payload["parameters"].(map[string]any)
	if params["width"] != 2048 {
		t.Fatalf("width = %v", params["width"])
	}
	if params["height"] != 1024 {
		t.Fatalf("height = %v", params["height"])
	}
}
