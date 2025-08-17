package slack

import (
	"context"
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// History represents a conversation history storage record
type History struct {
	ID        types.HistoryID `json:"id"`
	ThreadID  types.ThreadID  `json:"thread_id"`
	CreatedAt time.Time       `json:"created_at"`
}

// NewHistory creates a new History instance
func NewHistory(ctx context.Context, threadID types.ThreadID) *History {
	return &History{
		ID:        types.NewHistoryID(ctx),
		ThreadID:  threadID,
		CreatedAt: time.Now(),
	}
}

// Validate checks if the history has valid fields
func (h *History) Validate() error {
	if !h.ID.IsValid() {
		return ErrInvalidHistoryID
	}
	if !h.ThreadID.IsValid() {
		return ErrInvalidThreadID
	}
	return nil
}
