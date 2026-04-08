package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterHealthz(t *testing.T) {
	router := NewRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}
