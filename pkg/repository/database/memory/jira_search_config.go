package memory

import (
	"context"
	"sync"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
)

type jiraSearchConfigMemoryRepository struct {
	mu      sync.RWMutex
	configs map[string]*agent.JiraSearchConfig
}

// NewJiraSearchConfigRepository creates a new memory-based Jira search config repository
func NewJiraSearchConfigRepository() interfaces.JiraSearchConfigRepository {
	return &jiraSearchConfigMemoryRepository{
		configs: make(map[string]*agent.JiraSearchConfig),
	}
}

func (r *jiraSearchConfigMemoryRepository) Create(ctx context.Context, config *agent.JiraSearchConfig) error {
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

func (r *jiraSearchConfigMemoryRepository) GetByAgentID(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.JiraSearchConfig
	for _, config := range r.configs {
		if config.AgentID == agentID {
			configCopy := *config
			result = append(result, &configCopy)
		}
	}

	return result, nil
}

func (r *jiraSearchConfigMemoryRepository) GetByID(ctx context.Context, id string) (*agent.JiraSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, exists := r.configs[id]
	if !exists {
		return nil, goerr.New("config not found", goerr.V("id", id))
	}

	configCopy := *config
	return &configCopy, nil
}

func (r *jiraSearchConfigMemoryRepository) Update(ctx context.Context, config *agent.JiraSearchConfig) error {
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

func (r *jiraSearchConfigMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[id]; !exists {
		return goerr.New("config not found", goerr.V("id", id))
	}

	delete(r.configs, id)
	return nil
}

func (r *jiraSearchConfigMemoryRepository) GetEnabledByAgentID(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.JiraSearchConfig
	for _, config := range r.configs {
		if config.AgentID == agentID && config.Enabled {
			configCopy := *config
			result = append(result, &configCopy)
		}
	}

	return result, nil
}

func (r *jiraSearchConfigMemoryRepository) ExistsByAgentIDAndProjectKey(ctx context.Context, agentID, projectKey string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, config := range r.configs {
		if config.AgentID == agentID && config.ProjectKey == projectKey {
			return true, nil
		}
	}

	return false, nil
}
