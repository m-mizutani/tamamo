package memory

import (
	"context"
	"sync"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
)

type slackSearchConfigMemoryRepository struct {
	mu      sync.RWMutex
	configs map[string]*agent.SlackSearchConfig
}

// NewSlackSearchConfigRepository creates a new memory-based Slack search config repository
func NewSlackSearchConfigRepository() interfaces.SlackSearchConfigRepository {
	return &slackSearchConfigMemoryRepository{
		configs: make(map[string]*agent.SlackSearchConfig),
	}
}

func (r *slackSearchConfigMemoryRepository) Create(ctx context.Context, config *agent.SlackSearchConfig) error {
	if err := config.Validate(); err != nil {
		return goerr.Wrap(err, "invalid config")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[config.ID]; exists {
		return goerr.New("config already exists", goerr.V("id", config.ID))
	}

	r.configs[config.ID] = config
	return nil
}

func (r *slackSearchConfigMemoryRepository) GetByAgentID(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.SlackSearchConfig
	for _, config := range r.configs {
		if config.AgentID == agentID {
			// Create a copy to avoid external modifications
			configCopy := *config
			result = append(result, &configCopy)
		}
	}

	return result, nil
}

func (r *slackSearchConfigMemoryRepository) GetByID(ctx context.Context, id string) (*agent.SlackSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, exists := r.configs[id]
	if !exists {
		return nil, goerr.New("config not found", goerr.V("id", id))
	}

	// Return a copy to avoid external modifications
	configCopy := *config
	return &configCopy, nil
}

func (r *slackSearchConfigMemoryRepository) Update(ctx context.Context, config *agent.SlackSearchConfig) error {
	if err := config.Validate(); err != nil {
		return goerr.Wrap(err, "invalid config")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[config.ID]; !exists {
		return goerr.New("config not found", goerr.V("id", config.ID))
	}

	r.configs[config.ID] = config
	return nil
}

func (r *slackSearchConfigMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[id]; !exists {
		return goerr.New("config not found", goerr.V("id", id))
	}

	delete(r.configs, id)
	return nil
}

func (r *slackSearchConfigMemoryRepository) GetEnabledByAgentID(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.SlackSearchConfig
	for _, config := range r.configs {
		if config.AgentID == agentID && config.Enabled {
			configCopy := *config
			result = append(result, &configCopy)
		}
	}

	return result, nil
}

func (r *slackSearchConfigMemoryRepository) ExistsByAgentIDAndChannelID(ctx context.Context, agentID, channelID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, config := range r.configs {
		if config.AgentID == agentID && config.ChannelID == channelID {
			return true, nil
		}
	}

	return false, nil
}
