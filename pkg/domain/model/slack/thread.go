package slack

import (
	"context"
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// Thread represents a Slack conversation thread
type Thread struct {
	ID           types.ThreadID
	TeamID       string
	ChannelID    string
	ThreadTS     string
	AgentUUID    *types.UUID // Agent UUID (nullable, special UUID for general mode)
	AgentVersion string      // Agent version
	CreatedAt    time.Time
	UpdatedAt    time.Time
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

// NewThreadWithAgent creates a new Thread instance with agent information
func NewThreadWithAgent(ctx context.Context, teamID, channelID, threadTS string, agentUUID *types.UUID, agentVersion string) *Thread {
	now := time.Now()
	return &Thread{
		ID:           types.NewThreadID(ctx),
		TeamID:       teamID,
		ChannelID:    channelID,
		ThreadTS:     threadTS,
		AgentUUID:    agentUUID,
		AgentVersion: agentVersion,
		CreatedAt:    now,
		UpdatedAt:    now,
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
	// AgentUUID can be nil (for backward compatibility)
	if t.AgentUUID != nil && !t.AgentUUID.IsValid() {
		return ErrInvalidAgentUUID
	}
	// ThreadTS can be empty for new threads starting from channel-level messages
	return nil
}
