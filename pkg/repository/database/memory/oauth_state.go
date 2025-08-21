package memory

import (
	"context"
	"sync"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
)

type oauthStateRepository struct {
	mu     sync.RWMutex
	states map[string]*auth.OAuthState
}

// NewOAuthStateRepository creates a new in-memory OAuth state repository
func NewOAuthStateRepository() interfaces.OAuthStateRepository {
	return &oauthStateRepository{
		states: make(map[string]*auth.OAuthState),
	}
}

func (r *oauthStateRepository) SaveState(ctx context.Context, state *auth.OAuthState) error {
	if state == nil {
		return goerr.New("state is nil")
	}
	if state.State == "" {
		return goerr.New("state token is empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if state already exists
	if _, exists := r.states[state.State]; exists {
		return goerr.New("state already exists", goerr.V("state", state.State))
	}

	r.states[state.State] = state

	// Clean up expired states while we have the lock
	r.cleanupExpiredStatesLocked()

	return nil
}

func (r *oauthStateRepository) GetState(ctx context.Context, stateToken string) (*auth.OAuthState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, exists := r.states[stateToken]
	if !exists {
		return nil, auth.ErrStateNotFound
	}

	// Check if state is expired
	if state.IsExpired() {
		return nil, auth.ErrStateExpired
	}

	return state, nil
}

func (r *oauthStateRepository) ValidateAndDeleteState(ctx context.Context, stateToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, exists := r.states[stateToken]
	if !exists {
		return auth.ErrStateNotFound
	}

	// Check if state is expired
	if state.IsExpired() {
		delete(r.states, stateToken)
		return auth.ErrStateExpired
	}

	// Delete the state after successful validation
	delete(r.states, stateToken)
	return nil
}

// cleanupExpiredStatesLocked removes expired states (must be called with lock held)
func (r *oauthStateRepository) cleanupExpiredStatesLocked() {
	now := time.Now()
	for token, state := range r.states {
		if now.After(state.ExpiresAt) {
			delete(r.states, token)
		}
	}
}
