package agent

import (
	"time"

	"github.com/google/uuid"
	"github.com/m-mizutani/goerr/v2"
)

type JiraSearchConfig struct {
	ID          string    `json:"id"`
	AgentID     string    `json:"agentId"`
	ProjectKey  string    `json:"projectKey"`
	ProjectName string    `json:"projectName"`
	BoardID     *string   `json:"boardId,omitempty"`
	BoardName   *string   `json:"boardName,omitempty"`
	Description *string   `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewJiraSearchConfig creates a new JiraSearchConfig
func NewJiraSearchConfig(agentID, projectKey, projectName string, boardID, boardName, description *string, enabled bool) *JiraSearchConfig {
	now := time.Now()
	return &JiraSearchConfig{
		ID:          uuid.New().String(),
		AgentID:     agentID,
		ProjectKey:  projectKey,
		ProjectName: projectName,
		BoardID:     boardID,
		BoardName:   boardName,
		Description: description,
		Enabled:     enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate validates the JiraSearchConfig
func (c *JiraSearchConfig) Validate() error {
	if c.ID == "" {
		return goerr.New("ID is required")
	}
	if c.AgentID == "" {
		return goerr.New("AgentID is required")
	}
	if c.ProjectKey == "" {
		return goerr.New("ProjectKey is required")
	}
	if c.ProjectName == "" {
		return goerr.New("ProjectName is required")
	}
	return nil
}

// Update updates the JiraSearchConfig
func (c *JiraSearchConfig) Update(projectName string, boardID, boardName, description *string, enabled bool) *JiraSearchConfig {
	updated := *c
	updated.ProjectName = projectName
	updated.BoardID = boardID
	updated.BoardName = boardName
	updated.Description = description
	updated.Enabled = enabled
	updated.UpdatedAt = time.Now()
	return &updated
}
