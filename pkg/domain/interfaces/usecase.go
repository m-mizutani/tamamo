package interfaces

import (
	"context"

	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
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
	LLMProvider  agent.LLMProvider `json:"llm_provider"`
	LLMModel     string            `json:"llm_model"`
	Version      string            `json:"version"` // Initial version, defaults to "1.0.0"
}

type UpdateAgentRequest struct {
	AgentID      *string            `json:"agent_id,omitempty"`
	Name         *string            `json:"name,omitempty"`
	Description  *string            `json:"description,omitempty"`
	SystemPrompt *string            `json:"system_prompt,omitempty"`
	LLMProvider  *agent.LLMProvider `json:"llm_provider,omitempty"`
	LLMModel     *string            `json:"llm_model,omitempty"`
}

type CreateVersionRequest struct {
	AgentUUID    types.UUID        `json:"agent_uuid"`
	Version      string            `json:"version"`
	SystemPrompt *string           `json:"system_prompt,omitempty"`
	LLMProvider  agent.LLMProvider `json:"llm_provider"`
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

	// Version management
	CreateAgentVersion(ctx context.Context, req *CreateVersionRequest) (*agent.AgentVersion, error)
	GetAgentVersions(ctx context.Context, agentUUID types.UUID) ([]*agent.AgentVersion, error)

	// Validation (independent execution)
	CheckAgentIDAvailability(ctx context.Context, agentID string) (*AgentIDAvailability, error)
	ValidateAgentID(agentID string) error
	ValidateVersion(version string) error
}
