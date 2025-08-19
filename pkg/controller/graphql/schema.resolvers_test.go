package graphql_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/controller/graphql"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	agentmodel "github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	graphqlmodel "github.com/m-mizutani/tamamo/pkg/domain/model/graphql"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// Helper function to convert string to *string
func stringPtr(s string) *string {
	return &s
}

func TestQueryResolver_Thread_Success(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testThread := &slack.Thread{
		ID:        types.NewThreadID(ctx),
		TeamID:    "T123456",
		ChannelID: "C123456",
		ThreadTS:  "1234567890.123456",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{
		GetThreadFunc: func(ctx context.Context, id types.ThreadID) (*slack.Thread, error) {
			if id == testThread.ID {
				return testThread, nil
			}
			return nil, errors.New("thread not found")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo, nil)
	queryResolver := resolver.Query()

	// Execute test
	result, err := queryResolver.Thread(ctx, string(testThread.ID))

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, result.ID, testThread.ID)
	gt.Equal(t, result.TeamID, testThread.TeamID)
	gt.Equal(t, result.ChannelID, testThread.ChannelID)
	gt.Equal(t, result.ThreadTS, testThread.ThreadTS)
}

func TestQueryResolver_Thread_InvalidID(t *testing.T) {
	ctx := context.Background()

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo, nil)
	queryResolver := resolver.Query()

	// Execute test with invalid ID
	result, err := queryResolver.Thread(ctx, "invalid-id")

	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
	gt.V(t, err.Error()).Equal("invalid thread ID")
}

func TestQueryResolver_Thread_RepositoryError(t *testing.T) {
	ctx := context.Background()
	testID := types.NewThreadID(ctx)

	// Setup mock to return error
	mockRepo := &mock.ThreadRepositoryMock{
		GetThreadFunc: func(ctx context.Context, id types.ThreadID) (*slack.Thread, error) {
			return nil, errors.New("repository error")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo, nil)
	queryResolver := resolver.Query()

	// Execute test
	result, err := queryResolver.Thread(ctx, string(testID))

	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
	// Error message contains wrapped error details
}

func TestThreadResolver_ID(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testThread := &slack.Thread{
		ID: types.NewThreadID(ctx),
	}

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo, nil)
	threadResolver := resolver.Thread()

	// Execute test
	result, err := threadResolver.ID(ctx, testThread)

	// Verify results
	gt.NoError(t, err)
	gt.Equal(t, result, string(testThread.ID))
}

func TestQueryResolver_Thread_NotFound(t *testing.T) {
	ctx := context.Background()
	testID := types.NewThreadID(ctx)

	// Setup mock to return "not found" error
	mockRepo := &mock.ThreadRepositoryMock{
		GetThreadFunc: func(ctx context.Context, id types.ThreadID) (*slack.Thread, error) {
			return nil, errors.New("thread not found")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo, nil)
	queryResolver := resolver.Query()

	// Execute test
	result, err := queryResolver.Thread(ctx, string(testID))

	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
}

func TestQueryResolver_Threads_WithNilParameters(t *testing.T) {
	ctx := context.Background()

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{
		ListThreadsFunc: func(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error) {
			// Return empty result for this test
			return []*slack.Thread{}, 0, nil
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo, nil)
	queryResolver := resolver.Query()

	// Execute test with nil parameters
	result, err := queryResolver.Threads(ctx, nil, nil)

	// Verify results (should handle nil gracefully with defaults)
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, len(result.Threads), 0)
	gt.Equal(t, result.TotalCount, 0)

	// Verify mock was called with defaults
	calls := mockRepo.ListThreadsCalls()
	gt.Equal(t, len(calls), 1)
	gt.Equal(t, calls[0].Offset, 0) // Default offset
	gt.Equal(t, calls[0].Limit, 50) // Default limit
}

func TestQueryResolver_Threads_WithValidParameters(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testThreads := []*slack.Thread{
		{
			ID:        types.NewThreadID(ctx),
			TeamID:    "T123456",
			ChannelID: "C123456",
			ThreadTS:  "1234567890.123456",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{
		ListThreadsFunc: func(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error) {
			return testThreads, len(testThreads), nil
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo, nil)
	queryResolver := resolver.Query()

	// Execute test with valid parameters
	offset := 5
	limit := 15
	result, err := queryResolver.Threads(ctx, &offset, &limit)

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, len(result.Threads), 1)
	gt.Equal(t, result.TotalCount, 1)

	// Verify mock was called with correct parameters
	calls := mockRepo.ListThreadsCalls()
	gt.Equal(t, len(calls), 1)
	gt.Equal(t, calls[0].Offset, 5)
	gt.Equal(t, calls[0].Limit, 15)
}

func TestQueryResolver_Threads_RepositoryError(t *testing.T) {
	ctx := context.Background()

	// Setup mock to return error
	mockRepo := &mock.ThreadRepositoryMock{
		ListThreadsFunc: func(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error) {
			return nil, 0, errors.New("repository error")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo, nil)
	queryResolver := resolver.Query()

	// Execute test
	offset := 0
	limit := 10
	result, err := queryResolver.Threads(ctx, &offset, &limit)

	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
}

func TestQueryResolver_Threads_LimitCapping(t *testing.T) {
	ctx := context.Background()

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{
		ListThreadsFunc: func(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error) {
			return []*slack.Thread{}, 0, nil
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo, nil)
	queryResolver := resolver.Query()

	// Execute test with excessive limit
	offset := 0
	limit := 5000 // Should be capped to 1000
	result, err := queryResolver.Threads(ctx, &offset, &limit)

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()

	// Verify mock was called with capped limit
	calls := mockRepo.ListThreadsCalls()
	gt.Equal(t, len(calls), 1)
	gt.Equal(t, calls[0].Offset, 0)
	gt.Equal(t, calls[0].Limit, 1000) // Should be capped
}

// Agent resolver tests

func TestMutationResolver_CreateAgent_Success(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testAgent := &agentmodel.Agent{
		ID:          types.NewUUID(ctx),
		AgentID:     "test-agent",
		Name:        "Test Agent",
		Description: "A test agent",
		Author:      "test-author",
		Latest:      "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	testVersion := &agentmodel.AgentVersion{
		AgentUUID:    testAgent.ID,
		Version:      "1.0.0",
		SystemPrompt: "You are a helpful assistant.",
		LLMProvider:  agentmodel.LLMProviderOpenAI,
		LLMModel:     "gpt-4",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{
		CreateAgentFunc: func(ctx context.Context, req *interfaces.CreateAgentRequest) (*agentmodel.Agent, error) {
			return testAgent, nil
		},
		GetAgentFunc: func(ctx context.Context, id types.UUID) (*interfaces.AgentWithVersion, error) {
			if id == testAgent.ID {
				return &interfaces.AgentWithVersion{
					Agent:         testAgent,
					LatestVersion: testVersion,
				}, nil
			}
			return nil, errors.New("agent not found")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	mutationResolver := resolver.Mutation()

	// Prepare input
	input := graphqlmodel.CreateAgentInput{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LlmProvider:  graphqlmodel.LLMProviderOpenai,
		LlmModel:     "gpt-4",
	}

	// Execute test
	result, err := mutationResolver.CreateAgent(ctx, input)

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, result.ID, testAgent.ID.String())
	gt.Equal(t, result.AgentID, testAgent.AgentID)
	gt.Equal(t, result.Name, testAgent.Name)
	gt.Equal(t, result.Description, testAgent.Description)
	gt.V(t, result.LatestVersion).NotNil()
	gt.Equal(t, result.LatestVersion.Version, testVersion.Version)
}

func TestMutationResolver_CreateAgent_UseCaseError(t *testing.T) {
	ctx := context.Background()

	// Setup mock to return error
	mockAgentUseCase := &mock.AgentUseCasesMock{
		CreateAgentFunc: func(ctx context.Context, req *interfaces.CreateAgentRequest) (*agentmodel.Agent, error) {
			return nil, errors.New("agent creation failed")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	mutationResolver := resolver.Mutation()

	// Prepare input
	input := graphqlmodel.CreateAgentInput{
		AgentID:      "test-agent",
		Name:         "Test Agent",
		Description:  stringPtr("A test agent"),
		SystemPrompt: stringPtr("You are a helpful assistant."),
		LlmProvider:  graphqlmodel.LLMProviderOpenai,
		LlmModel:     "gpt-4",
	}

	// Execute test
	result, err := mutationResolver.CreateAgent(ctx, input)

	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
}

func TestMutationResolver_UpdateAgent_Success(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testAgent := &agentmodel.Agent{
		ID:          types.NewUUID(ctx),
		AgentID:     "test-agent",
		Name:        "Updated Test Agent",
		Description: "An updated test agent",
		Author:      "test-author",
		Latest:      "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	testAgentWithVersion := &interfaces.AgentWithVersion{
		Agent: testAgent,
		LatestVersion: &agentmodel.AgentVersion{
			AgentUUID:    testAgent.ID,
			Version:      "1.0.0",
			SystemPrompt: "You are a helpful assistant.",
			LLMProvider:  agentmodel.LLMProviderOpenAI,
			LLMModel:     "gpt-4",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{
		UpdateAgentFunc: func(ctx context.Context, id types.UUID, req *interfaces.UpdateAgentRequest) (*agentmodel.Agent, error) {
			if id == testAgent.ID {
				return testAgent, nil
			}
			return nil, errors.New("agent not found")
		},
		GetAgentFunc: func(ctx context.Context, id types.UUID) (*interfaces.AgentWithVersion, error) {
			if id == testAgent.ID {
				return testAgentWithVersion, nil
			}
			return nil, errors.New("agent not found")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	mutationResolver := resolver.Mutation()

	// Prepare input
	newName := "Updated Test Agent"
	input := graphqlmodel.UpdateAgentInput{
		Name: &newName,
	}

	// Execute test
	result, err := mutationResolver.UpdateAgent(ctx, testAgent.ID.String(), input)

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, result.ID, testAgent.ID.String())
	gt.Equal(t, result.Name, "Updated Test Agent")
}

func TestMutationResolver_UpdateAgent_InvalidID(t *testing.T) {
	ctx := context.Background()

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	mutationResolver := resolver.Mutation()

	// Prepare input
	newName := "Updated Test Agent"
	input := graphqlmodel.UpdateAgentInput{
		Name: &newName,
	}

	// Execute test with invalid ID
	result, err := mutationResolver.UpdateAgent(ctx, "invalid-id", input)

	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
	gt.V(t, err.Error()).Equal("invalid agent ID")
}

func TestMutationResolver_DeleteAgent_Success(t *testing.T) {
	ctx := context.Background()
	testAgentID := types.NewUUID(ctx)

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{
		DeleteAgentFunc: func(ctx context.Context, id types.UUID) error {
			if id == testAgentID {
				return nil
			}
			return errors.New("agent not found")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	mutationResolver := resolver.Mutation()

	// Execute test
	result, err := mutationResolver.DeleteAgent(ctx, testAgentID.String())

	// Verify results
	gt.NoError(t, err)
	gt.Equal(t, result, true)
}

func TestMutationResolver_DeleteAgent_UseCaseError(t *testing.T) {
	ctx := context.Background()
	testAgentID := types.NewUUID(ctx)

	// Setup mock to return error
	mockAgentUseCase := &mock.AgentUseCasesMock{
		DeleteAgentFunc: func(ctx context.Context, id types.UUID) error {
			return errors.New("deletion failed")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	mutationResolver := resolver.Mutation()

	// Execute test
	result, err := mutationResolver.DeleteAgent(ctx, testAgentID.String())

	// Verify results
	gt.Error(t, err)
	gt.Equal(t, result, false)
}

func TestMutationResolver_CreateAgentVersion_Success(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testVersion := &agentmodel.AgentVersion{
		AgentUUID:    types.NewUUID(ctx),
		Version:      "1.1.0",
		SystemPrompt: "You are an improved helpful assistant.",
		LLMProvider:  agentmodel.LLMProviderClaude,
		LLMModel:     "claude-3-opus",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{
		CreateAgentVersionFunc: func(ctx context.Context, req *interfaces.CreateVersionRequest) (*agentmodel.AgentVersion, error) {
			return testVersion, nil
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	mutationResolver := resolver.Mutation()

	// Prepare input
	input := graphqlmodel.CreateAgentVersionInput{
		AgentUUID:    testVersion.AgentUUID.String(),
		Version:      "1.1.0",
		SystemPrompt: stringPtr("You are an improved helpful assistant."),
		LlmProvider:  graphqlmodel.LLMProviderClaude,
		LlmModel:     "claude-3-opus",
	}

	// Execute test
	result, err := mutationResolver.CreateAgentVersion(ctx, input)

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, result.AgentUUID, testVersion.AgentUUID.String())
	gt.Equal(t, result.Version, testVersion.Version)
	gt.Equal(t, result.SystemPrompt, testVersion.SystemPrompt)
	gt.Equal(t, result.LlmProvider, graphqlmodel.LLMProviderClaude)
	gt.Equal(t, result.LlmModel, testVersion.LLMModel)
}

func TestQueryResolver_Agent_Success(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testAgent := &agentmodel.Agent{
		ID:          types.NewUUID(ctx),
		AgentID:     "test-agent",
		Name:        "Test Agent",
		Description: "A test agent",
		Author:      "test-author",
		Latest:      "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	testAgentWithVersion := &interfaces.AgentWithVersion{
		Agent: testAgent,
		LatestVersion: &agentmodel.AgentVersion{
			AgentUUID:    testAgent.ID,
			Version:      "1.0.0",
			SystemPrompt: "You are a helpful assistant.",
			LLMProvider:  agentmodel.LLMProviderOpenAI,
			LLMModel:     "gpt-4",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{
		GetAgentFunc: func(ctx context.Context, id types.UUID) (*interfaces.AgentWithVersion, error) {
			if id == testAgent.ID {
				return testAgentWithVersion, nil
			}
			return nil, errors.New("agent not found")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	queryResolver := resolver.Query()

	// Execute test
	result, err := queryResolver.Agent(ctx, testAgent.ID.String())

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, result.ID, testAgent.ID.String())
	gt.Equal(t, result.AgentID, testAgent.AgentID)
	gt.Equal(t, result.Name, testAgent.Name)
	gt.V(t, result.LatestVersion).NotNil()
	gt.Equal(t, result.LatestVersion.Version, "1.0.0")
}

func TestQueryResolver_Agent_InvalidID(t *testing.T) {
	ctx := context.Background()

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	queryResolver := resolver.Query()

	// Execute test with invalid ID
	result, err := queryResolver.Agent(ctx, "invalid-id")

	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
	gt.V(t, err.Error()).Equal("invalid agent ID")
}

func TestQueryResolver_Agents_Success(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testAgents := []*interfaces.AgentWithVersion{
		{
			Agent: &agentmodel.Agent{
				ID:          types.NewUUID(ctx),
				AgentID:     "test-agent-1",
				Name:        "Test Agent 1",
				Description: "First test agent",
				Author:      "test-author",
				Latest:      "1.0.0",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			LatestVersion: &agentmodel.AgentVersion{
				AgentUUID:    types.NewUUID(ctx),
				Version:      "1.0.0",
				SystemPrompt: "You are a helpful assistant.",
				LLMProvider:  agentmodel.LLMProviderOpenAI,
				LLMModel:     "gpt-4",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
		{
			Agent: &agentmodel.Agent{
				ID:          types.NewUUID(ctx),
				AgentID:     "test-agent-2",
				Name:        "Test Agent 2",
				Description: "Second test agent",
				Author:      "test-author",
				Latest:      "1.1.0",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			LatestVersion: &agentmodel.AgentVersion{
				AgentUUID:    types.NewUUID(ctx),
				Version:      "1.1.0",
				SystemPrompt: "You are an improved assistant.",
				LLMProvider:  agentmodel.LLMProviderClaude,
				LLMModel:     "claude-3-opus",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}

	agentListResponse := &interfaces.AgentListResponse{
		Agents:     testAgents,
		TotalCount: 2,
	}

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{
		ListAgentsFunc: func(ctx context.Context, offset, limit int) (*interfaces.AgentListResponse, error) {
			return agentListResponse, nil
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	queryResolver := resolver.Query()

	// Execute test
	offset := 0
	limit := 10
	result, err := queryResolver.Agents(ctx, &offset, &limit)

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, len(result.Agents), 2)
	gt.Equal(t, result.TotalCount, 2)
	gt.Equal(t, result.Agents[0].AgentID, "test-agent-1")
	gt.Equal(t, result.Agents[1].AgentID, "test-agent-2")

	// Verify mock was called with correct parameters
	calls := mockAgentUseCase.ListAgentsCalls()
	gt.Equal(t, len(calls), 1)
	gt.Equal(t, calls[0].Offset, 0)
	gt.Equal(t, calls[0].Limit, 10)
}

func TestQueryResolver_Agents_DefaultPagination(t *testing.T) {
	ctx := context.Background()

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{
		ListAgentsFunc: func(ctx context.Context, offset, limit int) (*interfaces.AgentListResponse, error) {
			return &interfaces.AgentListResponse{
				Agents:     []*interfaces.AgentWithVersion{},
				TotalCount: 0,
			}, nil
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	queryResolver := resolver.Query()

	// Execute test with nil parameters
	result, err := queryResolver.Agents(ctx, nil, nil)

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()

	// Verify mock was called with defaults
	calls := mockAgentUseCase.ListAgentsCalls()
	gt.Equal(t, len(calls), 1)
	gt.Equal(t, calls[0].Offset, 0) // Default offset
	gt.Equal(t, calls[0].Limit, 50) // Default limit
}

func TestQueryResolver_CheckAgentIDAvailability_Available(t *testing.T) {
	ctx := context.Background()

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{
		CheckAgentIDAvailabilityFunc: func(ctx context.Context, agentID string) (*interfaces.AgentIDAvailability, error) {
			return &interfaces.AgentIDAvailability{
				Available: true,
				Message:   "Agent ID is available",
			}, nil
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	queryResolver := resolver.Query()

	// Execute test
	result, err := queryResolver.CheckAgentIDAvailability(ctx, "available-agent-id")

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, result.Available, true)
	gt.Equal(t, result.Message, "Agent ID is available")
}

func TestQueryResolver_CheckAgentIDAvailability_Taken(t *testing.T) {
	ctx := context.Background()

	// Setup mock
	mockAgentUseCase := &mock.AgentUseCasesMock{
		CheckAgentIDAvailabilityFunc: func(ctx context.Context, agentID string) (*interfaces.AgentIDAvailability, error) {
			return &interfaces.AgentIDAvailability{
				Available: false,
				Message:   "Agent ID is already taken",
			}, nil
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(nil, mockAgentUseCase)
	queryResolver := resolver.Query()

	// Execute test
	result, err := queryResolver.CheckAgentIDAvailability(ctx, "taken-agent-id")

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, result.Available, false)
	gt.Equal(t, result.Message, "Agent ID is already taken")
}
