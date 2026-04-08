package novelai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientStoriesListObjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/objects/stories" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"objects": []any{},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	result, err := client.ListObjects(t.Context(), "token", "stories")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result["objects"]; !ok {
		t.Fatalf("result = %#v", result)
	}
}

func TestClientStoriesDeleteObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/user/objects/shelf/abc" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	if err := client.DeleteObject(t.Context(), "token", "shelf", "abc"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateObjectType(t *testing.T) {
	if err := ValidateObjectType("stories"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateObjectType("../stories"); err == nil {
		t.Fatal("expected validation error")
	}
}

