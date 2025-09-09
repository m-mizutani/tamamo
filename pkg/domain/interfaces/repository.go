package interfaces

import (
	"context"
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
	"github.com/m-mizutani/tamamo/pkg/domain/model/image"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// ThreadRepository manages thread, message and history persistence
type ThreadRepository interface {
	// Thread operations
	GetThread(ctx context.Context, id types.ThreadID) (*slack.Thread, error)
	GetThreadByTS(ctx context.Context, channelID, threadTS string) (*slack.Thread, error)
	GetOrPutThread(ctx context.Context, teamID, channelID, threadTS string) (*slack.Thread, error)
	GetOrPutThreadWithAgent(ctx context.Context, teamID, channelID, threadTS string, agentUUID *types.UUID, agentVersion string) (*slack.Thread, error)
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
	ListActiveAgentsWithLatestVersions(ctx context.Context, offset, limit int) ([]*agent.Agent, []*agent.AgentVersion, int, error)
	ListAgentsByStatusWithLatestVersions(ctx context.Context, status agent.Status, offset, limit int) ([]*agent.Agent, []*agent.AgentVersion, int, error)

	// Status management
	UpdateAgentStatus(ctx context.Context, id types.UUID, status agent.Status) error

	// Filtered queries
	ListActiveAgents(ctx context.Context, offset, limit int) ([]*agent.Agent, int, error)
	ListAgentsByStatus(ctx context.Context, status agent.Status, offset, limit int) ([]*agent.Agent, int, error)
	GetAgentByAgentIDActive(ctx context.Context, agentID string) (*agent.Agent, error)

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

// SessionRepository manages user session persistence
type SessionRepository interface {
	CreateSession(ctx context.Context, session *auth.Session) error
	GetSession(ctx context.Context, sessionID string) (*auth.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	CleanupExpiredSessions(ctx context.Context) error
}

// OAuthStateRepository manages OAuth state for CSRF protection
type OAuthStateRepository interface {
	SaveState(ctx context.Context, state *auth.OAuthState) error
	GetState(ctx context.Context, state string) (*auth.OAuthState, error)
	ValidateAndDeleteState(ctx context.Context, state string) error
}

// UserRepository manages user persistence
type UserRepository interface {
	GetByID(ctx context.Context, id types.UserID) (*user.User, error)
	GetBySlackIDAndTeamID(ctx context.Context, slackID, teamID string) (*user.User, error)
	Create(ctx context.Context, user *user.User) error
	Update(ctx context.Context, user *user.User) error

	// Jira Integration methods
	SaveJiraIntegration(ctx context.Context, integration *integration.JiraIntegration) error
	GetJiraIntegration(ctx context.Context, userID string) (*integration.JiraIntegration, error)
	DeleteJiraIntegration(ctx context.Context, userID string) error
}

// AgentImageRepository manages agent image persistence
type AgentImageRepository interface {
	// Create creates a new agent image
	Create(ctx context.Context, agentImage *image.AgentImage) error

	// GetByID retrieves an agent image by its ID
	GetByID(ctx context.Context, id types.UUID) (*image.AgentImage, error)

	// Update updates an existing agent image
	Update(ctx context.Context, agentImage *image.AgentImage) error
}

// SlackMessageLogRepository manages Slack message log persistence
type SlackMessageLogRepository interface {
	// PutSlackMessageLog stores a Slack message log entry
	PutSlackMessageLog(ctx context.Context, messageLog *slack.SlackMessageLog) error

	// GetSlackMessageLogs retrieves message logs with filtering (primarily for channel and time period)
	GetSlackMessageLogs(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error)
}
