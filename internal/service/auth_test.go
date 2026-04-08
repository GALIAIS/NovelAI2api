package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"novelai/internal/model"
	"novelai/internal/novelai"
)

type authClientStub struct{}

func (authClientStub) Login(accessKey, captchaToken string) (*novelai.LoginResult, error) {
	return &novelai.LoginResult{AccessToken: "token"}, nil
}

func (authClientStub) UserData(token string) (*novelai.UserDataResult, error) {
	return &novelai.UserDataResult{
		Subscription: map[string]any{"tier": "paper"},
		Priority:     map[string]any{"level": 1},
		Information:  map[string]any{"email_verified": true, "email": "user@example.com"},
	}, nil
}

func TestAuthService(t *testing.T) {
	store := NewMemorySessionStore()
	svc := &AuthService{
		Store:  store,
		Client: authClientStub{},
		TTL:    time.Hour,
	}
	resp, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected session id")
	}
}

func TestAuthServiceLoginWithAPIToken(t *testing.T) {
	store := NewMemorySessionStore()
	svc := &AuthService{
		Store:  store,
		Client: authClientStub{},
		TTL:    time.Hour,
	}
	resp, err := svc.Login(context.Background(), model.LoginRequest{
		APIToken: "token-from-api",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected session id")
	}
	if resp.User.Email != "user@example.com" {
		t.Fatalf("email = %q", resp.User.Email)
	}
	session, err := store.Get(context.Background(), resp.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if session.AuthToken != "token-from-api" {
		t.Fatalf("auth token = %q", session.AuthToken)
	}
}

func TestAuthServiceRejectsEmptyLogin(t *testing.T) {
	store := NewMemorySessionStore()
	svc := &AuthService{
		Store:  store,
		Client: authClientStub{},
		TTL:    time.Hour,
	}
	if _, err := svc.Login(context.Background(), model.LoginRequest{}); !errors.Is(err, ErrInvalidLoginRequest) {
		t.Fatalf("err = %v", err)
	}
}

func TestAuthServiceMeFromSessionWithDirectToken(t *testing.T) {
	svc := &AuthService{
		Store:  NewMemorySessionStore(),
		Client: authClientStub{},
		TTL:    time.Hour,
	}
	resp, err := svc.MeFromSession(context.Background(), &Session{
		AuthToken: "token-from-api",
		Email:     "user@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Email != "user@example.com" {
		t.Fatalf("email = %q", resp.Email)
	}
}
