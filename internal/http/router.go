package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"novelai/internal/http/handler"
	"novelai/internal/http/middleware"
	httpapi "novelai/internal/httpapi"
	"novelai/internal/service"
)

type Dependencies struct {
	SessionStore     service.SessionStore
	AuthHandler      *handler.AuthHandler
	TextHandler      *handler.TextHandler
	TokenizerHandler *handler.TokenizerHandler
	ImageHandler     *handler.ImageHandler
	StoriesHandler   *handler.StoriesHandler
	OpenAIHandler    *handler.OpenAIHandler
}

func NewRouter(deps *Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), middleware.Recovery())
	r.GET("/healthz", func(c *gin.Context) { httpapi.WriteSuccess(c, http.StatusOK, gin.H{"status": "ok"}) })

	api := r.Group("/api")
	if deps != nil {
		api.POST("/auth/login", deps.AuthHandler.Login)
		api.GET("/auth/me", middleware.Session(deps.SessionStore), deps.AuthHandler.Me)
		api.POST("/tokenizer/encode", deps.TokenizerHandler.Encode)
		api.POST("/tokenizer/decode", deps.TokenizerHandler.Decode)
		api.POST("/text/completions", middleware.Session(deps.SessionStore), deps.TextHandler.Completions)
		api.POST("/text/chat/completions", middleware.Session(deps.SessionStore), deps.TextHandler.ChatCompletions)
		api.POST("/text/models/probe", middleware.Session(deps.SessionStore), deps.TextHandler.ProbeModel)
		api.POST("/image/generate", middleware.Session(deps.SessionStore), deps.ImageHandler.Generate)
		api.POST("/image/director-tools", middleware.Session(deps.SessionStore), deps.ImageHandler.DirectorTools)
		api.POST("/image/encode-vibe", middleware.Session(deps.SessionStore), deps.ImageHandler.EncodeVibe)
		api.GET("/stories/keystore", middleware.Session(deps.SessionStore), deps.StoriesHandler.GetKeystore)
		api.PUT("/stories/keystore", middleware.Session(deps.SessionStore), deps.StoriesHandler.PutKeystore)
		api.GET("/account/subscription", middleware.Session(deps.SessionStore), deps.StoriesHandler.GetSubscription)
		api.GET("/stories/objects/:object_type", middleware.Session(deps.SessionStore), deps.StoriesHandler.ListObjects)
		api.PUT("/stories/objects/:object_type", middleware.Session(deps.SessionStore), deps.StoriesHandler.PutObject)
		api.PATCH("/stories/objects/:object_type/:id", middleware.Session(deps.SessionStore), deps.StoriesHandler.PatchObject)
		api.DELETE("/stories/objects/:object_type/:id", middleware.Session(deps.SessionStore), deps.StoriesHandler.DeleteObject)
	}
	if deps != nil && deps.OpenAIHandler != nil {
		v1 := r.Group("/v1")
		v1.Use(middleware.Session(deps.SessionStore))
		v1.GET("/models", deps.OpenAIHandler.ListModels)
		v1.POST("/completions", deps.OpenAIHandler.Completions)
		v1.POST("/chat/completions", deps.OpenAIHandler.ChatCompletions)
		v1.POST("/responses", deps.OpenAIHandler.Responses)
		v1.POST("/images/generations", deps.OpenAIHandler.ImageGenerations)
	}
	return r
}
