package interfaces

import (
	"context"

	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// ThreadRepository manages thread, message and history persistence
type ThreadRepository interface {
	// Thread operations
	GetThread(ctx context.Context, id types.ThreadID) (*slack.Thread, error)
	GetThreadByTS(ctx context.Context, channelID, threadTS string) (*slack.Thread, error)
	GetOrPutThread(ctx context.Context, teamID, channelID, threadTS string) (*slack.Thread, error)
	ListThreads(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error)

	// Message operations
	PutThreadMessage(ctx context.Context, threadID types.ThreadID, message *slack.Message) error
	GetThreadMessages(ctx context.Context, threadID types.ThreadID) ([]*slack.Message, error)

	// History operations
	PutHistory(ctx context.Context, history *slack.History) error
	GetLatestHistory(ctx context.Context, threadID types.ThreadID) (*slack.History, error)
	GetHistoryByID(ctx context.Context, id types.HistoryID) (*slack.History, error)
}

// AgentRepository manages agent and agent version persistence
type AgentRepository interface {
	// Agent CRUD
	CreateAgent(ctx context.Context, agent *agent.Agent) error
	GetAgent(ctx context.Context, id types.UUID) (*agent.Agent, error)
	GetAgentByAgentID(ctx context.Context, agentID string) (*agent.Agent, error)
	UpdateAgent(ctx context.Context, agent *agent.Agent) error
	DeleteAgent(ctx context.Context, id types.UUID) error
	ListAgents(ctx context.Context, offset, limit int) ([]*agent.Agent, int, error)

	// Version management
	CreateAgentVersion(ctx context.Context, version *agent.AgentVersion) error
	GetAgentVersion(ctx context.Context, agentUUID types.UUID, version string) (*agent.AgentVersion, error)
	GetLatestAgentVersion(ctx context.Context, agentUUID types.UUID) (*agent.AgentVersion, error)
	ListAgentVersions(ctx context.Context, agentUUID types.UUID) ([]*agent.AgentVersion, error)
	UpdateAgentVersion(ctx context.Context, version *agent.AgentVersion) error

	// Efficient queries for performance optimization
	ListAgentsWithLatestVersions(ctx context.Context, offset, limit int) ([]*agent.Agent, []*agent.AgentVersion, int, error)

	// Utilities
	AgentIDExists(ctx context.Context, agentID string) (bool, error)
}

// HistoryRepository is deprecated - use ThreadRepository instead
// Kept for backward compatibility
type HistoryRepository interface {
	// History operations
	PutHistory(ctx context.Context, history *slack.History) error
	GetLatestHistory(ctx context.Context, threadID types.ThreadID) (*slack.History, error)
	GetHistoryByID(ctx context.Context, id types.HistoryID) (*slack.History, error)
}
