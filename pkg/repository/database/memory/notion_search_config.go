package memory

import (
	"context"
	"sync"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
)

type notionSearchConfigMemoryRepository struct {
	mu      sync.RWMutex
	configs map[string]*agent.NotionSearchConfig
}

// NewNotionSearchConfigRepository creates a new memory-based Notion search config repository
func NewNotionSearchConfigRepository() interfaces.NotionSearchConfigRepository {
	return &notionSearchConfigMemoryRepository{
		configs: make(map[string]*agent.NotionSearchConfig),
	}
}

func (r *notionSearchConfigMemoryRepository) Create(ctx context.Context, config *agent.NotionSearchConfig) error {
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

func (r *notionSearchConfigMemoryRepository) GetByAgentID(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.NotionSearchConfig
	for _, config := range r.configs {
		if config.AgentID == agentID {
			configCopy := *config
			result = append(result, &configCopy)
		}
	}

	return result, nil
}

func (r *notionSearchConfigMemoryRepository) GetByID(ctx context.Context, id string) (*agent.NotionSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, exists := r.configs[id]
	if !exists {
		return nil, goerr.New("config not found", goerr.V("id", id))
	}

	configCopy := *config
	return &configCopy, nil
}

func (r *notionSearchConfigMemoryRepository) Update(ctx context.Context, config *agent.NotionSearchConfig) error {
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

func (r *notionSearchConfigMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[id]; !exists {
		return goerr.New("config not found", goerr.V("id", id))
	}

	delete(r.configs, id)
	return nil
}

func (r *notionSearchConfigMemoryRepository) GetEnabledByAgentID(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.NotionSearchConfig
	for _, config := range r.configs {
		if config.AgentID == agentID && config.Enabled {
			configCopy := *config
			result = append(result, &configCopy)
		}
	}

	return result, nil
}

func (r *notionSearchConfigMemoryRepository) ExistsByAgentIDAndDatabaseID(ctx context.Context, agentID, databaseID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, config := range r.configs {
		if config.AgentID == agentID && config.DatabaseID == databaseID {
			return true, nil
		}
	}

	return false, nil
}
