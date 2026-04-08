package app

import (
	"net/http"

	"novelai/internal/config"
	apphttp "novelai/internal/http"
	"novelai/internal/http/handler"
	"novelai/internal/novelai"
	"novelai/internal/service"
)

func NewRouter(cfg config.Config) http.Handler {
	client := novelai.NewClient(cfg.NovelAIAPIBase, cfg.NovelAIImageBase, cfg.NovelAITextBase, nil)
	sessionStore := service.NewMemorySessionStore()

	authHandler := &handler.AuthHandler{
		Service: &service.AuthService{
			Store:  sessionStore,
			Client: client,
			TTL:    cfg.SessionTTL,
		},
	}
	textHandler := &handler.TextHandler{
		Service: &service.TextService{Client: client},
	}
	tokenizerHandler := &handler.TokenizerHandler{
		Service: &service.TokenizerService{Tokenizer: novelai.NewLocalTokenizer()},
	}
	imageHandler := &handler.ImageHandler{
		Service: &service.ImageService{Client: client},
	}
	storiesHandler := &handler.StoriesHandler{
		Service: &service.StoriesService{Client: client},
	}
	openAIHandler := &handler.OpenAIHandler{
		TextService:  &service.TextService{Client: client},
		ImageService: &service.ImageService{Client: client},
	}

	return apphttp.NewRouter(&apphttp.Dependencies{
		SessionStore:     sessionStore,
		AuthHandler:      authHandler,
		TextHandler:      textHandler,
		TokenizerHandler: tokenizerHandler,
		ImageHandler:     imageHandler,
		StoriesHandler:   storiesHandler,
		OpenAIHandler:    openAIHandler,
	})
}
