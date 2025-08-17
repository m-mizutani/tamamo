package types

import (
	"context"

	"github.com/google/uuid"
)

type HistoryID string

func NewHistoryID(ctx context.Context) HistoryID {
	return HistoryID(newUUID(ctx))
}

func (id HistoryID) String() string {
	return string(id)
}

// IsValid checks if the HistoryID is valid
func (id HistoryID) IsValid() bool {
	if id == "" {
		return false
	}
	_, err := uuid.Parse(string(id))
	return err == nil
}
