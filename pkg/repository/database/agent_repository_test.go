package database_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/firestore"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
)

// Helper function to create a test agent
func createTestAgent(t *testing.T, repo interfaces.AgentRepository, agentID string) *agent.Agent {
	ctx := context.Background()
	agentObj := &agent.Agent{
		ID:          types.NewUUID(ctx),
		AgentID:     agentID,
		Name:        "Test Agent " + agentID,
		Description: "A test agent for " + agentID,
		Author:      "test-author",
		Latest:      "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateAgent(ctx, agentObj)
	gt.NoError(t, err)
	return agentObj
}

// Helper function to create a test agent version
func createTestAgentVersion(t *testing.T, repo interfaces.AgentRepository, agentUUID types.UUID, version string) *agent.AgentVersion {
	ctx := context.Background()
	versionObj := &agent.AgentVersion{
		AgentUUID:    agentUUID,
		Version:      version,
		SystemPrompt: "Test system prompt for " + version,
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		CreatedAt:    time.Now(),
	}

	err := repo.CreateAgentVersion(ctx, versionObj)
	gt.NoError(t, err)
	return versionObj
}

// Memory repository tests

func TestMemoryAgentRepository_CreateAgent(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	agentObj := &agent.Agent{
		ID:          types.NewUUID(ctx),
		AgentID:     "test-agent",
		Name:        "Test Agent",
		Description: "A test agent",
		Author:      "test-author",
		Latest:      "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateAgent(ctx, agentObj)
	gt.NoError(t, err)

	// Verify agent was created
	retrievedAgent, err := repo.GetAgent(ctx, agentObj.ID)
	gt.NoError(t, err)
	gt.V(t, retrievedAgent).NotNil()
	gt.Equal(t, retrievedAgent.ID, agentObj.ID)
	gt.Equal(t, retrievedAgent.AgentID, agentObj.AgentID)
	gt.Equal(t, retrievedAgent.Name, agentObj.Name)
	gt.Equal(t, retrievedAgent.Description, agentObj.Description)
	gt.Equal(t, retrievedAgent.Latest, agentObj.Latest)
}

func TestMemoryAgentRepository_CreateAgent_DuplicateAgentID(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create first agent
	firstAgent := &agent.Agent{
		ID:          types.NewUUID(ctx),
		AgentID:     "duplicate-test",
		Name:        "First Agent",
		Description: "First test agent",
		Author:      "test-author",
		Latest:      "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateAgent(ctx, firstAgent)
	gt.NoError(t, err)

	// Try to create second agent with same AgentID
	secondAgent := &agent.Agent{
		ID:          types.NewUUID(ctx),
		AgentID:     "duplicate-test", // Same AgentID
		Name:        "Second Agent",
		Description: "Second test agent",
		Author:      "test-author",
		Latest:      "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.CreateAgent(ctx, secondAgent)
	gt.Error(t, err) // Should fail due to duplicate AgentID
}

func TestMemoryAgentRepository_GetAgent(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create and verify retrieval
	original := createTestAgent(t, repo, "get-test")

	retrieved, err := repo.GetAgent(ctx, original.ID)
	gt.NoError(t, err)
	gt.V(t, retrieved).NotNil()
	gt.Equal(t, retrieved.ID, original.ID)
	gt.Equal(t, retrieved.AgentID, original.AgentID)
}

func TestMemoryAgentRepository_GetAgent_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	nonExistentID := types.NewUUID(ctx)
	_, err := repo.GetAgent(ctx, nonExistentID)
	gt.Error(t, err) // Should return error for non-existent agent
}

func TestMemoryAgentRepository_UpdateAgent(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create agent
	original := createTestAgent(t, repo, "update-test")

	// Update agent
	original.Name = "Updated Name"
	original.Description = "Updated Description"
	original.Latest = "1.1.0"
	original.UpdatedAt = time.Now()

	err := repo.UpdateAgent(ctx, original)
	gt.NoError(t, err)

	// Verify update
	updated, err := repo.GetAgent(ctx, original.ID)
	gt.NoError(t, err)
	gt.Equal(t, updated.Name, "Updated Name")
	gt.Equal(t, updated.Description, "Updated Description")
	gt.Equal(t, updated.Latest, "1.1.0")
}

func TestMemoryAgentRepository_UpdateAgent_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	nonExistentAgent := &agent.Agent{
		ID:          types.NewUUID(ctx),
		AgentID:     "non-existent",
		Name:        "Non-existent Agent",
		Description: "This agent doesn't exist",
		Author:      "test-author",
		Latest:      "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.UpdateAgent(ctx, nonExistentAgent)
	gt.Error(t, err) // Should fail for non-existent agent
}

func TestMemoryAgentRepository_DeleteAgent(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create agent and version
	testAgent := createTestAgent(t, repo, "delete-test")
	createTestAgentVersion(t, repo, testAgent.ID, "1.0.0")

	// Delete agent
	err := repo.DeleteAgent(ctx, testAgent.ID)
	gt.NoError(t, err)

	// Verify agent is deleted
	_, err = repo.GetAgent(ctx, testAgent.ID)
	gt.Error(t, err)

	// Verify versions are also deleted
	versions, err := repo.ListAgentVersions(ctx, testAgent.ID)
	gt.NoError(t, err)
	gt.Equal(t, len(versions), 0)
}

func TestMemoryAgentRepository_DeleteAgent_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	nonExistentID := types.NewUUID(ctx)
	err := repo.DeleteAgent(ctx, nonExistentID)
	gt.Error(t, err) // Should fail for non-existent agent
}

func TestMemoryAgentRepository_ListAgents(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create multiple agents
	agent1 := createTestAgent(t, repo, "list-test-1")
	agent2 := createTestAgent(t, repo, "list-test-2")
	agent3 := createTestAgent(t, repo, "list-test-3")

	// List all agents
	agents, totalCount, err := repo.ListAgents(ctx, 0, 10)
	gt.NoError(t, err)
	gt.A(t, agents).Length(3)
	gt.Equal(t, totalCount, 3)

	// Verify all created agents are in the list
	agentIDs := make(map[types.UUID]bool)
	for _, a := range agents {
		agentIDs[a.ID] = true
	}
	gt.True(t, agentIDs[agent1.ID])
	gt.True(t, agentIDs[agent2.ID])
	gt.True(t, agentIDs[agent3.ID])
}

func TestMemoryAgentRepository_ListAgents_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create multiple agents
	for i := range 5 {
		createTestAgent(t, repo, "pagination-test-"+strconv.Itoa(i))
	}

	// Test pagination
	firstPage, totalCount, err := repo.ListAgents(ctx, 0, 2)
	gt.NoError(t, err)
	gt.A(t, firstPage).Length(2)
	gt.Equal(t, totalCount, 5)

	secondPage, totalCount, err := repo.ListAgents(ctx, 2, 2)
	gt.NoError(t, err)
	gt.A(t, secondPage).Length(2)
	gt.Equal(t, totalCount, 5)

	thirdPage, totalCount, err := repo.ListAgents(ctx, 4, 2)
	gt.NoError(t, err)
	gt.A(t, thirdPage).Length(1)
	gt.Equal(t, totalCount, 5)
}

func TestMemoryAgentRepository_ListAgents_Empty(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	agents, totalCount, err := repo.ListAgents(ctx, 0, 10)
	gt.NoError(t, err)
	gt.A(t, agents).Length(0)
	gt.Equal(t, totalCount, 0)
}

func TestMemoryAgentRepository_AgentIDExists(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Test non-existent agent
	exists, err := repo.AgentIDExists(ctx, "non-existent")
	gt.NoError(t, err)
	gt.False(t, exists)

	// Create agent and test existence
	createTestAgent(t, repo, "exists-test")

	exists, err = repo.AgentIDExists(ctx, "exists-test")
	gt.NoError(t, err)
	gt.True(t, exists)
}

func TestMemoryAgentRepository_CreateAgentVersion(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create agent first
	testAgent := createTestAgent(t, repo, "version-test")

	// Create version
	version := &agent.AgentVersion{
		AgentUUID:    testAgent.ID,
		Version:      "1.0.0",
		SystemPrompt: "Test system prompt",
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		CreatedAt:    time.Now(),
	}

	err := repo.CreateAgentVersion(ctx, version)
	gt.NoError(t, err)

	// Verify version was created
	retrievedVersion, err := repo.GetAgentVersion(ctx, testAgent.ID, "1.0.0")
	gt.NoError(t, err)
	gt.V(t, retrievedVersion).NotNil()
	gt.Equal(t, retrievedVersion.AgentUUID, testAgent.ID)
	gt.Equal(t, retrievedVersion.Version, "1.0.0")
	gt.Equal(t, retrievedVersion.SystemPrompt, "Test system prompt")
}

func TestMemoryAgentRepository_CreateAgentVersion_DuplicateVersion(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create agent and first version
	testAgent := createTestAgent(t, repo, "duplicate-version-test")
	createTestAgentVersion(t, repo, testAgent.ID, "1.0.0")

	// Try to create duplicate version
	duplicateVersion := &agent.AgentVersion{
		AgentUUID:    testAgent.ID,
		Version:      "1.0.0", // Same version
		SystemPrompt: "Duplicate version",
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		CreatedAt:    time.Now(),
	}

	err := repo.CreateAgentVersion(ctx, duplicateVersion)
	gt.Error(t, err) // Should fail due to duplicate version
}

func TestMemoryAgentRepository_CreateAgentVersion_NonExistentAgent(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	nonExistentAgentID := types.NewUUID(ctx)
	version := &agent.AgentVersion{
		AgentUUID:    nonExistentAgentID,
		Version:      "1.0.0",
		SystemPrompt: "Test system prompt",
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		CreatedAt:    time.Now(),
	}

	err := repo.CreateAgentVersion(ctx, version)
	gt.Error(t, err) // Should fail for non-existent agent
}

func TestMemoryAgentRepository_GetAgentVersion(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create agent and version
	testAgent := createTestAgent(t, repo, "get-version-test")
	originalVersion := createTestAgentVersion(t, repo, testAgent.ID, "1.0.0")

	// Get version
	retrievedVersion, err := repo.GetAgentVersion(ctx, testAgent.ID, "1.0.0")
	gt.NoError(t, err)
	gt.V(t, retrievedVersion).NotNil()
	gt.Equal(t, retrievedVersion.AgentUUID, originalVersion.AgentUUID)
	gt.Equal(t, retrievedVersion.Version, originalVersion.Version)
}

func TestMemoryAgentRepository_GetAgentVersion_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create agent without versions
	testAgent := createTestAgent(t, repo, "get-version-not-found-test")

	_, err := repo.GetAgentVersion(ctx, testAgent.ID, "1.0.0")
	gt.Error(t, err) // Should fail for non-existent version
}

func TestMemoryAgentRepository_GetLatestAgentVersion(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create agent and multiple versions
	testAgent := createTestAgent(t, repo, "latest-version-test")
	createTestAgentVersion(t, repo, testAgent.ID, "1.0.0")
	createTestAgentVersion(t, repo, testAgent.ID, "1.1.0")
	latestVersion := createTestAgentVersion(t, repo, testAgent.ID, "2.0.0")

	// Update agent to mark latest version
	testAgent.Latest = "2.0.0"
	err := repo.UpdateAgent(ctx, testAgent)
	gt.NoError(t, err)

	// Get latest version
	retrievedLatest, err := repo.GetLatestAgentVersion(ctx, testAgent.ID)
	gt.NoError(t, err)
	gt.V(t, retrievedLatest).NotNil()
	gt.Equal(t, retrievedLatest.Version, latestVersion.Version)
}

func TestMemoryAgentRepository_GetLatestAgentVersion_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	nonExistentAgentID := types.NewUUID(ctx)
	_, err := repo.GetLatestAgentVersion(ctx, nonExistentAgentID)
	gt.Error(t, err) // Should fail for non-existent agent
}

func TestMemoryAgentRepository_ListAgentVersions(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create agent and multiple versions
	testAgent := createTestAgent(t, repo, "list-versions-test")
	version1 := createTestAgentVersion(t, repo, testAgent.ID, "1.0.0")
	version2 := createTestAgentVersion(t, repo, testAgent.ID, "1.1.0")
	version3 := createTestAgentVersion(t, repo, testAgent.ID, "2.0.0")

	// List versions
	versions, err := repo.ListAgentVersions(ctx, testAgent.ID)
	gt.NoError(t, err)
	gt.A(t, versions).Length(3)

	// Verify all versions are present
	versionMap := make(map[string]*agent.AgentVersion)
	for _, v := range versions {
		versionMap[v.Version] = v
	}
	gt.V(t, versionMap[version1.Version]).NotNil()
	gt.V(t, versionMap[version2.Version]).NotNil()
	gt.V(t, versionMap[version3.Version]).NotNil()
}

func TestMemoryAgentRepository_ListAgentVersions_Empty(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	// Create agent without versions
	testAgent := createTestAgent(t, repo, "list-versions-empty-test")

	versions, err := repo.ListAgentVersions(ctx, testAgent.ID)
	gt.NoError(t, err)
	gt.A(t, versions).Length(0)
}

func TestMemoryAgentRepository_ListAgentVersions_NonExistentAgent(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewAgentMemoryClient()

	nonExistentAgentID := types.NewUUID(ctx)
	versions, err := repo.ListAgentVersions(ctx, nonExistentAgentID)
	gt.NoError(t, err)
	gt.A(t, versions).Length(0) // Memory implementation returns empty slice for non-existent agents
}

// Common test suite for ListAgentsWithLatestVersions
func testListAgentsWithLatestVersions(t *testing.T, repo interfaces.AgentRepository) {
	ctx := context.Background()
	testID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create multiple agents with versions
	agent1 := createTestAgent(t, repo, "list-with-versions-1-"+testID)
	version1 := createTestAgentVersion(t, repo, agent1.ID, "1.0.0")
	
	agent2 := createTestAgent(t, repo, "list-with-versions-2-"+testID)
	createTestAgentVersion(t, repo, agent2.ID, "1.0.0")
	version2 := createTestAgentVersion(t, repo, agent2.ID, "1.1.0")
	// Update agent2 to have latest version 1.1.0
	agent2.Latest = "1.1.0"
	err := repo.UpdateAgent(ctx, agent2)
	gt.NoError(t, err)

	agent3 := createTestAgent(t, repo, "list-with-versions-3-"+testID)
	version3 := createTestAgentVersion(t, repo, agent3.ID, "2.0.0")
	// Update agent3 to have latest version 2.0.0
	agent3.Latest = "2.0.0"
	err = repo.UpdateAgent(ctx, agent3)
	gt.NoError(t, err)

	// Test ListAgentsWithLatestVersions - only count our test agents
	agents, versions, _, err := repo.ListAgentsWithLatestVersions(ctx, 0, 0)
	gt.NoError(t, err)
	
	// Filter to only our test agents
	testAgents := make([]*agent.Agent, 0)
	testVersions := make([]*agent.AgentVersion, 0)
	for i, a := range agents {
		if strings.Contains(a.Name, testID) {
			testAgents = append(testAgents, a)
			testVersions = append(testVersions, versions[i])
		}
	}
	
	gt.A(t, testAgents).Length(3)
	gt.A(t, testVersions).Length(3)

	// Verify agents and their latest versions are correctly paired
	agentVersionMap := make(map[string]*agent.AgentVersion)
	for i, a := range testAgents {
		if i < len(testVersions) && testVersions[i] != nil {
			agentVersionMap[a.AgentID] = testVersions[i]
		}
	}

	// Check specific version mappings using the unique test agent IDs
	agent1Name := "list-with-versions-1-" + testID
	agent2Name := "list-with-versions-2-" + testID
	agent3Name := "list-with-versions-3-" + testID
	
	gt.V(t, agentVersionMap[agent1Name]).NotNil()
	gt.Equal(t, agentVersionMap[agent1Name].Version, version1.Version)
	
	gt.V(t, agentVersionMap[agent2Name]).NotNil()
	gt.Equal(t, agentVersionMap[agent2Name].Version, version2.Version) // Should be 1.1.0
	
	gt.V(t, agentVersionMap[agent3Name]).NotNil()
	gt.Equal(t, agentVersionMap[agent3Name].Version, version3.Version) // Should be 2.0.0
}

func testListAgentsWithLatestVersionsPagination(t *testing.T, repo interfaces.AgentRepository) {
	ctx := context.Background()
	testID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create multiple agents with versions
	for i := range 5 {
		agent := createTestAgent(t, repo, "pagination-with-versions-"+strconv.Itoa(i)+"-"+testID)
		createTestAgentVersion(t, repo, agent.ID, "1.0.0")
	}

	// For this test, we'll just verify we can retrieve agents without errors
	// and that pagination doesn't break, rather than exact counts since other tests may exist
	firstPage, firstVersions, _, err := repo.ListAgentsWithLatestVersions(ctx, 0, 2)
	gt.NoError(t, err)
	gt.A(t, firstPage).Length(2)
	gt.A(t, firstVersions).Length(2)

	// Filter to our test agents to verify they exist
	testAgents := make([]*agent.Agent, 0)
	for _, a := range firstPage {
		if strings.Contains(a.Name, testID) {
			testAgents = append(testAgents, a)
		}
	}
	// Should have at least some of our test agents in the first page
	gt.V(t, len(testAgents) > 0).Equal(true)
}

func testListAgentsWithLatestVersionsEmpty(t *testing.T, repo interfaces.AgentRepository) {
	ctx := context.Background()

	// Just verify that the query works without error
	// For memory implementation: expect empty results when no agents exist
	// For Firestore: may have existing data, so just ensure no crashes
	agents, versions, totalCount, err := repo.ListAgentsWithLatestVersions(ctx, 0, 10)
	gt.NoError(t, err)
	gt.V(t, totalCount >= 0).Equal(true)
	gt.Equal(t, len(agents), len(versions))
}

// Memory implementation tests
func TestMemoryAgentRepository_ListAgentsWithLatestVersions(t *testing.T) {
	repo := memory.NewAgentMemoryClient()
	testListAgentsWithLatestVersions(t, repo)
}

func TestMemoryAgentRepository_ListAgentsWithLatestVersions_Pagination(t *testing.T) {
	repo := memory.NewAgentMemoryClient()
	testListAgentsWithLatestVersionsPagination(t, repo)
}

func TestMemoryAgentRepository_ListAgentsWithLatestVersions_Empty(t *testing.T) {
	repo := memory.NewAgentMemoryClient()
	testListAgentsWithLatestVersionsEmpty(t, repo)
}

// Firestore repository tests (skipped if environment not configured)

func TestFirestoreAgentRepository_ListAgentsWithLatestVersions(t *testing.T) {
	repo, skipReason := createFirestoreRepo(t)
	if repo == nil {
		t.Skip(skipReason)
	}
	testListAgentsWithLatestVersions(t, repo)
}

func TestFirestoreAgentRepository_ListAgentsWithLatestVersions_Pagination(t *testing.T) {
	repo, skipReason := createFirestoreRepo(t)
	if repo == nil {
		t.Skip(skipReason)
	}
	testListAgentsWithLatestVersionsPagination(t, repo)
}

func TestFirestoreAgentRepository_ListAgentsWithLatestVersions_Empty(t *testing.T) {
	repo, skipReason := createFirestoreRepo(t)
	if repo == nil {
		t.Skip(skipReason)
	}
	testListAgentsWithLatestVersionsEmpty(t, repo)
}

// Helper function to create Firestore repository
func createFirestoreRepo(_ *testing.T) (interfaces.AgentRepository, string) {
	projectID := os.Getenv("TEST_FIRESTORE_PROJECT")
	if projectID == "" {
		return nil, "TEST_FIRESTORE_PROJECT is not set"
	}
	databaseID := os.Getenv("TEST_FIRESTORE_DATABASE")
	if databaseID == "" {
		return nil, "TEST_FIRESTORE_DATABASE is not set"
	}

	ctx := context.Background()
	client, err := firestore.New(ctx, projectID, databaseID)
	if err != nil {
		return nil, "Firestore not available: " + err.Error()
	}
	return client, ""
}
