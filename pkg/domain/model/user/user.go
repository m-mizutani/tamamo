package user

import (
	"context"
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// User represents a user entity
type User struct {
	ID          types.UserID
	SlackID     string
	SlackName   string
	DisplayName string
	Email       string
	TeamID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewUser creates a new User entity
func NewUser(slackID, slackName, displayName, email, teamID string) *User {
	now := time.Now()
	return &User{
		ID:          types.NewUserID(context.Background()),
		SlackID:     slackID,
		SlackName:   slackName,
		DisplayName: displayName,
		Email:       email,
		TeamID:      teamID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// UpdateSlackInfo updates the user's Slack information
func (u *User) UpdateSlackInfo(slackName, displayName, email string) {
	u.SlackName = slackName
	u.DisplayName = displayName
	u.Email = email
	u.UpdatedAt = time.Now()
}

// IsUpdateRequired checks if the user information needs to be updated
func (u *User) IsUpdateRequired(updateInterval time.Duration) bool {
	return time.Since(u.UpdatedAt) > updateInterval
}
