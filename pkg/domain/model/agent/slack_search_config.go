package agent

import (
	"time"

	"github.com/google/uuid"
	"github.com/m-mizutani/goerr/v2"
)

type SlackSearchConfig struct {
	ID          string    `json:"id"`
	AgentID     string    `json:"agentId"`
	ChannelID   string    `json:"channelId"`
	ChannelName string    `json:"channelName"`
	Description *string   `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewSlackSearchConfig creates a new SlackSearchConfig
func NewSlackSearchConfig(agentID, channelID, channelName string, description *string, enabled bool) *SlackSearchConfig {
	now := time.Now()
	return &SlackSearchConfig{
		ID:          uuid.New().String(),
		AgentID:     agentID,
		ChannelID:   channelID,
		ChannelName: channelName,
		Description: description,
		Enabled:     enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate validates the SlackSearchConfig
func (c *SlackSearchConfig) Validate() error {
	if c.ID == "" {
		return goerr.New("ID is required")
	}
	if c.AgentID == "" {
		return goerr.New("AgentID is required")
	}
	if c.ChannelID == "" {
		return goerr.New("ChannelID is required")
	}
	if c.ChannelName == "" {
		return goerr.New("ChannelName is required")
	}
	return nil
}

// Update updates the SlackSearchConfig
func (c *SlackSearchConfig) Update(channelName string, description *string, enabled bool) *SlackSearchConfig {
	updated := *c
	updated.ChannelName = channelName
	updated.Description = description
	updated.Enabled = enabled
	updated.UpdatedAt = time.Now()
	return &updated
}
