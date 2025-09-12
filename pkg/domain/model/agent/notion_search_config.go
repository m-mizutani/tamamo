package agent

import (
	"time"

	"github.com/google/uuid"
	"github.com/m-mizutani/goerr/v2"
)

type NotionSearchConfig struct {
	ID           string    `json:"id"`
	AgentID      string    `json:"agentId"`
	DatabaseID   string    `json:"databaseId"`
	DatabaseName string    `json:"databaseName"`
	WorkspaceID  string    `json:"workspaceId"`
	Description  *string   `json:"description,omitempty"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// NewNotionSearchConfig creates a new NotionSearchConfig
func NewNotionSearchConfig(agentID, databaseID, databaseName, workspaceID string, description *string, enabled bool) *NotionSearchConfig {
	now := time.Now()
	return &NotionSearchConfig{
		ID:           uuid.New().String(),
		AgentID:      agentID,
		DatabaseID:   databaseID,
		DatabaseName: databaseName,
		WorkspaceID:  workspaceID,
		Description:  description,
		Enabled:      enabled,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Validate validates the NotionSearchConfig
func (c *NotionSearchConfig) Validate() error {
	if c.ID == "" {
		return goerr.New("ID is required")
	}
	if c.AgentID == "" {
		return goerr.New("AgentID is required")
	}
	if c.DatabaseID == "" {
		return goerr.New("DatabaseID is required")
	}
	if c.DatabaseName == "" {
		return goerr.New("DatabaseName is required")
	}
	if c.WorkspaceID == "" {
		return goerr.New("WorkspaceID is required")
	}
	return nil
}

// Update updates the NotionSearchConfig
func (c *NotionSearchConfig) Update(databaseName, workspaceID string, description *string, enabled bool) *NotionSearchConfig {
	updated := *c
	updated.DatabaseName = databaseName
	updated.WorkspaceID = workspaceID
	updated.Description = description
	updated.Enabled = enabled
	updated.UpdatedAt = time.Now()
	return &updated
}
