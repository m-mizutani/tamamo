package slack

import (
	"context"
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// ThreadMessage represents a message within a thread (for persistence)
type ThreadMessage struct {
	ID        types.MessageID `json:"id"`
	ThreadID  types.ThreadID  `json:"thread_id"`
	UserID    string          `json:"user_id"`
	UserName  string          `json:"user_name"`
	Text      string          `json:"text"`
	Timestamp string          `json:"timestamp"`
	CreatedAt time.Time       `json:"created_at"`
}

// NewThreadMessage creates a new ThreadMessage instance
func NewThreadMessage(threadID types.ThreadID, userID, userName, text, timestamp string) *ThreadMessage {
	return &ThreadMessage{
		ID:        types.NewMessageID(context.Background()),
		ThreadID:  threadID,
		UserID:    userID,
		UserName:  userName,
		Text:      text,
		Timestamp: timestamp,
		CreatedAt: time.Now(),
	}
}

// Validate checks if the message has valid fields
func (m *ThreadMessage) Validate() error {
	if m.ID == "" {
		return ErrEmptyMessageID
	}
	if m.ID != "" && !m.ID.IsValid() {
		return ErrInvalidMessageID
	}
	if !m.ThreadID.IsValid() {
		return ErrInvalidThreadID
	}
	if m.UserID == "" {
		return ErrEmptyUserID
	}
	if m.Text == "" {
		return ErrEmptyText
	}
	if m.Timestamp == "" {
		return ErrEmptyTimestamp
	}
	return nil
}
