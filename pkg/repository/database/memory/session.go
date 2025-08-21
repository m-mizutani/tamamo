package memory

import (
	"context"
	"sync"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
)

type sessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*auth.Session
}

// NewSessionRepository creates a new in-memory session repository
func NewSessionRepository() interfaces.SessionRepository {
	return &sessionRepository{
		sessions: make(map[string]*auth.Session),
	}
}

func (r *sessionRepository) CreateSession(ctx context.Context, session *auth.Session) error {
	if session == nil {
		return goerr.New("session is nil")
	}
	if !session.ID.IsValid() {
		return goerr.New("invalid session ID")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	sessionID := session.ID.String()
	if _, exists := r.sessions[sessionID]; exists {
		return goerr.New("session already exists", goerr.V("sessionID", sessionID))
	}

	r.sessions[sessionID] = session
	return nil
}

func (r *sessionRepository) GetSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, auth.ErrSessionNotFound
	}

	// Check if session is expired
	if session.IsExpired() {
		return nil, auth.ErrSessionExpired
	}

	return session, nil
}

func (r *sessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[sessionID]; !exists {
		return auth.ErrSessionNotFound
	}

	delete(r.sessions, sessionID)
	return nil
}

func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, session := range r.sessions {
		if now.After(session.ExpiresAt) {
			delete(r.sessions, id)
		}
	}

	return nil
}
