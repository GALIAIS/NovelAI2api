package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"novelai/internal/service"
)

func TestSessionAcceptsStoredSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := service.NewMemorySessionStore()
	if err := store.Create(t.Context(), &service.Session{
		SessionID: "sess_123",
		AuthToken: "token",
		ExpiresAt: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.Use(Session(store))
	r.GET("/protected", func(c *gin.Context) {
		session := c.MustGet("session").(*service.Session)
		c.JSON(http.StatusOK, gin.H{"auth_token": session.AuthToken})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer sess_123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestSessionAcceptsDirectAPIToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Session(service.NewMemorySessionStore()))
	r.GET("/protected", func(c *gin.Context) {
		session := c.MustGet("session").(*service.Session)
		c.JSON(http.StatusOK, gin.H{"auth_token": session.AuthToken})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer pst-test-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["auth_token"] != "pst-test-token" {
		t.Fatalf("auth token = %v", body["auth_token"])
	}
}

func TestSessionRejectsMissingCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Session(service.NewMemorySessionStore()))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}
