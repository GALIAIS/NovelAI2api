package service

import (
	"context"
	"testing"
	"time"
)

func TestMemorySessionStoreCreateAndGet(t *testing.T) {
	store := NewMemorySessionStore()
	session := &Session{SessionID: "sess_1", AuthToken: "token", ExpiresAt: time.Now().Add(time.Hour)}

	if err := store.Create(context.Background(), session); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(context.Background(), "sess_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.AuthToken != "token" {
		t.Fatalf("AuthToken = %q", got.AuthToken)
	}
}
