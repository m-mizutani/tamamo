package types

import (
	"context"

	"github.com/google/uuid"
)

// UserID represents a unique identifier for a user
type UserID string

// Special UserID constants
const (
	AnonymousUserID UserID = "anonymous"
)

// NewUserID creates a new UserID with a UUID
func NewUserID(ctx context.Context) UserID {
	return UserID(uuid.New().String())
}

// String returns the string representation of UserID
func (id UserID) String() string {
	return string(id)
}

// IsValid checks if the UserID is a valid UUID or a special constant
func (id UserID) IsValid() bool {
	// Allow special constants
	if id == AnonymousUserID {
		return true
	}
	// Check if it's a valid UUID
	_, err := uuid.Parse(string(id))
	return err == nil
}
