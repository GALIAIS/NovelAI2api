package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"novelai/internal/novelai"
	"novelai/internal/service"
)

type authClientTestStub struct{}

func (authClientTestStub) Login(accessKey, captchaToken string) (*novelai.LoginResult, error) {
	return &novelai.LoginResult{AccessToken: "token"}, nil
}

func (authClientTestStub) UserData(token string) (*novelai.UserDataResult, error) {
	return &novelai.UserDataResult{
		Subscription: map[string]any{"tier": "paper"},
		Information:  map[string]any{"email": "user@example.com"},
	}, nil
}

func TestAuthLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"user@example.com","password":"password123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &AuthHandler{
		Service: &service.AuthService{
			Store:  service.NewMemorySessionStore(),
			Client: authClientTestStub{},
			TTL:    time.Hour,
		},
	}
	h.Login(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v", body["code"])
	}
}

func TestAuthLoginWithAPIToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"api_token":"token-from-api"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &AuthHandler{
		Service: &service.AuthService{
			Store:  service.NewMemorySessionStore(),
			Client: authClientTestStub{},
			TTL:    time.Hour,
		},
	}
	h.Login(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v", body["code"])
	}
}

func TestAuthLoginRejectsEmptyPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &AuthHandler{
		Service: &service.AuthService{
			Store:  service.NewMemorySessionStore(),
			Client: authClientTestStub{},
			TTL:    time.Hour,
		},
	}
	h.Login(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 40002 {
		t.Fatalf("code = %v", body["code"])
	}
}

func TestAuthMeWithDirectAPITokenSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/auth/me", nil)
	c.Set("session", &service.Session{AuthToken: "token-from-api", Email: "user@example.com"})

	h := &AuthHandler{
		Service: &service.AuthService{
			Store:  service.NewMemorySessionStore(),
			Client: authClientTestStub{},
			TTL:    time.Hour,
		},
	}
	h.Me(c)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v", body["code"])
	}
}
