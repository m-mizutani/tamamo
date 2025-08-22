package auth

import (
	"context"
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// Session represents an authenticated user session
type Session struct {
	ID        types.UUID   `json:"id" firestore:"id"`
	UserID    types.UserID `json:"user_id" firestore:"user_id"`
	UserName  string       `json:"user_name" firestore:"user_name"`
	Email     string       `json:"email" firestore:"email"`
	TeamID    string       `json:"team_id" firestore:"team_id"`
	TeamName  string       `json:"team_name" firestore:"team_name"`
	ExpiresAt time.Time    `json:"expires_at" firestore:"expires_at"`
	CreatedAt time.Time    `json:"created_at" firestore:"created_at"`
}

// NewSession creates a new session with default expiration
func NewSession(ctx context.Context, userID types.UserID, userName, email, teamID, teamName string) *Session {
	now := time.Now()
	return &Session{
		ID:        types.NewUUID(ctx),
		UserID:    userID,
		UserName:  userName,
		Email:     email,
		TeamID:    teamID,
		TeamName:  teamName,
		ExpiresAt: now.Add(7 * 24 * time.Hour), // 7 days
		CreatedAt: now,
	}
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired and has required fields)
func (s *Session) IsValid() bool {
	return !s.IsExpired() && s.ID.IsValid() && s.UserID.IsValid() && s.TeamID != ""
}
