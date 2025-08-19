package firestore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	firestorepb "cloud.google.com/go/firestore/apiv1/firestorepb"
)

const (
	collectionAgents      = "agents"
	subCollectionVersions = "versions"
)

// Agent Firestore document structure
type agentDoc struct {
	ID          string    `firestore:"id"`
	AgentID     string    `firestore:"agent_id"`
	Name        string    `firestore:"name"`
	Description string    `firestore:"description"`
	Author      string    `firestore:"author"`
	Latest      string    `firestore:"latest"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
}

// AgentVersion Firestore document structure
type agentVersionDoc struct {
	AgentUUID    string    `firestore:"agent_uuid"`
	Version      string    `firestore:"version"`
	SystemPrompt string    `firestore:"system_prompt"`
	LLMProvider  string    `firestore:"llm_provider"`
	LLMModel     string    `firestore:"llm_model"`
	CreatedAt    time.Time `firestore:"created_at"`
	UpdatedAt    time.Time `firestore:"updated_at"`
}

// CreateAgent creates a new agent
func (c *Client) CreateAgent(ctx context.Context, agent *agent.Agent) error {
	if agent == nil {
		return goerr.New("agent cannot be nil")
	}

	if !agent.ID.IsValid() {
		agent.ID = types.NewUUID(ctx)
	}

	now := time.Now()
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = now
	}
	agent.UpdatedAt = now

	doc := &agentDoc{
		ID:          agent.ID.String(),
		AgentID:     agent.AgentID,
		Name:        agent.Name,
		Description: agent.Description,
		Author:      agent.Author,
		Latest:      agent.Latest,
		CreatedAt:   agent.CreatedAt,
		UpdatedAt:   agent.UpdatedAt,
	}

	_, err := c.client.Collection(collectionAgents).Doc(agent.ID.String()).Set(ctx, doc)
	if err != nil {
		return goerr.Wrap(err, "failed to create agent",
			goerr.V("agent_id", agent.AgentID),
			goerr.V("id", agent.ID.String()))
	}

	return nil
}

// GetAgent retrieves an agent by ID
func (c *Client) GetAgent(ctx context.Context, id types.UUID) (*agent.Agent, error) {
	if !id.IsValid() {
		return nil, goerr.New("invalid agent ID")
	}

	doc, err := c.client.Collection(collectionAgents).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, goerr.New("agent not found", goerr.V("id", id.String()))
		}
		return nil, goerr.Wrap(err, "failed to get agent", goerr.V("id", id.String()))
	}

	var agentDoc agentDoc
	if err := doc.DataTo(&agentDoc); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal agent data", goerr.V("id", id.String()))
	}

	return &agent.Agent{
		ID:          types.UUID(agentDoc.ID),
		AgentID:     agentDoc.AgentID,
		Name:        agentDoc.Name,
		Description: agentDoc.Description,
		Author:      agentDoc.Author,
		Latest:      agentDoc.Latest,
		CreatedAt:   agentDoc.CreatedAt,
		UpdatedAt:   agentDoc.UpdatedAt,
	}, nil
}

// GetAgentByAgentID retrieves an agent by AgentID
func (c *Client) GetAgentByAgentID(ctx context.Context, agentID string) (*agent.Agent, error) {
	if agentID == "" {
		return nil, goerr.New("agent ID cannot be empty")
	}

	iter := c.client.Collection(collectionAgents).Where("agent_id", "==", agentID).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, goerr.New("agent not found", goerr.V("agent_id", agentID))
		}
		return nil, goerr.Wrap(err, "failed to query agent", goerr.V("agent_id", agentID))
	}

	var agentDoc agentDoc
	if err := doc.DataTo(&agentDoc); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal agent data", goerr.V("agent_id", agentID))
	}

	return &agent.Agent{
		ID:          types.UUID(agentDoc.ID),
		AgentID:     agentDoc.AgentID,
		Name:        agentDoc.Name,
		Description: agentDoc.Description,
		Author:      agentDoc.Author,
		Latest:      agentDoc.Latest,
		CreatedAt:   agentDoc.CreatedAt,
		UpdatedAt:   agentDoc.UpdatedAt,
	}, nil
}

// UpdateAgent updates an existing agent
func (c *Client) UpdateAgent(ctx context.Context, agent *agent.Agent) error {
	if agent == nil {
		return goerr.New("agent cannot be nil")
	}

	if !agent.ID.IsValid() {
		return goerr.New("invalid agent ID")
	}

	agent.UpdatedAt = time.Now()

	doc := &agentDoc{
		ID:          agent.ID.String(),
		AgentID:     agent.AgentID,
		Name:        agent.Name,
		Description: agent.Description,
		Author:      agent.Author,
		Latest:      agent.Latest,
		CreatedAt:   agent.CreatedAt,
		UpdatedAt:   agent.UpdatedAt,
	}

	_, err := c.client.Collection(collectionAgents).Doc(agent.ID.String()).Set(ctx, doc)
	if err != nil {
		return goerr.Wrap(err, "failed to update agent",
			goerr.V("agent_id", agent.AgentID),
			goerr.V("id", agent.ID.String()))
	}

	return nil
}

// DeleteAgent deletes an agent and all its versions
func (c *Client) DeleteAgent(ctx context.Context, id types.UUID) error {
	if !id.IsValid() {
		return goerr.New("invalid agent ID")
	}

	// Delete all versions first
	versions, err := c.client.Collection(collectionAgents).Doc(id.String()).Collection(subCollectionVersions).Documents(ctx).GetAll()
	if err != nil {
		return goerr.Wrap(err, "failed to get agent versions for deletion", goerr.V("id", id.String()))
	}

	// Use Transaction for atomic deletion instead of deprecated Batch
	err = c.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Delete all versions
		for _, version := range versions {
			if err := tx.Delete(version.Ref); err != nil {
				return err
			}
		}

		// Delete the agent document
		agentRef := c.client.Collection(collectionAgents).Doc(id.String())
		return tx.Delete(agentRef)
	})
	if err != nil {
		return goerr.Wrap(err, "failed to delete agent", goerr.V("id", id.String()))
	}

	return nil
}

// ListAgents retrieves a list of agents with pagination
func (c *Client) ListAgents(ctx context.Context, offset, limit int) ([]*agent.Agent, int, error) {
	if offset < 0 || limit < 0 {
		return nil, 0, goerr.New("offset and limit must be non-negative")
	}

	// Get total count using efficient aggregation query
	aggregationQuery := c.client.Collection(collectionAgents).NewAggregationQuery().WithCount("total")
	result, err := aggregationQuery.Get(ctx)
	if err != nil {
		return nil, 0, goerr.Wrap(err, "failed to count agents")
	}
	countValue, ok := result["total"]
	if !ok {
		return nil, 0, goerr.New("count result not found")
	}
	totalCount := int(countValue.(int64))

	// Get agents with pagination
	query := c.client.Collection(collectionAgents).OrderBy("created_at", firestore.Desc)
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var agents []*agent.Agent
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, goerr.Wrap(err, "failed to iterate agents")
		}

		var agentDoc agentDoc
		if err := doc.DataTo(&agentDoc); err != nil {
			return nil, 0, goerr.Wrap(err, "failed to unmarshal agent data")
		}

		agents = append(agents, &agent.Agent{
			ID:          types.UUID(agentDoc.ID),
			AgentID:     agentDoc.AgentID,
			Name:        agentDoc.Name,
			Description: agentDoc.Description,
			Author:      agentDoc.Author,
			Latest:      agentDoc.Latest,
			CreatedAt:   agentDoc.CreatedAt,
			UpdatedAt:   agentDoc.UpdatedAt,
		})
	}

	return agents, totalCount, nil
}

// CreateAgentVersion creates a new agent version
func (c *Client) CreateAgentVersion(ctx context.Context, version *agent.AgentVersion) error {
	if version == nil {
		return goerr.New("agent version cannot be nil")
	}

	if !version.AgentUUID.IsValid() {
		return goerr.New("invalid agent UUID")
	}

	now := time.Now()
	if version.CreatedAt.IsZero() {
		version.CreatedAt = now
	}
	version.UpdatedAt = now

	doc := &agentVersionDoc{
		AgentUUID:    version.AgentUUID.String(),
		Version:      version.Version,
		SystemPrompt: version.SystemPrompt,
		LLMProvider:  version.LLMProvider.String(),
		LLMModel:     version.LLMModel,
		CreatedAt:    version.CreatedAt,
		UpdatedAt:    version.UpdatedAt,
	}

	_, err := c.client.Collection(collectionAgents).Doc(version.AgentUUID.String()).Collection(subCollectionVersions).Doc(version.Version).Set(ctx, doc)
	if err != nil {
		return goerr.Wrap(err, "failed to create agent version",
			goerr.V("agent_uuid", version.AgentUUID.String()),
			goerr.V("version", version.Version))
	}

	return nil
}

// GetAgentVersion retrieves a specific version of an agent
func (c *Client) GetAgentVersion(ctx context.Context, agentUUID types.UUID, version string) (*agent.AgentVersion, error) {
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent UUID")
	}

	if version == "" {
		return nil, goerr.New("version cannot be empty")
	}

	doc, err := c.client.Collection(collectionAgents).Doc(agentUUID.String()).Collection(subCollectionVersions).Doc(version).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, goerr.New("agent version not found",
				goerr.V("agent_uuid", agentUUID.String()),
				goerr.V("version", version))
		}
		return nil, goerr.Wrap(err, "failed to get agent version",
			goerr.V("agent_uuid", agentUUID.String()),
			goerr.V("version", version))
	}

	var versionDoc agentVersionDoc
	if err := doc.DataTo(&versionDoc); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal agent version data",
			goerr.V("agent_uuid", agentUUID.String()),
			goerr.V("version", version))
	}

	return &agent.AgentVersion{
		AgentUUID:    types.UUID(versionDoc.AgentUUID),
		Version:      versionDoc.Version,
		SystemPrompt: versionDoc.SystemPrompt,
		LLMProvider:  agent.LLMProvider(versionDoc.LLMProvider),
		LLMModel:     versionDoc.LLMModel,
		CreatedAt:    versionDoc.CreatedAt,
		UpdatedAt:    versionDoc.UpdatedAt,
	}, nil
}

// GetLatestAgentVersion retrieves the latest version of an agent
func (c *Client) GetLatestAgentVersion(ctx context.Context, agentUUID types.UUID) (*agent.AgentVersion, error) {
	// Get the agent to find the latest version
	agentDoc, err := c.GetAgent(ctx, agentUUID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent for latest version")
	}

	return c.GetAgentVersion(ctx, agentUUID, agentDoc.Latest)
}

// ListAgentVersions retrieves all versions of an agent
func (c *Client) ListAgentVersions(ctx context.Context, agentUUID types.UUID) ([]*agent.AgentVersion, error) {
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent UUID")
	}

	iter := c.client.Collection(collectionAgents).Doc(agentUUID.String()).Collection(subCollectionVersions).OrderBy("created_at", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var versions []*agent.AgentVersion
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, goerr.Wrap(err, "failed to iterate agent versions",
				goerr.V("agent_uuid", agentUUID.String()))
		}

		var versionDoc agentVersionDoc
		if err := doc.DataTo(&versionDoc); err != nil {
			return nil, goerr.Wrap(err, "failed to unmarshal agent version data",
				goerr.V("agent_uuid", agentUUID.String()))
		}

		versions = append(versions, &agent.AgentVersion{
			AgentUUID:    types.UUID(versionDoc.AgentUUID),
			Version:      versionDoc.Version,
			SystemPrompt: versionDoc.SystemPrompt,
			LLMProvider:  agent.LLMProvider(versionDoc.LLMProvider),
			LLMModel:     versionDoc.LLMModel,
			CreatedAt:    versionDoc.CreatedAt,
			UpdatedAt:    versionDoc.UpdatedAt,
		})
	}

	return versions, nil
}

// UpdateAgentVersion updates an existing agent version
func (c *Client) UpdateAgentVersion(ctx context.Context, version *agent.AgentVersion) error {
	if version == nil {
		return goerr.New("agent version cannot be nil")
	}

	if !version.AgentUUID.IsValid() {
		return goerr.New("invalid agent UUID")
	}

	if version.Version == "" {
		return goerr.New("version cannot be empty")
	}

	version.UpdatedAt = time.Now()

	doc := &agentVersionDoc{
		AgentUUID:    version.AgentUUID.String(),
		Version:      version.Version,
		SystemPrompt: version.SystemPrompt,
		LLMProvider:  version.LLMProvider.String(),
		LLMModel:     version.LLMModel,
		CreatedAt:    version.CreatedAt,
		UpdatedAt:    version.UpdatedAt,
	}

	_, err := c.client.Collection(collectionAgents).Doc(version.AgentUUID.String()).Collection(subCollectionVersions).Doc(version.Version).Set(ctx, doc)
	if err != nil {
		return goerr.Wrap(err, "failed to update agent version",
			goerr.V("agent_uuid", version.AgentUUID.String()),
			goerr.V("version", version.Version))
	}

	return nil
}

// ListAgentsWithLatestVersions efficiently retrieves agents and their latest versions in a single operation
func (c *Client) ListAgentsWithLatestVersions(ctx context.Context, offset, limit int) ([]*agent.Agent, []*agent.AgentVersion, int, error) {
	if offset < 0 || limit < 0 {
		return nil, nil, 0, goerr.New("offset and limit must be non-negative")
	}

	// Get total count using efficient aggregation query
	aggregationQuery := c.client.Collection(collectionAgents).NewAggregationQuery().WithCount("total")
	result, err := aggregationQuery.Get(ctx)
	if err != nil {
		return nil, nil, 0, goerr.Wrap(err, "failed to count agents")
	}
	countValue, ok := result["total"]
	if !ok {
		return nil, nil, 0, goerr.New("count result not found")
	}
	
	// Handle Firestore aggregation result type conversion
	var totalCount int
	switch v := countValue.(type) {
	case int64:
		totalCount = int(v)
	case int:
		totalCount = v
	case *firestorepb.Value:
		// Extract integer value from Firestore protobuf Value
		if intVal := v.GetIntegerValue(); intVal != 0 || v.GetValueType() != nil {
			totalCount = int(intVal)
		} else {
			return nil, nil, 0, goerr.New("count value is not an integer")
		}
	default:
		return nil, nil, 0, goerr.Wrap(fmt.Errorf("unexpected count value type: %T", v), "failed to convert count result")
	}

	// Get agents with pagination
	query := c.client.Collection(collectionAgents).OrderBy("created_at", firestore.Desc)
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var agents []*agent.Agent
	var agentIDs []types.UUID
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, nil, 0, goerr.Wrap(err, "failed to iterate agents")
		}

		var agentDoc agentDoc
		if err := doc.DataTo(&agentDoc); err != nil {
			return nil, nil, 0, goerr.Wrap(err, "failed to unmarshal agent data")
		}

		agentObj := &agent.Agent{
			ID:          types.UUID(agentDoc.ID),
			AgentID:     agentDoc.AgentID,
			Name:        agentDoc.Name,
			Description: agentDoc.Description,
			Author:      agentDoc.Author,
			Latest:      agentDoc.Latest,
			CreatedAt:   agentDoc.CreatedAt,
			UpdatedAt:   agentDoc.UpdatedAt,
		}
		agents = append(agents, agentObj)
		agentIDs = append(agentIDs, agentObj.ID)
	}

	// Batch fetch latest versions for all agents
	var versions []*agent.AgentVersion
	for i, agentObj := range agents {
		// Get latest version for this agent
		versionDoc, err := c.client.Collection(collectionAgents).Doc(agentIDs[i].String()).Collection(subCollectionVersions).Doc(agentObj.Latest).Get(ctx)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// If version not found, skip this agent's version
				versions = append(versions, nil)
				continue
			}
			return nil, nil, 0, goerr.Wrap(err, "failed to get latest version", 
				goerr.V("agent_id", agentObj.AgentID),
				goerr.V("version", agentObj.Latest))
		}

		var versionDocData agentVersionDoc
		if err := versionDoc.DataTo(&versionDocData); err != nil {
			return nil, nil, 0, goerr.Wrap(err, "failed to unmarshal version data",
				goerr.V("agent_id", agentObj.AgentID),
				goerr.V("version", agentObj.Latest))
		}

		versions = append(versions, &agent.AgentVersion{
			AgentUUID:    types.UUID(versionDocData.AgentUUID),
			Version:      versionDocData.Version,
			SystemPrompt: versionDocData.SystemPrompt,
			LLMProvider:  agent.LLMProvider(versionDocData.LLMProvider),
			LLMModel:     versionDocData.LLMModel,
			CreatedAt:    versionDocData.CreatedAt,
			UpdatedAt:    versionDocData.UpdatedAt,
		})
	}

	return agents, versions, totalCount, nil
}

// AgentIDExists checks if an agent ID already exists
func (c *Client) AgentIDExists(ctx context.Context, agentID string) (bool, error) {
	if agentID == "" {
		return false, goerr.New("agent ID cannot be empty")
	}

	iter := c.client.Collection(collectionAgents).Where("agent_id", "==", agentID).Limit(1).Documents(ctx)
	defer iter.Stop()

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil // Not found
	}
	if err != nil {
		return false, goerr.Wrap(err, "failed to check agent ID existence", goerr.V("agent_id", agentID))
	}

	return true, nil // Found
}
