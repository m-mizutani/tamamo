package interfaces

import (
	"context"
	"io"

	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
	"github.com/m-mizutani/tamamo/pkg/domain/model/image"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/slack-go/slack/slackevents"
)

type SlackEventUseCases interface {
	HandleSlackAppMention(ctx context.Context, slackMsg slack.Message) error
	HandleSlackMessage(ctx context.Context, slackMsg slack.Message) error
	LogSlackAppMentionMessage(ctx context.Context, event *slackevents.AppMentionEvent, teamID string) error
	LogSlackMessage(ctx context.Context, event *slackevents.MessageEvent, teamID string) error
}

// Agent use case request/response types
type CreateAgentRequest struct {
	AgentID      string            `json:"agent_id"`
	Name         string            `json:"name"`
	Description  *string           `json:"description,omitempty"`
	SystemPrompt *string           `json:"system_prompt,omitempty"`
	LLMProvider  types.LLMProvider `json:"llm_provider"`
	LLMModel     string            `json:"llm_model"`
	Version      string            `json:"version"` // Initial version, defaults to "1.0.0"
}

type UpdateAgentRequest struct {
	AgentID      *string            `json:"agent_id,omitempty"`
	Name         *string            `json:"name,omitempty"`
	Description  *string            `json:"description,omitempty"`
	SystemPrompt *string            `json:"system_prompt,omitempty"`
	LLMProvider  *types.LLMProvider `json:"llm_provider,omitempty"`
	LLMModel     *string            `json:"llm_model,omitempty"`
}

type CreateVersionRequest struct {
	AgentUUID    types.UUID        `json:"agent_uuid"`
	Version      string            `json:"version"`
	SystemPrompt *string           `json:"system_prompt,omitempty"`
	LLMProvider  types.LLMProvider `json:"llm_provider"`
	LLMModel     string            `json:"llm_model"`
}

type AgentWithVersion struct {
	Agent         *agent.Agent        `json:"agent"`
	LatestVersion *agent.AgentVersion `json:"latest_version"`
}

type AgentListResponse struct {
	Agents     []*AgentWithVersion `json:"agents"`
	TotalCount int                 `json:"total_count"`
}

type AgentIDAvailability struct {
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

type AgentUseCases interface {
	// Agent management
	CreateAgent(ctx context.Context, req *CreateAgentRequest) (*agent.Agent, error)
	GetAgent(ctx context.Context, id types.UUID) (*AgentWithVersion, error)
	UpdateAgent(ctx context.Context, id types.UUID, req *UpdateAgentRequest) (*agent.Agent, error)
	DeleteAgent(ctx context.Context, id types.UUID) error
	ListAgents(ctx context.Context, offset, limit int) (*AgentListResponse, error)
	ListAllAgents(ctx context.Context, offset, limit int) (*AgentListResponse, error)
	ListAgentsByStatus(ctx context.Context, status agent.Status, offset, limit int) (*AgentListResponse, error)

	// Agent status management
	ArchiveAgent(ctx context.Context, id types.UUID) (*AgentWithVersion, error)
	UnarchiveAgent(ctx context.Context, id types.UUID) (*AgentWithVersion, error)

	// Version management
	CreateAgentVersion(ctx context.Context, req *CreateVersionRequest) (*agent.AgentVersion, error)
	GetAgentVersions(ctx context.Context, agentUUID types.UUID) ([]*agent.AgentVersion, error)

	// Validation (independent execution)
	CheckAgentIDAvailability(ctx context.Context, agentID string) (*AgentIDAvailability, error)
	ValidateAgentID(agentID string) error
	ValidateVersion(version string) error
}

// AuthUseCases handles authentication and session management
type AuthUseCases interface {
	// OAuth flow
	GenerateLoginURL(ctx context.Context, state string) (string, error)
	HandleCallback(ctx context.Context, code string) (*auth.Session, error)

	// Session management
	GetSession(ctx context.Context, sessionID string) (*auth.Session, error)
	Logout(ctx context.Context, sessionID string) error

	// Session cleanup (should be called periodically)
	CleanupExpiredSessions(ctx context.Context) error
}

// UserUseCases handles user management
type UserUseCases interface {
	GetOrCreateUser(ctx context.Context, slackID, slackName, email, teamID string) (*user.User, error)
	GetUserByID(ctx context.Context, userID types.UserID) (*user.User, error)
	UpdateUser(ctx context.Context, user *user.User) error
	GetUserAvatar(ctx context.Context, userID types.UserID, size int) ([]byte, error)
	InvalidateUserAvatarCache(ctx context.Context, userID types.UserID) error
}

// UploadImageRequest represents an image upload request
type UploadImageRequest struct {
	AgentID     types.UUID    `json:"agent_id"`
	FileReader  io.ReadSeeker `json:"-"`
	ContentType string        `json:"content_type"`
	FileSize    int64         `json:"file_size"`
}

// ImageData represents image data with metadata
type ImageData struct {
	Data        []byte `json:"-"`
	ContentType string `json:"content_type"`
}

// ImageUseCases handles image management operations
type ImageUseCases interface {
	// Upload and process agent image
	UploadAgentImage(ctx context.Context, req *UploadImageRequest) (*image.AgentImage, error)

	// Get agent image data (original or thumbnail)
	GetAgentImageData(ctx context.Context, agentID types.UUID, thumbnailSize string) (*ImageData, error)

	// Get agent image metadata
	GetAgentImageInfo(ctx context.Context, agentID types.UUID) (*image.AgentImage, error)
}

// SlackSearchConfigUseCases handles Slack search configuration management
type SlackSearchConfigUseCases interface {
	// Create a new Slack search configuration
	CreateSlackSearchConfig(ctx context.Context, agentID, channelID, channelName string, description *string, enabled bool) (*agent.SlackSearchConfig, error)

	// Get all Slack search configurations for an agent
	GetSlackSearchConfigs(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error)

	// Get a specific Slack search configuration by ID
	GetSlackSearchConfig(ctx context.Context, id string) (*agent.SlackSearchConfig, error)

	// Update a Slack search configuration
	UpdateSlackSearchConfig(ctx context.Context, id, channelName string, description *string, enabled bool) (*agent.SlackSearchConfig, error)

	// Delete a Slack search configuration
	DeleteSlackSearchConfig(ctx context.Context, id string) error

	// Get enabled Slack search configurations for an agent
	GetEnabledSlackSearchConfigs(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error)
}

// JiraSearchConfigUseCases handles Jira search configuration management
type JiraSearchConfigUseCases interface {
	// Create a new Jira search configuration
	CreateJiraSearchConfig(ctx context.Context, agentID, projectKey, projectName string, boardID, boardName, description *string, enabled bool) (*agent.JiraSearchConfig, error)

	// Get all Jira search configurations for an agent
	GetJiraSearchConfigs(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error)

	// Get a specific Jira search configuration by ID
	GetJiraSearchConfig(ctx context.Context, id string) (*agent.JiraSearchConfig, error)

	// Update a Jira search configuration
	UpdateJiraSearchConfig(ctx context.Context, id, projectName string, boardID, boardName, description *string, enabled bool) (*agent.JiraSearchConfig, error)

	// Delete a Jira search configuration
	DeleteJiraSearchConfig(ctx context.Context, id string) error

	// Get enabled Jira search configurations for an agent
	GetEnabledJiraSearchConfigs(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error)
}

// NotionSearchConfigUseCases handles Notion search configuration management
type NotionSearchConfigUseCases interface {
	// Create a new Notion search configuration
	CreateNotionSearchConfig(ctx context.Context, agentID, databaseID, databaseName, workspaceID string, description *string, enabled bool) (*agent.NotionSearchConfig, error)

	// Get all Notion search configurations for an agent
	GetNotionSearchConfigs(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error)

	// Get a specific Notion search configuration by ID
	GetNotionSearchConfig(ctx context.Context, id string) (*agent.NotionSearchConfig, error)

	// Update a Notion search configuration
	UpdateNotionSearchConfig(ctx context.Context, id, databaseName, workspaceID string, description *string, enabled bool) (*agent.NotionSearchConfig, error)

	// Delete a Notion search configuration
	DeleteNotionSearchConfig(ctx context.Context, id string) error

	// Get enabled Notion search configurations for an agent
	GetEnabledNotionSearchConfigs(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error)
}
