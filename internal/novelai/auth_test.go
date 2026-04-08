package novelai

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/login" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accessToken":"abc"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	got, err := client.Login("key", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "abc" {
		t.Fatalf("AccessToken = %q", got.AccessToken)
	}
}

func TestUserData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/data" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subscription":{"tier":"paper"},"priority":{"level":1},"information":{"email_verified":true}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL, server.URL, server.Client())
	got, err := client.UserData("token")
	if err != nil {
		t.Fatal(err)
	}
	if got.Subscription["tier"] != "paper" {
		t.Fatalf("tier = %v", got.Subscription["tier"])
	}
}
