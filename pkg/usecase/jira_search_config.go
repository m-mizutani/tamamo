package usecase

import (
	"context"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/domain/types/apperr"
)

// JiraSearchConfig holds dependencies for Jira search configuration use cases
type JiraSearchConfig struct {
	jiraConfigRepo interfaces.JiraSearchConfigRepository
	agentRepo      interfaces.AgentRepository
}

// JiraSearchConfigOption is a functional option for JiraSearchConfig
type JiraSearchConfigOption func(*JiraSearchConfig)

// WithJiraSearchConfigRepository sets the Jira search config repository
func WithJiraSearchConfigRepository(repo interfaces.JiraSearchConfigRepository) JiraSearchConfigOption {
	return func(uc *JiraSearchConfig) {
		uc.jiraConfigRepo = repo
	}
}

// WithJiraSearchConfigAgentRepository sets the agent repository
func WithJiraSearchConfigAgentRepository(repo interfaces.AgentRepository) JiraSearchConfigOption {
	return func(uc *JiraSearchConfig) {
		uc.agentRepo = repo
	}
}

// NewJiraSearchConfig creates a new JiraSearchConfig instance
func NewJiraSearchConfig(opts ...JiraSearchConfigOption) *JiraSearchConfig {
	uc := &JiraSearchConfig{}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// CreateJiraSearchConfig creates a new Jira search configuration
func (uc *JiraSearchConfig) CreateJiraSearchConfig(ctx context.Context, agentID, projectKey, projectName string, boardID, boardName, description *string, enabled bool) (*agent.JiraSearchConfig, error) {
	// Parse agentID as UUID
	agentUUID := types.UUID(agentID)
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent ID format", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Validate that the agent exists
	if _, err := uc.agentRepo.GetAgent(ctx, agentUUID); err != nil {
		return nil, goerr.Wrap(err, "failed to validate agent existence", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Check if configuration already exists for this agent and project
	exists, err := uc.jiraConfigRepo.ExistsByAgentIDAndProjectKey(ctx, agentID, projectKey)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to check existing configuration",
			goerr.TV(apperr.AgentIDKey, agentID), goerr.TV(apperr.ProjectKeyKey, projectKey))
	}
	if exists {
		return nil, goerr.New("configuration already exists for this agent and project",
			goerr.TV(apperr.AgentIDKey, agentID), goerr.TV(apperr.ProjectKeyKey, projectKey))
	}

	// Create new configuration
	config := agent.NewJiraSearchConfig(agentID, projectKey, projectName, boardID, boardName, description, enabled)
	if err := config.Validate(); err != nil {
		return nil, goerr.Wrap(err, "validation failed for new Jira search config")
	}

	// Save to repository
	if err := uc.jiraConfigRepo.Create(ctx, config); err != nil {
		return nil, goerr.Wrap(err, "failed to create Jira search config", goerr.TV(apperr.SearchConfigIDKey, config.ID))
	}

	return config, nil
}

// GetJiraSearchConfigs gets all Jira search configurations for an agent
func (uc *JiraSearchConfig) GetJiraSearchConfigs(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error) {
	// Parse agentID as UUID
	agentUUID := types.UUID(agentID)
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent ID format", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Validate that the agent exists
	if _, err := uc.agentRepo.GetAgent(ctx, agentUUID); err != nil {
		return nil, goerr.Wrap(err, "failed to validate agent existence", goerr.TV(apperr.AgentIDKey, agentID))
	}

	configs, err := uc.jiraConfigRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get Jira search configs", goerr.TV(apperr.AgentIDKey, agentID))
	}

	return configs, nil
}

// GetJiraSearchConfig gets a specific Jira search configuration by ID
func (uc *JiraSearchConfig) GetJiraSearchConfig(ctx context.Context, id string) (*agent.JiraSearchConfig, error) {
	config, err := uc.jiraConfigRepo.GetByID(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get Jira search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	return config, nil
}

// UpdateJiraSearchConfig updates a Jira search configuration
func (uc *JiraSearchConfig) UpdateJiraSearchConfig(ctx context.Context, id, projectName string, boardID, boardName, description *string, enabled bool) (*agent.JiraSearchConfig, error) {
	// Get existing configuration
	existing, err := uc.jiraConfigRepo.GetByID(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get existing Jira search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	// Create updated configuration
	updated := existing.Update(projectName, boardID, boardName, description, enabled)
	if err := updated.Validate(); err != nil {
		return nil, goerr.Wrap(err, "validation failed for updated Jira search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	// Save updated configuration
	if err := uc.jiraConfigRepo.Update(ctx, updated); err != nil {
		return nil, goerr.Wrap(err, "failed to update Jira search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	return updated, nil
}

// DeleteJiraSearchConfig deletes a Jira search configuration
func (uc *JiraSearchConfig) DeleteJiraSearchConfig(ctx context.Context, id string) error {
	// Check if configuration exists
	if _, err := uc.jiraConfigRepo.GetByID(ctx, id); err != nil {
		return goerr.Wrap(err, "failed to get Jira search config for deletion", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	// Delete configuration
	if err := uc.jiraConfigRepo.Delete(ctx, id); err != nil {
		return goerr.Wrap(err, "failed to delete Jira search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	return nil
}

// GetEnabledJiraSearchConfigs gets enabled Jira search configurations for an agent
func (uc *JiraSearchConfig) GetEnabledJiraSearchConfigs(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error) {
	// Parse agentID as UUID
	agentUUID := types.UUID(agentID)
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent ID format", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Validate that the agent exists
	if _, err := uc.agentRepo.GetAgent(ctx, agentUUID); err != nil {
		return nil, goerr.Wrap(err, "failed to validate agent existence", goerr.TV(apperr.AgentIDKey, agentID))
	}

	configs, err := uc.jiraConfigRepo.GetEnabledByAgentID(ctx, agentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get enabled Jira search configs", goerr.TV(apperr.AgentIDKey, agentID))
	}

	return configs, nil
}

// Ensure JiraSearchConfig implements required interfaces
var _ interfaces.JiraSearchConfigUseCases = (*JiraSearchConfig)(nil)
