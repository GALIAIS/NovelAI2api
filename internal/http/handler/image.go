package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	httpapi "novelai/internal/httpapi"
	"novelai/internal/model"
	"novelai/internal/service"
)

type ImageHandler struct {
	Service *service.ImageService
}

func (h *ImageHandler) Generate(c *gin.Context) {
	var req model.ImageGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40021, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	session := c.MustGet("session").(*service.Session)
	if req.Stream {
		events, err := h.Service.GenerateStream(c.Request.Context(), session.AuthToken, req)
		if err != nil {
			writeEnvelopeUpstreamAwareError(c, 400, 40027, "image generation stream failed", "bad_request", err)
			return
		}
		writeSSEHeaders(c)
		c.Status(http.StatusOK)
		for _, event := range events {
			c.SSEvent(event.EventType, gin.H{
				"image":   base64.StdEncoding.EncodeToString(event.Image),
				"samp_ix": event.SampIX,
				"step_ix": event.StepIX,
			})
			c.Writer.Flush()
		}
		c.SSEvent("done", gin.H{"done": true})
		c.Writer.Flush()
		return
	}
	resp, err := h.Service.Generate(c.Request.Context(), session.AuthToken, req)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40022, "image generation failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}

func (h *ImageHandler) DirectorTools(c *gin.Context) {
	var req model.DirectorToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40023, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	session := c.MustGet("session").(*service.Session)
	resp, err := h.Service.DirectorTool(c.Request.Context(), session.AuthToken, req)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40024, "director tool failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}

func (h *ImageHandler) EncodeVibe(c *gin.Context) {
	var req model.EncodeVibeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40025, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	session := c.MustGet("session").(*service.Session)
	resp, err := h.Service.EncodeVibe(c.Request.Context(), session.AuthToken, req)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40026, "encode vibe failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, 200, resp)
}
