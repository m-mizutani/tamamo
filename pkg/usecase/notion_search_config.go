package usecase

import (
	"context"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/domain/types/apperr"
)

// NotionSearchConfig holds dependencies for Notion search configuration use cases
type NotionSearchConfig struct {
	notionConfigRepo interfaces.NotionSearchConfigRepository
	agentRepo        interfaces.AgentRepository
}

// NotionSearchConfigOption is a functional option for NotionSearchConfig
type NotionSearchConfigOption func(*NotionSearchConfig)

// WithNotionSearchConfigRepository sets the Notion search config repository
func WithNotionSearchConfigRepository(repo interfaces.NotionSearchConfigRepository) NotionSearchConfigOption {
	return func(uc *NotionSearchConfig) {
		uc.notionConfigRepo = repo
	}
}

// WithNotionSearchConfigAgentRepository sets the agent repository
func WithNotionSearchConfigAgentRepository(repo interfaces.AgentRepository) NotionSearchConfigOption {
	return func(uc *NotionSearchConfig) {
		uc.agentRepo = repo
	}
}

// NewNotionSearchConfig creates a new NotionSearchConfig instance
func NewNotionSearchConfig(opts ...NotionSearchConfigOption) *NotionSearchConfig {
	uc := &NotionSearchConfig{}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// CreateNotionSearchConfig creates a new Notion search configuration
func (uc *NotionSearchConfig) CreateNotionSearchConfig(ctx context.Context, agentID, databaseID, databaseName, workspaceID string, description *string, enabled bool) (*agent.NotionSearchConfig, error) {
	// Parse agentID as UUID
	agentUUID := types.UUID(agentID)
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent ID format", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Validate that the agent exists
	if _, err := uc.agentRepo.GetAgent(ctx, agentUUID); err != nil {
		return nil, goerr.Wrap(err, "failed to validate agent existence", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Check if configuration already exists for this agent and database
	exists, err := uc.notionConfigRepo.ExistsByAgentIDAndDatabaseID(ctx, agentID, databaseID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to check existing configuration",
			goerr.TV(apperr.AgentIDKey, agentID), goerr.TV(apperr.DatabaseIDKey, databaseID))
	}
	if exists {
		return nil, goerr.New("configuration already exists for this agent and database",
			goerr.TV(apperr.AgentIDKey, agentID), goerr.TV(apperr.DatabaseIDKey, databaseID))
	}

	// Create new configuration
	config := agent.NewNotionSearchConfig(agentID, databaseID, databaseName, workspaceID, description, enabled)
	if err := config.Validate(); err != nil {
		return nil, goerr.Wrap(err, "validation failed for new Notion search config")
	}

	// Save to repository
	if err := uc.notionConfigRepo.Create(ctx, config); err != nil {
		return nil, goerr.Wrap(err, "failed to create Notion search config", goerr.TV(apperr.SearchConfigIDKey, config.ID))
	}

	return config, nil
}

// GetNotionSearchConfigs gets all Notion search configurations for an agent
func (uc *NotionSearchConfig) GetNotionSearchConfigs(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error) {
	// Parse agentID as UUID
	agentUUID := types.UUID(agentID)
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent ID format", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Validate that the agent exists
	if _, err := uc.agentRepo.GetAgent(ctx, agentUUID); err != nil {
		return nil, goerr.Wrap(err, "failed to validate agent existence", goerr.TV(apperr.AgentIDKey, agentID))
	}

	configs, err := uc.notionConfigRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get Notion search configs", goerr.TV(apperr.AgentIDKey, agentID))
	}

	return configs, nil
}

// GetNotionSearchConfig gets a specific Notion search configuration by ID
func (uc *NotionSearchConfig) GetNotionSearchConfig(ctx context.Context, id string) (*agent.NotionSearchConfig, error) {
	config, err := uc.notionConfigRepo.GetByID(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get Notion search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	return config, nil
}

// UpdateNotionSearchConfig updates a Notion search configuration
func (uc *NotionSearchConfig) UpdateNotionSearchConfig(ctx context.Context, id, databaseName, workspaceID string, description *string, enabled bool) (*agent.NotionSearchConfig, error) {
	// Get existing configuration
	existing, err := uc.notionConfigRepo.GetByID(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get existing Notion search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	// Create updated configuration
	updated := existing.Update(databaseName, workspaceID, description, enabled)
	if err := updated.Validate(); err != nil {
		return nil, goerr.Wrap(err, "validation failed for updated Notion search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	// Save updated configuration
	if err := uc.notionConfigRepo.Update(ctx, updated); err != nil {
		return nil, goerr.Wrap(err, "failed to update Notion search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	return updated, nil
}

// DeleteNotionSearchConfig deletes a Notion search configuration
func (uc *NotionSearchConfig) DeleteNotionSearchConfig(ctx context.Context, id string) error {
	// Check if configuration exists
	if _, err := uc.notionConfigRepo.GetByID(ctx, id); err != nil {
		return goerr.Wrap(err, "failed to get Notion search config for deletion", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	// Delete configuration
	if err := uc.notionConfigRepo.Delete(ctx, id); err != nil {
		return goerr.Wrap(err, "failed to delete Notion search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	return nil
}

// GetEnabledNotionSearchConfigs gets enabled Notion search configurations for an agent
func (uc *NotionSearchConfig) GetEnabledNotionSearchConfigs(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error) {
	// Parse agentID as UUID
	agentUUID := types.UUID(agentID)
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent ID format", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Validate that the agent exists
	if _, err := uc.agentRepo.GetAgent(ctx, agentUUID); err != nil {
		return nil, goerr.Wrap(err, "failed to validate agent existence", goerr.TV(apperr.AgentIDKey, agentID))
	}

	configs, err := uc.notionConfigRepo.GetEnabledByAgentID(ctx, agentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get enabled Notion search configs", goerr.TV(apperr.AgentIDKey, agentID))
	}

	return configs, nil
}

// Ensure NotionSearchConfig implements required interfaces
var _ interfaces.NotionSearchConfigUseCases = (*NotionSearchConfig)(nil)
