package interfaces

import (
	"context"

	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
)

// SlackSearchConfigRepository defines interface for Slack search configuration persistence
type SlackSearchConfigRepository interface {
	// CRUD operations
	Create(ctx context.Context, config *agent.SlackSearchConfig) error
	GetByAgentID(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error)
	GetByID(ctx context.Context, id string) (*agent.SlackSearchConfig, error)
	Update(ctx context.Context, config *agent.SlackSearchConfig) error
	Delete(ctx context.Context, id string) error

	// Business queries
	GetEnabledByAgentID(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error)
	ExistsByAgentIDAndChannelID(ctx context.Context, agentID, channelID string) (bool, error)
}

// JiraSearchConfigRepository defines interface for Jira search configuration persistence
type JiraSearchConfigRepository interface {
	// CRUD operations
	Create(ctx context.Context, config *agent.JiraSearchConfig) error
	GetByAgentID(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error)
	GetByID(ctx context.Context, id string) (*agent.JiraSearchConfig, error)
	Update(ctx context.Context, config *agent.JiraSearchConfig) error
	Delete(ctx context.Context, id string) error

	// Business queries
	GetEnabledByAgentID(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error)
	ExistsByAgentIDAndProjectKey(ctx context.Context, agentID, projectKey string) (bool, error)
}

// NotionSearchConfigRepository defines interface for Notion search configuration persistence
type NotionSearchConfigRepository interface {
	// CRUD operations
	Create(ctx context.Context, config *agent.NotionSearchConfig) error
	GetByAgentID(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error)
	GetByID(ctx context.Context, id string) (*agent.NotionSearchConfig, error)
	Update(ctx context.Context, config *agent.NotionSearchConfig) error
	Delete(ctx context.Context, id string) error

	// Business queries
	GetEnabledByAgentID(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error)
	ExistsByAgentIDAndDatabaseID(ctx context.Context, agentID, databaseID string) (bool, error)
}
