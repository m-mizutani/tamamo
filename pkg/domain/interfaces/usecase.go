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
)

type SlackEventUseCases interface {
	HandleSlackAppMention(ctx context.Context, slackMsg slack.Message) error
	HandleSlackMessage(ctx context.Context, slackMsg slack.Message) error
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
