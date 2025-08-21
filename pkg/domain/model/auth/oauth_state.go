package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// OAuthState represents a temporary OAuth state for CSRF protection
type OAuthState struct {
	State     string    `json:"state"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// NewOAuthState creates a new OAuth state with a random token
func NewOAuthState() (*OAuthState, error) {
	// Generate a random state token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	now := time.Now()
	return &OAuthState{
		State:     hex.EncodeToString(b),
		ExpiresAt: now.Add(5 * time.Minute), // 5 minutes expiration
		CreatedAt: now,
	}, nil
}

// IsExpired checks if the OAuth state has expired
func (s *OAuthState) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the OAuth state is valid
func (s *OAuthState) IsValid() bool {
	return !s.IsExpired() && s.State != ""
}
