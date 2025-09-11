package integration

import (
	"time"
)

// NotionIntegration represents a Notion OAuth integration for a user
type NotionIntegration struct {
	UserID        string
	WorkspaceID   string
	WorkspaceName string
	WorkspaceIcon string // URL to workspace icon
	BotID         string
	AccessToken   string // Notion tokens don't expire and don't have refresh tokens
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewNotionIntegration creates a new NotionIntegration instance
func NewNotionIntegration(userID string) *NotionIntegration {
	now := time.Now()
	return &NotionIntegration{
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateTokens updates the access token
// Note: Notion doesn't provide refresh tokens and tokens don't expire
func (n *NotionIntegration) UpdateTokens(accessToken string) {
	n.AccessToken = accessToken
	n.UpdatedAt = time.Now()
}

// UpdateWorkspaceInfo updates the workspace information
func (n *NotionIntegration) UpdateWorkspaceInfo(workspaceID, workspaceName, workspaceIcon, botID string) {
	n.WorkspaceID = workspaceID
	n.WorkspaceName = workspaceName
	n.WorkspaceIcon = workspaceIcon
	n.BotID = botID
	n.UpdatedAt = time.Now()
}

// IsConnected checks if the integration is connected
// Since Notion tokens don't expire, we just check if token exists
func (n *NotionIntegration) IsConnected() bool {
	return n.AccessToken != ""
}
