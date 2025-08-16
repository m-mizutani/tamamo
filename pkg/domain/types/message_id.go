package types

import (
	"context"

	"github.com/google/uuid"
)

type MessageID string

func NewMessageID(ctx context.Context) MessageID {
	return MessageID(newUUID(ctx))
}

func (id MessageID) String() string {
	return string(id)
}

// IsValid checks if the MessageID is valid
func (id MessageID) IsValid() bool {
	if id == "" {
		return false
	}
	_, err := uuid.Parse(string(id))
	return err == nil
}
