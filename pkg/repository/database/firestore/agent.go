package firestore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	firestorepb "cloud.google.com/go/firestore/apiv1/firestorepb"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// extractCountFromAggregation safely extracts an integer count from Firestore aggregation result
func extractCountFromAggregation(value interface{}) (int, error) {
	switch v := value.(type) {
	case int64:
		return int(v), nil
	case *firestorepb.Value:
		if intVal := v.GetIntegerValue(); intVal != 0 {
			return int(intVal), nil
		}
		return 0, nil
	case int:
		return v, nil
	default:
		return 0, goerr.New("unexpected aggregation result type", goerr.V("type", fmt.Sprintf("%T", value)))
	}
}

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
	Status      string    `firestore:"status"`
	Latest      string    `firestore:"latest"`
	ImageID     *string   `firestore:"image_id,omitempty"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
}

// toAgent converts agentDoc to domain Agent
func (d *agentDoc) toAgent() *agent.Agent {
	// Set default status for backward compatibility
	status := agent.Status(d.Status)
	if status == "" {
		status = agent.StatusActive
	}

	// Convert ImageID from *string to *types.UUID
	var imageID *types.UUID
	if d.ImageID != nil && *d.ImageID != "" {
		uuid := types.UUID(*d.ImageID)
		if uuid.IsValid() {
			imageID = &uuid
		}
	}

	return &agent.Agent{
		ID:          types.UUID(d.ID),
		AgentID:     d.AgentID,
		Name:        d.Name,
		Description: d.Description,
		Author:      types.UserID(d.Author),
		Status:      status,
		Latest:      d.Latest,
		ImageID:     imageID,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// toAgentDoc converts domain Agent to agentDoc for Firestore storage
func toAgentDoc(agentObj *agent.Agent) *agentDoc {
	// Convert ImageID from *types.UUID to *string
	var imageID *string
	if agentObj.ImageID != nil {
		idStr := agentObj.ImageID.String()
		imageID = &idStr
	}

	return &agentDoc{
		ID:          agentObj.ID.String(),
		AgentID:     agentObj.AgentID,
		Name:        agentObj.Name,
		Description: agentObj.Description,
		Author:      agentObj.Author.String(),
		Status:      agentObj.Status.String(),
		Latest:      agentObj.Latest,
		ImageID:     imageID,
		CreatedAt:   agentObj.CreatedAt,
		UpdatedAt:   agentObj.UpdatedAt,
	}
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
func (c *Client) CreateAgent(ctx context.Context, agentObj *agent.Agent) error {
	if agentObj == nil {
		return goerr.New("agent cannot be nil")
	}

	if !agentObj.ID.IsValid() {
		agentObj.ID = types.NewUUID(ctx)
	}

	now := time.Now()
	if agentObj.CreatedAt.IsZero() {
		agentObj.CreatedAt = now
	}
	agentObj.UpdatedAt = now

	// Set default status if not specified
	if agentObj.Status == "" {
		agentObj.Status = agent.StatusActive
	}

	doc := toAgentDoc(agentObj)

	_, err := c.client.Collection(collectionAgents).Doc(agentObj.ID.String()).Set(ctx, doc)
	if err != nil {
		return goerr.Wrap(err, "failed to create agent",
			goerr.V("agent_id", agentObj.AgentID),
			goerr.V("id", agentObj.ID.String()))
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

	return agentDoc.toAgent(), nil
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

	return agentDoc.toAgent(), nil
}

// UpdateAgent updates an existing agent
func (c *Client) UpdateAgent(ctx context.Context, agentObj *agent.Agent) error {
	if agentObj == nil {
		return goerr.New("agent cannot be nil")
	}

	if !agentObj.ID.IsValid() {
		return goerr.New("invalid agent ID")
	}

	agentObj.UpdatedAt = time.Now()

	// Set default status if not specified
	if agentObj.Status == "" {
		agentObj.Status = agent.StatusActive
	}

	doc := toAgentDoc(agentObj)

	_, err := c.client.Collection(collectionAgents).Doc(agentObj.ID.String()).Set(ctx, doc)
	if err != nil {
		return goerr.Wrap(err, "failed to update agent",
			goerr.V("agent_id", agentObj.AgentID),
			goerr.V("id", agentObj.ID.String()))
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
	totalCount, err := extractCountFromAggregation(countValue)
	if err != nil {
		return nil, 0, goerr.Wrap(err, "failed to extract count from aggregation result")
	}

	// Get agents with pagination (using __name__ for sorting to avoid composite index)
	query := c.client.Collection(collectionAgents).OrderBy(firestore.DocumentID, firestore.Asc)
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

		agents = append(agents, agentDoc.toAgent())
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

	// Ensure provider is normalized before saving
	normalizedProvider := types.LLMProviderFromString(string(version.LLMProvider))

	doc := &agentVersionDoc{
		AgentUUID:    version.AgentUUID.String(),
		Version:      version.Version,
		SystemPrompt: version.SystemPrompt,
		LLMProvider:  normalizedProvider.String(),
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

	// Normalize provider to ensure lowercase format
	provider := types.LLMProviderFromString(versionDoc.LLMProvider)

	return &agent.AgentVersion{
		AgentUUID:    types.UUID(versionDoc.AgentUUID),
		Version:      versionDoc.Version,
		SystemPrompt: versionDoc.SystemPrompt,
		LLMProvider:  provider,
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
			LLMProvider:  types.LLMProviderFromString(versionDoc.LLMProvider),
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

	// Ensure provider is normalized before saving
	normalizedProvider := types.LLMProviderFromString(string(version.LLMProvider))

	doc := &agentVersionDoc{
		AgentUUID:    version.AgentUUID.String(),
		Version:      version.Version,
		SystemPrompt: version.SystemPrompt,
		LLMProvider:  normalizedProvider.String(),
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

		agentObj := agentDoc.toAgent()
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
			LLMProvider:  types.LLMProviderFromString(versionDocData.LLMProvider),
			LLMModel:     versionDocData.LLMModel,
			CreatedAt:    versionDocData.CreatedAt,
			UpdatedAt:    versionDocData.UpdatedAt,
		})
	}

	return agents, versions, totalCount, nil
}

// ListActiveAgentsWithLatestVersions efficiently retrieves active agents and their latest versions in a single operation
func (c *Client) ListActiveAgentsWithLatestVersions(ctx context.Context, offset, limit int) ([]*agent.Agent, []*agent.AgentVersion, int, error) {
	if offset < 0 || limit < 0 {
		return nil, nil, 0, goerr.New("offset and limit must be non-negative")
	}

	// Get total count of active agents using efficient aggregation query
	baseQuery := c.client.Collection(collectionAgents).
		Where("status", "==", agent.StatusActive.String())
	aggregationQuery := baseQuery.NewAggregationQuery().WithCount("total")
	result, err := aggregationQuery.Get(ctx)
	if err != nil {
		return nil, nil, 0, goerr.Wrap(err, "failed to count active agents")
	}

	countValue, ok := result["total"]
	if !ok {
		return nil, nil, 0, goerr.New("count result not found")
	}

	totalCount, err := extractCountFromAggregation(countValue)
	if err != nil {
		return nil, nil, 0, goerr.Wrap(err, "failed to extract count from aggregation result")
	}

	// Get active agents with pagination
	query := c.client.Collection(collectionAgents).
		Where("status", "==", agent.StatusActive.String()).
		OrderBy(firestore.DocumentID, firestore.Asc)
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
			return nil, nil, 0, goerr.Wrap(err, "failed to iterate active agents")
		}

		var agentDoc agentDoc
		if err := doc.DataTo(&agentDoc); err != nil {
			return nil, nil, 0, goerr.Wrap(err, "failed to unmarshal agent data")
		}

		agentObj := agentDoc.toAgent()
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
			LLMProvider:  types.LLMProviderFromString(versionDocData.LLMProvider),
			LLMModel:     versionDocData.LLMModel,
			CreatedAt:    versionDocData.CreatedAt,
			UpdatedAt:    versionDocData.UpdatedAt,
		})
	}

	return agents, versions, totalCount, nil
}

// ListAgentsByStatusWithLatestVersions efficiently retrieves agents with specific status and their latest versions
func (c *Client) ListAgentsByStatusWithLatestVersions(ctx context.Context, agentStatus agent.Status, offset, limit int) ([]*agent.Agent, []*agent.AgentVersion, int, error) {
	if offset < 0 || limit < 0 {
		return nil, nil, 0, goerr.New("offset and limit must be non-negative")
	}

	// Get total count of agents with the specified status using efficient aggregation query
	baseQuery := c.client.Collection(collectionAgents).
		Where("status", "==", agentStatus.String())
	aggregationQuery := baseQuery.NewAggregationQuery().WithCount("total")
	result, err := aggregationQuery.Get(ctx)
	if err != nil {
		return nil, nil, 0, goerr.Wrap(err, "failed to count agents by status",
			goerr.V("status", agentStatus))
	}

	countValue, ok := result["total"]
	if !ok {
		return nil, nil, 0, goerr.New("count result not found")
	}

	totalCount, err := extractCountFromAggregation(countValue)
	if err != nil {
		return nil, nil, 0, goerr.Wrap(err, "failed to extract count from aggregation result")
	}

	// Get agents with specified status and pagination
	query := c.client.Collection(collectionAgents).
		Where("status", "==", agentStatus.String()).
		OrderBy(firestore.DocumentID, firestore.Asc)
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
			return nil, nil, 0, goerr.Wrap(err, "failed to iterate agents by status",
				goerr.V("status", agentStatus))
		}

		var agentDoc agentDoc
		if err := doc.DataTo(&agentDoc); err != nil {
			return nil, nil, 0, goerr.Wrap(err, "failed to unmarshal agent data")
		}

		agentObj := agentDoc.toAgent()
		agentObj.Status = agentStatus // Override with the filtered status
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
			LLMProvider:  types.LLMProviderFromString(versionDocData.LLMProvider),
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

// UpdateAgentStatus updates the status of an agent
func (c *Client) UpdateAgentStatus(ctx context.Context, id types.UUID, agentStatus agent.Status) error {
	if !id.IsValid() {
		return goerr.New("invalid agent ID")
	}

	agentRef := c.client.Collection(collectionAgents).Doc(id.String())

	_, err := agentRef.Update(ctx, []firestore.Update{
		{Path: "status", Value: agentStatus.String()},
		{Path: "updated_at", Value: time.Now()},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return goerr.New("agent not found", goerr.V("id", id.String()))
		}
		return goerr.Wrap(err, "failed to update agent status",
			goerr.V("id", id.String()),
			goerr.V("status", agentStatus.String()))
	}

	return nil
}

// ListActiveAgents retrieves a list of active agents with pagination
func (c *Client) ListActiveAgents(ctx context.Context, offset, limit int) ([]*agent.Agent, int, error) {
	if offset < 0 || limit < 0 {
		return nil, 0, goerr.New("offset and limit must be non-negative")
	}

	// Get total count of active agents using efficient aggregation query
	baseQuery := c.client.Collection(collectionAgents).
		Where("status", "==", agent.StatusActive.String())
	aggregationQuery := baseQuery.NewAggregationQuery().WithCount("total")
	result, err := aggregationQuery.Get(ctx)
	if err != nil {
		return nil, 0, goerr.Wrap(err, "failed to count active agents")
	}
	countValue, ok := result["total"]
	if !ok {
		return nil, 0, goerr.New("count result not found")
	}
	totalCount, err := extractCountFromAggregation(countValue)
	if err != nil {
		return nil, 0, goerr.Wrap(err, "failed to extract count from aggregation result")
	}

	// Get active agents with pagination (using __name__ for sorting to avoid composite index)
	query := c.client.Collection(collectionAgents).
		Where("status", "==", agent.StatusActive.String()).
		OrderBy(firestore.DocumentID, firestore.Asc)
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
			return nil, 0, goerr.Wrap(err, "failed to iterate active agents")
		}

		var agentDoc agentDoc
		if err := doc.DataTo(&agentDoc); err != nil {
			return nil, 0, goerr.Wrap(err, "failed to unmarshal agent data")
		}

		agents = append(agents, agentDoc.toAgent())
	}

	return agents, totalCount, nil
}

// ListAgentsByStatus retrieves a list of agents with specific status and pagination
func (c *Client) ListAgentsByStatus(ctx context.Context, status agent.Status, offset, limit int) ([]*agent.Agent, int, error) {
	if offset < 0 || limit < 0 {
		return nil, 0, goerr.New("offset and limit must be non-negative")
	}

	// Get total count of agents with the specified status using efficient aggregation query
	baseQuery := c.client.Collection(collectionAgents).
		Where("status", "==", status.String())
	aggregationQuery := baseQuery.NewAggregationQuery().WithCount("total")
	result, err := aggregationQuery.Get(ctx)
	if err != nil {
		return nil, 0, goerr.Wrap(err, "failed to count agents by status",
			goerr.V("status", status))
	}
	countValue, ok := result["total"]
	if !ok {
		return nil, 0, goerr.New("count result not found")
	}
	totalCount, err := extractCountFromAggregation(countValue)
	if err != nil {
		return nil, 0, goerr.Wrap(err, "failed to extract count from aggregation result")
	}

	// Get agents with specified status and pagination (using __name__ for sorting to avoid composite index)
	query := c.client.Collection(collectionAgents).
		Where("status", "==", status.String()).
		OrderBy(firestore.DocumentID, firestore.Asc)
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
			return nil, 0, goerr.Wrap(err, "failed to iterate agents by status",
				goerr.V("status", status))
		}

		var agentDoc agentDoc
		if err := doc.DataTo(&agentDoc); err != nil {
			return nil, 0, goerr.Wrap(err, "failed to unmarshal agent data")
		}

		agentObj := agentDoc.toAgent()
		agentObj.Status = status // Override with the filtered status
		agents = append(agents, agentObj)
	}

	return agents, totalCount, nil
}

// GetAgentByAgentIDActive retrieves an active agent by AgentID
func (c *Client) GetAgentByAgentIDActive(ctx context.Context, agentID string) (*agent.Agent, error) {
	if agentID == "" {
		return nil, goerr.New("agent ID cannot be empty")
	}

	iter := c.client.Collection(collectionAgents).
		Where("agent_id", "==", agentID).
		Where("status", "==", agent.StatusActive.String()).
		Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			// Check if agent exists but is archived
			checkIter := c.client.Collection(collectionAgents).
				Where("agent_id", "==", agentID).
				Limit(1).Documents(ctx)
			defer checkIter.Stop()

			_, checkErr := checkIter.Next()
			if checkErr == iterator.Done {
				return nil, goerr.New("agent not found", goerr.V("agent_id", agentID))
			} else {
				return nil, goerr.New("agent is archived", goerr.V("agent_id", agentID))
			}
		}
		return nil, goerr.Wrap(err, "failed to query active agent", goerr.V("agent_id", agentID))
	}

	var agentDoc agentDoc
	if err := doc.DataTo(&agentDoc); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal agent data", goerr.V("agent_id", agentID))
	}

	return agentDoc.toAgent(), nil
}
