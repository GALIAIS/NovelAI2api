package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")
var ErrSessionExpired = errors.New("session expired")

type Session struct {
	SessionID     string
	AuthToken     string
	AccessKey     string
	EncryptionKey string
	Email         string
	CreatedAt     time.Time
	ExpiresAt     time.Time
}

type SessionStore interface {
	Create(ctx context.Context, session *Session) error
	Get(ctx context.Context, sessionID string) (*Session, error)
	Delete(ctx context.Context, sessionID string) error
	Refresh(ctx context.Context, sessionID string, ttl time.Duration) error
}

type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{sessions: map[string]*Session{}}
}

func (s *MemorySessionStore) Create(_ context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.SessionID] = session
	return nil
}

func (s *MemorySessionStore) Get(_ context.Context, sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}
	return session, nil
}

func (s *MemorySessionStore) Delete(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	return nil
}

func (s *MemorySessionStore) Refresh(_ context.Context, sessionID string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}
	session.ExpiresAt = time.Now().Add(ttl)
	return nil
}
