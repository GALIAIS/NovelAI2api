package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"novelai/internal/model"
	"novelai/internal/novelai"
)

var ErrInvalidLoginRequest = errors.New("invalid login request")

type AuthService struct {
	Store  SessionStore
	Client AuthClient
	TTL    time.Duration
}

type AuthClient interface {
	Login(accessKey, captchaToken string) (*novelai.LoginResult, error)
	UserData(token string) (*novelai.UserDataResult, error)
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	if strings.TrimSpace(req.APIToken) != "" {
		return s.loginWithAPIToken(ctx, req)
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrInvalidLoginRequest
	}

	accessKey, encryptionKey, err := novelai.DeriveKeys(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	loginResult, err := s.Client.Login(accessKey, req.CaptchaToken)
	if err != nil {
		return nil, err
	}
	userData, err := s.Client.UserData(loginResult.AccessToken)
	if err != nil {
		return nil, err
	}
	sessionID := "sess_" + uuid.NewString()
	if err := s.Store.Create(ctx, &Session{
		SessionID:     sessionID,
		AuthToken:     loginResult.AccessToken,
		AccessKey:     accessKey,
		EncryptionKey: encryptionKey,
		Email:         req.Email,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(s.TTL),
	}); err != nil {
		return nil, err
	}
	return &model.LoginResponse{
		SessionID: sessionID,
		User: model.UserSummary{
			Email: req.Email,
		},
		Subscription: userData.Subscription,
		Priority:     userData.Priority,
		Information:  userData.Information,
	}, nil
}

func (s *AuthService) loginWithAPIToken(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	token := strings.TrimSpace(req.APIToken)
	if token == "" {
		return nil, ErrInvalidLoginRequest
	}
	userData, err := s.Client.UserData(token)
	if err != nil {
		return nil, err
	}

	email := extractEmail(userData.Information)
	sessionID := "sess_" + uuid.NewString()
	if err := s.Store.Create(ctx, &Session{
		SessionID: sessionID,
		AuthToken: token,
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.TTL),
	}); err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		SessionID: sessionID,
		User: model.UserSummary{
			Email: email,
		},
		Subscription: userData.Subscription,
		Priority:     userData.Priority,
		Information:  userData.Information,
	}, nil
}

func (s *AuthService) Me(ctx context.Context, sessionID string) (*model.MeResponse, error) {
	session, err := s.Store.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return s.MeFromSession(ctx, session)
}

func (s *AuthService) MeFromSession(ctx context.Context, session *Session) (*model.MeResponse, error) {
	userData, err := s.Client.UserData(session.AuthToken)
	if err != nil {
		return nil, err
	}
	return &model.MeResponse{
		Email:        session.Email,
		Subscription: userData.Subscription,
		Priority:     userData.Priority,
		Information:  userData.Information,
	}, nil
}

func extractEmail(information map[string]any) string {
	if information == nil {
		return ""
	}
	if raw, ok := information["email"]; ok {
		if email, ok := raw.(string); ok {
			return email
		}
	}
	return ""
}
