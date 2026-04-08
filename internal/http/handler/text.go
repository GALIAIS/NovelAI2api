package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	httpapi "novelai/internal/httpapi"
	"novelai/internal/model"
	"novelai/internal/service"
)

type TextHandler struct {
	Service *service.TextService
}

func (h *TextHandler) Completions(c *gin.Context) {
	var req model.CompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40011, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	session := c.MustGet("session").(*service.Session)
	if req.Stream {
		chunks, err := h.Service.CompleteStream(c.Request.Context(), session.AuthToken, req)
		if err != nil {
			writeEnvelopeUpstreamAwareError(c, 500, 50013, "completion stream failed", "internal_error", err)
			return
		}
		writeSSEHeaders(c)
		c.Status(http.StatusOK)
		for _, chunk := range chunks {
			c.SSEvent("delta", gin.H{"text": chunk.Text()})
			c.Writer.Flush()
		}
		c.SSEvent("done", gin.H{"done": true})
		c.Writer.Flush()
		return
	}
	resp, err := h.Service.Complete(c.Request.Context(), session.AuthToken, req)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 500, 50011, "completion failed", "internal_error", err)
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}

func (h *TextHandler) ChatCompletions(c *gin.Context) {
	var req model.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40012, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	session := c.MustGet("session").(*service.Session)
	if req.Stream {
		chunks, err := h.Service.ChatStream(c.Request.Context(), session.AuthToken, req)
		if err != nil {
			writeEnvelopeUpstreamAwareError(c, 500, 50014, "chat completion stream failed", "internal_error", err)
			return
		}
		writeSSEHeaders(c)
		c.Status(http.StatusOK)
		for _, chunk := range chunks {
			c.SSEvent("delta", gin.H{"text": chunk.Text()})
			c.Writer.Flush()
		}
		c.SSEvent("done", gin.H{"done": true})
		c.Writer.Flush()
		return
	}
	resp, err := h.Service.Chat(c.Request.Context(), session.AuthToken, req)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 500, 50012, "chat completion failed", "internal_error", err)
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}

func (h *TextHandler) ProbeModel(c *gin.Context) {
	var req model.ModelProbeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40013, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	session := c.MustGet("session").(*service.Session)
	resp, err := h.Service.ProbeModel(c.Request.Context(), session.AuthToken, req)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 500, 50015, "model probe failed", "internal_error", err)
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}

func writeSSEHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
}
