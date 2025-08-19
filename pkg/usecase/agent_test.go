package usecase_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

// Helper function to convert string to *string
func stringPtr(s string) *string {
	return &s
}

func setupAgentTest(t *testing.T) (interfaces.AgentUseCases, interfaces.AgentRepository) {
	t.Helper()
	repo := memory.NewAgentMemoryClient()
	uc := usecase.NewAgentUseCases(repo)
	return uc, repo
}

func TestCreateAgent(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	req := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent for testing purposes"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	createdAgent, err := uc.CreateAgent(ctx, req)
	gt.NoError(t, err)
	gt.V(t, createdAgent).NotNil()
	gt.Equal(t, createdAgent.AgentID, req.AgentID)
	gt.Equal(t, createdAgent.Name, req.Name)
	gt.Equal(t, createdAgent.Description, *req.Description)
	gt.Equal(t, createdAgent.Latest, req.Version)
	// Note: LatestVersion is not returned in CreateAgent, only in GetAgent
}

func TestCreateAgent_InvalidAgentID(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Test with invalid AgentID (contains spaces)
	req := &interfaces.CreateAgentRequest{
		AgentID:      "invalid agent id",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	_, err := uc.CreateAgent(ctx, req)
	gt.Error(t, err)
}

func TestCreateAgent_InvalidVersion(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Test with invalid version format
	req := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "invalid-version",
	}

	_, err := uc.CreateAgent(ctx, req)
	gt.Error(t, err)
}

func TestCreateAgent_DuplicateAgentID(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	req := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	// Create first agent
	_, err := uc.CreateAgent(ctx, req)
	gt.NoError(t, err)

	// Try to create second agent with same AgentID
	_, err = uc.CreateAgent(ctx, req)
	gt.Error(t, err)
}

func TestGetAgent(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Create test agent
	req := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	createdAgent, err := uc.CreateAgent(ctx, req)
	gt.NoError(t, err)

	// Get agent
	retrievedAgent, err := uc.GetAgent(ctx, createdAgent.ID)
	gt.NoError(t, err)
	gt.V(t, retrievedAgent).NotNil()
	gt.Equal(t, retrievedAgent.Agent.ID, createdAgent.ID)
	gt.Equal(t, retrievedAgent.Agent.AgentID, createdAgent.AgentID)
	gt.Equal(t, retrievedAgent.Agent.Name, createdAgent.Name)
	gt.V(t, retrievedAgent.LatestVersion).NotNil()
}

func TestGetAgent_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Try to get non-existent agent
	nonExistentID := types.NewUUID(ctx)
	_, err := uc.GetAgent(ctx, nonExistentID)
	gt.Error(t, err)
}

func TestListAgents(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Create multiple test agents
	for i := 0; i < 3; i++ {
		req := &interfaces.CreateAgentRequest{
			AgentID:      "test-agent-" + strconv.Itoa(i),
			Name:         "Test Agent " + strconv.Itoa(i),
			Description:  stringPtr("A test agent"),
			SystemPrompt: stringPtr("You are a helpful assistant."),
			LLMProvider:  agent.LLMProviderOpenAI,
			LLMModel:     "gpt-4",
			Version:      "1.0.0",
		}
		_, err := uc.CreateAgent(ctx, req)
		gt.NoError(t, err)
	}

	// List agents
	agentList, err := uc.ListAgents(ctx, 0, 10)
	gt.NoError(t, err)
	gt.Equal(t, len(agentList.Agents), 3)
	gt.Equal(t, agentList.TotalCount, 3)
}

func TestListAgents_Pagination(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Create 5 test agents
	for i := 0; i < 5; i++ {
		req := &interfaces.CreateAgentRequest{
			AgentID:      "test-agent-" + strconv.Itoa(i),
			Name:         "Test Agent " + strconv.Itoa(i),
			Description:  stringPtr("A test agent"),
			SystemPrompt: stringPtr("You are a helpful assistant."),
			LLMProvider:  agent.LLMProviderOpenAI,
			LLMModel:     "gpt-4",
			Version:      "1.0.0",
		}
		_, err := uc.CreateAgent(ctx, req)
		gt.NoError(t, err)
	}

	// Test pagination
	agentList, err := uc.ListAgents(ctx, 2, 2)
	gt.NoError(t, err)
	gt.Equal(t, len(agentList.Agents), 2)
	gt.Equal(t, agentList.TotalCount, 5)
}

func TestUpdateAgent(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Create test agent
	req := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	createdAgent, err := uc.CreateAgent(ctx, req)
	gt.NoError(t, err)

	// Update agent
	updateReq := &interfaces.UpdateAgentRequest{
		Name:        stringPtr("Updated Test Agent"),
		Description: stringPtr("An updated test agent"),
	}

	updatedAgent, err := uc.UpdateAgent(ctx, createdAgent.ID, updateReq)
	gt.NoError(t, err)
	gt.V(t, updatedAgent).NotNil()
	gt.Equal(t, updatedAgent.Name, "Updated Test Agent")
	gt.Equal(t, updatedAgent.Description, "An updated test agent")
	gt.Equal(t, updatedAgent.AgentID, createdAgent.AgentID) // Should remain unchanged
}

func TestUpdateAgent_AgentID(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Create test agent
	req := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	createdAgent, err := uc.CreateAgent(ctx, req)
	gt.NoError(t, err)

	// Update AgentID
	updateReq := &interfaces.UpdateAgentRequest{
		AgentID: stringPtr("updated-agent-id"),
	}

	updatedAgent, err := uc.UpdateAgent(ctx, createdAgent.ID, updateReq)
	gt.NoError(t, err)
	gt.V(t, updatedAgent).NotNil()
	gt.Equal(t, updatedAgent.AgentID, "updated-agent-id")
}

func TestUpdateAgent_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Try to update non-existent agent
	nonExistentID := types.NewUUID(ctx)
	updateReq := &interfaces.UpdateAgentRequest{
		Name: stringPtr("Updated Name"),
	}

	_, err := uc.UpdateAgent(ctx, nonExistentID, updateReq)
	gt.Error(t, err)
}

func TestDeleteAgent(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Create test agent
	req := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	createdAgent, err := uc.CreateAgent(ctx, req)
	gt.NoError(t, err)

	// Delete agent
	err = uc.DeleteAgent(ctx, createdAgent.ID)
	gt.NoError(t, err)

	// Verify agent is deleted
	_, err = uc.GetAgent(ctx, createdAgent.ID)
	gt.Error(t, err)
}

func TestDeleteAgent_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Try to delete non-existent agent
	nonExistentID := types.NewUUID(ctx)
	err := uc.DeleteAgent(ctx, nonExistentID)
	gt.Error(t, err)
}

func TestCheckAgentIDAvailability(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Test available AgentID
	availability, err := uc.CheckAgentIDAvailability(ctx, "available-agent-id")
	gt.NoError(t, err)
	gt.V(t, availability.Available).Equal(true)
	gt.V(t, availability.Message).Equal("Agent ID is available")

	// Create an agent
	req := &interfaces.CreateAgentRequest{
		AgentID:      "taken-agent-id",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	_, err = uc.CreateAgent(ctx, req)
	gt.NoError(t, err)

	// Test taken AgentID
	availability, err = uc.CheckAgentIDAvailability(ctx, "taken-agent-id")
	gt.NoError(t, err)
	gt.V(t, availability.Available).Equal(false)
	gt.V(t, availability.Message).Equal("Agent ID is already taken")
}

func TestCreateAgentVersion(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Create initial agent
	createReq := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	createdAgent, err := uc.CreateAgent(ctx, createReq)
	gt.NoError(t, err)

	// Create new version
	versionReq := &interfaces.CreateVersionRequest{
		AgentUUID:    createdAgent.ID,
		Version:      "1.1.0",
		SystemPrompt: stringPtr("You are an improved helpful assistant."),
		LLMProvider:  agent.LLMProviderClaude,
		LLMModel:     "claude-3-opus",
	}

	agentVersion, err := uc.CreateAgentVersion(ctx, versionReq)
	gt.NoError(t, err)
	gt.V(t, agentVersion).NotNil()
	gt.Equal(t, agentVersion.AgentUUID, createdAgent.ID)
	gt.Equal(t, agentVersion.Version, versionReq.Version)
	gt.Equal(t, agentVersion.SystemPrompt, *versionReq.SystemPrompt)
	gt.Equal(t, agentVersion.LLMProvider, versionReq.LLMProvider)
	gt.Equal(t, agentVersion.LLMModel, versionReq.LLMModel)
}

func TestCreateAgentVersion_InvalidVersion(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Create initial agent
	createReq := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	createdAgent, err := uc.CreateAgent(ctx, createReq)
	gt.NoError(t, err)

	// Try to create version with invalid format
	versionReq := &interfaces.CreateVersionRequest{
		AgentUUID:    createdAgent.ID,
		Version:      "invalid-version",
		SystemPrompt: stringPtr("You are an improved helpful assistant."),
		LLMProvider:  agent.LLMProviderClaude,
		LLMModel:     "claude-3-opus",
	}

	_, err = uc.CreateAgentVersion(ctx, versionReq)
	gt.Error(t, err)
}

func TestGetAgentVersions(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Create initial agent
	createReq := &interfaces.CreateAgentRequest{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LLMProvider:  agent.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		Version:      "1.0.0",
	}

	createdAgent, err := uc.CreateAgent(ctx, createReq)
	gt.NoError(t, err)

	// Create additional versions
	versions := []string{"1.1.0", "1.2.0"}
	for _, version := range versions {
		versionReq := &interfaces.CreateVersionRequest{
			AgentUUID:    createdAgent.ID,
			Version:      version,
			SystemPrompt: stringPtr("Updated system prompt for " + version),
			LLMProvider:  agent.LLMProviderClaude,
			LLMModel:     "claude-3-opus",
		}
		_, err := uc.CreateAgentVersion(ctx, versionReq)
		gt.NoError(t, err)
	}

	// Get all versions
	agentVersions, err := uc.GetAgentVersions(ctx, createdAgent.ID)
	gt.NoError(t, err)
	gt.Equal(t, len(agentVersions), 3) // Initial + 2 additional versions
	
	// Verify versions are sorted by creation date (newest first)
	expectedVersions := []string{"1.2.0", "1.1.0", "1.0.0"}
	for i, version := range agentVersions {
		gt.Equal(t, version.Version, expectedVersions[i])
	}
}

func TestGetAgentVersions_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, _ := setupAgentTest(t)

	// Try to get versions for non-existent agent
	nonExistentID := types.NewUUID(ctx)
	_, err := uc.GetAgentVersions(ctx, nonExistentID)
	gt.Error(t, err)
}

