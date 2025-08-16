package types

import (
	"context"

	"github.com/google/uuid"
)

type ThreadID string

func NewThreadID(ctx context.Context) ThreadID {
	return ThreadID(newUUID(ctx))
}

func (id ThreadID) String() string {
	return string(id)
}

// IsValid checks if the ThreadID is valid
func (id ThreadID) IsValid() bool {
	if id == "" {
		return false
	}
	_, err := uuid.Parse(string(id))
	return err == nil
}
