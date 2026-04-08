package handler

import (
	"github.com/gin-gonic/gin"
	httpapi "novelai/internal/httpapi"
	"novelai/internal/model"
	"novelai/internal/service"
)

type TokenizerHandler struct {
	Service *service.TokenizerService
}

func (h *TokenizerHandler) Encode(c *gin.Context) {
	var req model.TokenizerEncodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40001, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	resp, err := h.Service.Encode(req)
	if err != nil {
		httpapi.WriteError(c, 400, 40002, "tokenizer encode failed", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}

func (h *TokenizerHandler) Decode(c *gin.Context) {
	var req model.TokenizerDecodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40003, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	resp, err := h.Service.Decode(req)
	if err != nil {
		httpapi.WriteError(c, 400, 40004, "tokenizer decode failed", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}
