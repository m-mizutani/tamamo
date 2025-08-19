package memory

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// AgentMemoryClient is an in-memory implementation of AgentRepository
type AgentMemoryClient struct {
	*Client  // Embed existing memory client
	agents   map[types.UUID]*agent.Agent
	versions map[types.UUID]map[string]*agent.AgentVersion // agentUUID -> version -> AgentVersion
}

// NewAgentMemoryClient creates a new in-memory agent client
func NewAgentMemoryClient() *AgentMemoryClient {
	return &AgentMemoryClient{
		Client:   New(),
		agents:   make(map[types.UUID]*agent.Agent),
		versions: make(map[types.UUID]map[string]*agent.AgentVersion),
	}
}

// CreateAgent creates a new agent
func (c *AgentMemoryClient) CreateAgent(ctx context.Context, agent *agent.Agent) error {
	if agent == nil {
		return goerr.New("agent cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if AgentID already exists
	for _, existing := range c.agents {
		if existing.AgentID == agent.AgentID {
			return goerr.New("agent ID already exists", goerr.V("agent_id", agent.AgentID))
		}
	}

	if !agent.ID.IsValid() {
		agent.ID = types.NewUUID(ctx)
	}

	now := time.Now()
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = now
	}
	agent.UpdatedAt = now

	// Create a copy to avoid external modifications
	agentCopy := *agent
	c.agents[agent.ID] = &agentCopy

	return nil
}

// GetAgent retrieves an agent by ID
func (c *AgentMemoryClient) GetAgent(ctx context.Context, id types.UUID) (*agent.Agent, error) {
	if !id.IsValid() {
		return nil, goerr.New("invalid agent ID")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	agent, exists := c.agents[id]
	if !exists {
		return nil, goerr.New("agent not found", goerr.V("id", id.String()))
	}

	// Return a copy to avoid external modifications
	agentCopy := *agent
	return &agentCopy, nil
}

// GetAgentByAgentID retrieves an agent by AgentID
func (c *AgentMemoryClient) GetAgentByAgentID(ctx context.Context, agentID string) (*agent.Agent, error) {
	if agentID == "" {
		return nil, goerr.New("agent ID cannot be empty")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, agent := range c.agents {
		if agent.AgentID == agentID {
			// Return a copy to avoid external modifications
			agentCopy := *agent
			return &agentCopy, nil
		}
	}

	return nil, goerr.New("agent not found", goerr.V("agent_id", agentID))
}

// UpdateAgent updates an existing agent
func (c *AgentMemoryClient) UpdateAgent(ctx context.Context, agent *agent.Agent) error {
	if agent == nil {
		return goerr.New("agent cannot be nil")
	}

	if !agent.ID.IsValid() {
		return goerr.New("invalid agent ID")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	existing, exists := c.agents[agent.ID]
	if !exists {
		return goerr.New("agent not found", goerr.V("id", agent.ID.String()))
	}

	// Check if AgentID conflicts with another agent
	for id, other := range c.agents {
		if id != agent.ID && other.AgentID == agent.AgentID {
			return goerr.New("agent ID already exists", goerr.V("agent_id", agent.AgentID))
		}
	}

	agent.UpdatedAt = time.Now()
	agent.CreatedAt = existing.CreatedAt // Preserve original creation time

	// Create a copy to avoid external modifications
	agentCopy := *agent
	c.agents[agent.ID] = &agentCopy

	return nil
}

// DeleteAgent deletes an agent and all its versions
func (c *AgentMemoryClient) DeleteAgent(ctx context.Context, id types.UUID) error {
	if !id.IsValid() {
		return goerr.New("invalid agent ID")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.agents[id]; !exists {
		return goerr.New("agent not found", goerr.V("id", id.String()))
	}

	// Delete all versions
	delete(c.versions, id)

	// Delete the agent
	delete(c.agents, id)

	return nil
}

// ListAgents retrieves a list of agents with pagination
func (c *AgentMemoryClient) ListAgents(ctx context.Context, offset, limit int) ([]*agent.Agent, int, error) {
	if offset < 0 || limit < 0 {
		return nil, 0, goerr.New("offset and limit must be non-negative")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Convert to slice and sort by creation time (newest first)
	agents := make([]*agent.Agent, 0, len(c.agents))
	for _, agent := range c.agents {
		// Create a copy to avoid external modifications
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}

	sort.Slice(agents, func(i, j int) bool {
		return agents[i].CreatedAt.After(agents[j].CreatedAt)
	})

	totalCount := len(agents)

	// Apply pagination
	start := offset
	if start > totalCount {
		start = totalCount
	}

	end := start + limit
	if limit == 0 || end > totalCount {
		end = totalCount
	}

	if start >= totalCount {
		return []*agent.Agent{}, totalCount, nil
	}

	return agents[start:end], totalCount, nil
}

// CreateAgentVersion creates a new agent version
func (c *AgentMemoryClient) CreateAgentVersion(ctx context.Context, version *agent.AgentVersion) error {
	if version == nil {
		return goerr.New("agent version cannot be nil")
	}

	if !version.AgentUUID.IsValid() {
		return goerr.New("invalid agent UUID")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if agent exists
	if _, exists := c.agents[version.AgentUUID]; !exists {
		return goerr.New("agent not found", goerr.V("agent_uuid", version.AgentUUID.String()))
	}

	// Initialize versions map for this agent if not exists
	if c.versions[version.AgentUUID] == nil {
		c.versions[version.AgentUUID] = make(map[string]*agent.AgentVersion)
	}

	// Check if version already exists
	if _, exists := c.versions[version.AgentUUID][version.Version]; exists {
		return goerr.New("version already exists",
			goerr.V("agent_uuid", version.AgentUUID.String()),
			goerr.V("version", version.Version))
	}

	now := time.Now()
	if version.CreatedAt.IsZero() {
		version.CreatedAt = now
	}
	version.UpdatedAt = now

	// Create a copy to avoid external modifications
	versionCopy := *version
	c.versions[version.AgentUUID][version.Version] = &versionCopy

	return nil
}

// GetAgentVersion retrieves a specific version of an agent
func (c *AgentMemoryClient) GetAgentVersion(ctx context.Context, agentUUID types.UUID, version string) (*agent.AgentVersion, error) {
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent UUID")
	}

	if version == "" {
		return nil, goerr.New("version cannot be empty")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	agentVersions, exists := c.versions[agentUUID]
	if !exists {
		return nil, goerr.New("agent not found", goerr.V("agent_uuid", agentUUID.String()))
	}

	agentVersion, exists := agentVersions[version]
	if !exists {
		return nil, goerr.New("agent version not found",
			goerr.V("agent_uuid", agentUUID.String()),
			goerr.V("version", version))
	}

	// Return a copy to avoid external modifications
	versionCopy := *agentVersion
	return &versionCopy, nil
}

// GetLatestAgentVersion retrieves the latest version of an agent
func (c *AgentMemoryClient) GetLatestAgentVersion(ctx context.Context, agentUUID types.UUID) (*agent.AgentVersion, error) {
	// Get the agent to find the latest version
	agent, err := c.GetAgent(ctx, agentUUID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent for latest version")
	}

	return c.GetAgentVersion(ctx, agentUUID, agent.Latest)
}

// ListAgentVersions retrieves all versions of an agent
func (c *AgentMemoryClient) ListAgentVersions(ctx context.Context, agentUUID types.UUID) ([]*agent.AgentVersion, error) {
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent UUID")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	agentVersions, exists := c.versions[agentUUID]
	if !exists {
		return []*agent.AgentVersion{}, nil // Return empty slice if no versions
	}

	// Convert to slice and sort by creation time (newest first)
	versions := make([]*agent.AgentVersion, 0, len(agentVersions))
	for _, version := range agentVersions {
		// Create a copy to avoid external modifications
		versionCopy := *version
		versions = append(versions, &versionCopy)
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].CreatedAt.After(versions[j].CreatedAt)
	})

	return versions, nil
}

// UpdateAgentVersion updates an existing agent version
func (c *AgentMemoryClient) UpdateAgentVersion(ctx context.Context, version *agent.AgentVersion) error {
	if version == nil {
		return goerr.New("agent version cannot be nil")
	}

	if !version.AgentUUID.IsValid() {
		return goerr.New("invalid agent UUID")
	}

	if version.Version == "" {
		return goerr.New("version cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	agentVersions, exists := c.versions[version.AgentUUID]
	if !exists {
		return goerr.New("agent not found", goerr.V("agent_uuid", version.AgentUUID.String()))
	}

	existing, exists := agentVersions[version.Version]
	if !exists {
		return goerr.New("agent version not found",
			goerr.V("agent_uuid", version.AgentUUID.String()),
			goerr.V("version", version.Version))
	}

	version.UpdatedAt = time.Now()
	version.CreatedAt = existing.CreatedAt // Preserve original creation time

	// Create a copy to avoid external modifications
	versionCopy := *version
	c.versions[version.AgentUUID][version.Version] = &versionCopy

	return nil
}

// ListAgentsWithLatestVersions efficiently retrieves agents and their latest versions
func (c *AgentMemoryClient) ListAgentsWithLatestVersions(ctx context.Context, offset, limit int) ([]*agent.Agent, []*agent.AgentVersion, int, error) {
	if offset < 0 || limit < 0 {
		return nil, nil, 0, goerr.New("offset and limit must be non-negative")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Convert to slice and sort by creation time (newest first)
	agents := make([]*agent.Agent, 0, len(c.agents))
	for _, agent := range c.agents {
		// Create a copy to avoid external modifications
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}

	sort.Slice(agents, func(i, j int) bool {
		return agents[i].CreatedAt.After(agents[j].CreatedAt)
	})

	totalCount := len(agents)

	// Apply pagination
	start := offset
	if start > totalCount {
		start = totalCount
	}

	end := start + limit
	if limit == 0 || end > totalCount {
		end = totalCount
	}

	if start >= totalCount {
		return []*agent.Agent{}, []*agent.AgentVersion{}, totalCount, nil
	}

	paginatedAgents := agents[start:end]

	// Get latest versions for the paginated agents
	versions := make([]*agent.AgentVersion, 0, len(paginatedAgents))
	for _, agentObj := range paginatedAgents {
		agentVersions, exists := c.versions[agentObj.ID]
		if !exists || len(agentVersions) == 0 {
			// No versions found for this agent
			versions = append(versions, nil)
			continue
		}

		// Find the latest version
		latestVersion, exists := agentVersions[agentObj.Latest]
		if !exists {
			// Latest version not found
			versions = append(versions, nil)
			continue
		}

		// Create a copy to avoid external modifications
		versionCopy := *latestVersion
		versions = append(versions, &versionCopy)
	}

	return paginatedAgents, versions, totalCount, nil
}

// AgentIDExists checks if an agent ID already exists
func (c *AgentMemoryClient) AgentIDExists(ctx context.Context, agentID string) (bool, error) {
	if agentID == "" {
		return false, goerr.New("agent ID cannot be empty")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, agent := range c.agents {
		if strings.EqualFold(agent.AgentID, agentID) {
			return true, nil
		}
	}

	return false, nil
}
