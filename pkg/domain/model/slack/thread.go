package slack

import (
	"context"
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// Thread represents a Slack conversation thread
type Thread struct {
	ID        types.ThreadID `json:"id"`
	TeamID    string         `json:"team_id"`
	ChannelID string         `json:"channel_id"`
	ThreadTS  string         `json:"thread_ts"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// NewThread creates a new Thread instance
func NewThread(ctx context.Context, teamID, channelID, threadTS string) *Thread {
	now := time.Now()
	return &Thread{
		ID:        types.NewThreadID(ctx),
		TeamID:    teamID,
		ChannelID: channelID,
		ThreadTS:  threadTS,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate checks if the thread has valid fields
func (t *Thread) Validate() error {
	if !t.ID.IsValid() {
		return ErrInvalidThreadID
	}
	if t.TeamID == "" {
		return ErrEmptyTeamID
	}
	if t.ChannelID == "" {
		return ErrEmptyChannelID
	}
	if t.ThreadTS == "" {
		return ErrEmptyThreadTS
	}
	return nil
}
