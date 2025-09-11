package usecase

import (
	"context"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/domain/types/apperr"
)

// SlackSearchConfig holds dependencies for Slack search configuration use cases
type SlackSearchConfig struct {
	slackConfigRepo interfaces.SlackSearchConfigRepository
	agentRepo       interfaces.AgentRepository
}

// SlackSearchConfigOption is a functional option for SlackSearchConfig
type SlackSearchConfigOption func(*SlackSearchConfig)

// WithSlackSearchConfigRepository sets the Slack search config repository
func WithSlackSearchConfigRepository(repo interfaces.SlackSearchConfigRepository) SlackSearchConfigOption {
	return func(uc *SlackSearchConfig) {
		uc.slackConfigRepo = repo
	}
}

// WithSlackSearchConfigAgentRepository sets the agent repository
func WithSlackSearchConfigAgentRepository(repo interfaces.AgentRepository) SlackSearchConfigOption {
	return func(uc *SlackSearchConfig) {
		uc.agentRepo = repo
	}
}

// NewSlackSearchConfig creates a new SlackSearchConfig instance
func NewSlackSearchConfig(opts ...SlackSearchConfigOption) *SlackSearchConfig {
	uc := &SlackSearchConfig{}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// CreateSlackSearchConfig creates a new Slack search configuration
func (uc *SlackSearchConfig) CreateSlackSearchConfig(ctx context.Context, agentID, channelID, channelName string, description *string, enabled bool) (*agent.SlackSearchConfig, error) {
	// Parse agentID as UUID
	agentUUID := types.UUID(agentID)
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent ID format", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Validate that the agent exists
	if _, err := uc.agentRepo.GetAgent(ctx, agentUUID); err != nil {
		return nil, goerr.Wrap(err, "failed to validate agent existence", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Check if configuration already exists for this agent and channel
	exists, err := uc.slackConfigRepo.ExistsByAgentIDAndChannelID(ctx, agentID, channelID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to check existing configuration",
			goerr.TV(apperr.AgentIDKey, agentID), goerr.TV(apperr.ChannelIDKey, channelID))
	}
	if exists {
		return nil, goerr.New("configuration already exists for this agent and channel",
			goerr.TV(apperr.AgentIDKey, agentID), goerr.TV(apperr.ChannelIDKey, channelID))
	}

	// Create new configuration
	config := agent.NewSlackSearchConfig(agentID, channelID, channelName, description, enabled)
	if err := config.Validate(); err != nil {
		return nil, goerr.Wrap(err, "validation failed for new Slack search config")
	}

	// Save to repository
	if err := uc.slackConfigRepo.Create(ctx, config); err != nil {
		return nil, goerr.Wrap(err, "failed to create Slack search config", goerr.TV(apperr.SearchConfigIDKey, config.ID))
	}

	return config, nil
}

// GetSlackSearchConfigs gets all Slack search configurations for an agent
func (uc *SlackSearchConfig) GetSlackSearchConfigs(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error) {
	// Parse agentID as UUID
	agentUUID := types.UUID(agentID)
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent ID format", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Validate that the agent exists
	if _, err := uc.agentRepo.GetAgent(ctx, agentUUID); err != nil {
		return nil, goerr.Wrap(err, "failed to validate agent existence", goerr.TV(apperr.AgentIDKey, agentID))
	}

	configs, err := uc.slackConfigRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get Slack search configs", goerr.TV(apperr.AgentIDKey, agentID))
	}

	return configs, nil
}

// GetSlackSearchConfig gets a specific Slack search configuration by ID
func (uc *SlackSearchConfig) GetSlackSearchConfig(ctx context.Context, id string) (*agent.SlackSearchConfig, error) {
	config, err := uc.slackConfigRepo.GetByID(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get Slack search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	return config, nil
}

// UpdateSlackSearchConfig updates a Slack search configuration
func (uc *SlackSearchConfig) UpdateSlackSearchConfig(ctx context.Context, id, channelName string, description *string, enabled bool) (*agent.SlackSearchConfig, error) {
	// Get existing configuration
	existing, err := uc.slackConfigRepo.GetByID(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get existing Slack search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	// Create updated configuration
	updated := existing.Update(channelName, description, enabled)
	if err := updated.Validate(); err != nil {
		return nil, goerr.Wrap(err, "validation failed for updated Slack search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	// Save updated configuration
	if err := uc.slackConfigRepo.Update(ctx, updated); err != nil {
		return nil, goerr.Wrap(err, "failed to update Slack search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	return updated, nil
}

// DeleteSlackSearchConfig deletes a Slack search configuration
func (uc *SlackSearchConfig) DeleteSlackSearchConfig(ctx context.Context, id string) error {
	// Check if configuration exists
	if _, err := uc.slackConfigRepo.GetByID(ctx, id); err != nil {
		return goerr.Wrap(err, "failed to get Slack search config for deletion", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	// Delete configuration
	if err := uc.slackConfigRepo.Delete(ctx, id); err != nil {
		return goerr.Wrap(err, "failed to delete Slack search config", goerr.TV(apperr.SearchConfigIDKey, id))
	}

	return nil
}

// GetEnabledSlackSearchConfigs gets enabled Slack search configurations for an agent
func (uc *SlackSearchConfig) GetEnabledSlackSearchConfigs(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error) {
	// Parse agentID as UUID
	agentUUID := types.UUID(agentID)
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent ID format", goerr.TV(apperr.AgentIDKey, agentID))
	}

	// Validate that the agent exists
	if _, err := uc.agentRepo.GetAgent(ctx, agentUUID); err != nil {
		return nil, goerr.Wrap(err, "failed to validate agent existence", goerr.TV(apperr.AgentIDKey, agentID))
	}

	configs, err := uc.slackConfigRepo.GetEnabledByAgentID(ctx, agentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get enabled Slack search configs", goerr.TV(apperr.AgentIDKey, agentID))
	}

	return configs, nil
}

// Ensure SlackSearchConfig implements required interfaces
var _ interfaces.SlackSearchConfigUseCases = (*SlackSearchConfig)(nil)
