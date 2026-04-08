package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	httpapi "novelai/internal/httpapi"
	"novelai/internal/model"
	"novelai/internal/service"
)

type AuthHandler struct {
	Service *service.AuthService
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40001, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	resp, err := h.Service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLoginRequest) {
			httpapi.WriteError(c, 400, 40002, "either email/password or api_token is required", "bad_request", map[string]any{"error": err.Error()})
			return
		}
		writeEnvelopeUpstreamAwareError(c, 500, 50001, "login failed", "internal_error", err)
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	raw, ok := c.Get("session")
	if !ok {
		httpapi.WriteError(c, 401, 40103, "missing session in context", "unauthorized", nil)
		return
	}
	session := raw.(*service.Session)
	resp, err := h.Service.MeFromSession(c.Request.Context(), session)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 500, 50002, "failed to load user", "internal_error", err)
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}
