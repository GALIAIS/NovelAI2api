package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	httpapi "novelai/internal/httpapi"
	"novelai/internal/service"
)

type StoriesHandler struct {
	Service *service.StoriesService
}

func (h *StoriesHandler) GetKeystore(c *gin.Context) {
	session := c.MustGet("session").(*service.Session)
	resp, err := h.Service.GetKeystore(c.Request.Context(), session.AuthToken)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40041, "get keystore failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, http.StatusOK, resp)
}

func (h *StoriesHandler) PutKeystore(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40042, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	session := c.MustGet("session").(*service.Session)
	resp, err := h.Service.PutKeystore(c.Request.Context(), session.AuthToken, req)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40043, "put keystore failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, http.StatusOK, resp)
}

func (h *StoriesHandler) GetSubscription(c *gin.Context) {
	session := c.MustGet("session").(*service.Session)
	resp, err := h.Service.GetSubscription(c.Request.Context(), session.AuthToken)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40044, "get subscription failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, http.StatusOK, resp)
}

func (h *StoriesHandler) ListObjects(c *gin.Context) {
	session := c.MustGet("session").(*service.Session)
	objectType := c.Param("object_type")
	resp, err := h.Service.ListObjects(c.Request.Context(), session.AuthToken, objectType)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40045, "list objects failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, http.StatusOK, resp)
}

func (h *StoriesHandler) PutObject(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40046, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	session := c.MustGet("session").(*service.Session)
	objectType := c.Param("object_type")
	resp, err := h.Service.PutObject(c.Request.Context(), session.AuthToken, objectType, req)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40047, "put object failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, http.StatusOK, resp)
}

func (h *StoriesHandler) PatchObject(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.WriteError(c, 400, 40048, "invalid request", "bad_request", map[string]any{"error": err.Error()})
		return
	}
	session := c.MustGet("session").(*service.Session)
	objectType := c.Param("object_type")
	id := c.Param("id")
	resp, err := h.Service.PatchObject(c.Request.Context(), session.AuthToken, objectType, id, req)
	if err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40049, "patch object failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, http.StatusOK, resp)
}

func (h *StoriesHandler) DeleteObject(c *gin.Context) {
	session := c.MustGet("session").(*service.Session)
	objectType := c.Param("object_type")
	id := c.Param("id")
	if err := h.Service.DeleteObject(c.Request.Context(), session.AuthToken, objectType, id); err != nil {
		writeEnvelopeUpstreamAwareError(c, 400, 40050, "delete object failed", "bad_request", err)
		return
	}
	httpapi.WriteSuccess(c, http.StatusOK, gin.H{"deleted": true})
}
