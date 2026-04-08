package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	httpapi "novelai/internal/httpapi"
	"novelai/internal/service"
)

func Session(store service.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := strings.TrimSpace(c.GetHeader("X-Session-ID"))
		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		apiToken := strings.TrimSpace(c.GetHeader("X-API-Token"))
		if strings.HasPrefix(auth, "Bearer ") {
			bearer := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			if strings.HasPrefix(bearer, "sess_") {
				sessionID = bearer
			} else if bearer != "" {
				apiToken = bearer
			}
		}
		if sessionID == "" && apiToken == "" {
			httpapi.WriteError(c, 401, 40101, "missing session", "unauthorized", nil)
			c.Abort()
			return
		}
		if sessionID != "" {
			session, err := store.Get(c.Request.Context(), sessionID)
			if err != nil {
				httpapi.WriteError(c, 401, 40102, "invalid session", "session_expired", nil)
				c.Abort()
				return
			}
			c.Set("session", session)
			c.Next()
			return
		}
		c.Set("session", &service.Session{AuthToken: apiToken})
		c.Next()
	}
}
