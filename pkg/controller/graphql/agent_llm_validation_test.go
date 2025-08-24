package graphql_test

import (
	"context"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/controller/graphql"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	agentmodel "github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	graphqlmodel "github.com/m-mizutani/tamamo/pkg/domain/model/graphql"
	domainLLM "github.com/m-mizutani/tamamo/pkg/domain/model/llm"
	usermodel "github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/service/llm"
)

func TestMutationResolver_CreateAgent_LLMValidation(t *testing.T) {
	ctx := context.Background()

	// Setup LLM configuration
	config := &domainLLM.ProvidersConfig{
		Providers: map[string]domainLLM.Provider{
			"openai": {
				ID:          "openai",
				DisplayName: "OpenAI",
				Models: []domainLLM.Model{
					{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
					{ID: "gpt-5-nano-2025-08-07", DisplayName: "GPT-5 Nano"},
				},
			},
			"gemini": {
				ID:          "gemini",
				DisplayName: "Gemini",
				Models: []domainLLM.Model{
					{ID: "gemini-2.0-flash", DisplayName: "Gemini 2.0"},
				},
			},
		},
	}

	credentials := map[types.LLMProvider]llm.Credential{
		types.LLMProviderOpenAI: {APIKey: "test-key"},
		types.LLMProviderGemini: {ProjectID: "test-project", Location: "us-central1"},
	}

	factory, err := llm.NewFactory(config, credentials)
	gt.NoError(t, err)

	// Setup mock use cases
	mockAgentUseCase := &mock.AgentUseCasesMock{
		CreateAgentFunc: func(ctx context.Context, req *interfaces.CreateAgentRequest) (*agentmodel.Agent, error) {
			return &agentmodel.Agent{
				ID:          types.NewUUID(ctx),
				AgentID:     req.AgentID,
				Name:        req.Name,
				Description: *req.Description,
				Author:      types.NewUserID(ctx),
				Status:      agentmodel.StatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
		GetAgentFunc: func(ctx context.Context, id types.UUID) (*interfaces.AgentWithVersion, error) {
			return &interfaces.AgentWithVersion{
				Agent: &agentmodel.Agent{
					ID:          id,
					AgentID:     "test-agent",
					Name:        "Test Agent",
					Description: "Test Description",
					Author:      types.NewUserID(ctx),
					Status:      agentmodel.StatusActive,
					Latest:      "1.0.0",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				LatestVersion: &agentmodel.AgentVersion{
					AgentUUID:    id,
					Version:      "1.0.0",
					SystemPrompt: "Test prompt",
					LLMProvider:  "openai",
					LLMModel:     "gpt-5-2025-08-07",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			}, nil
		},
	}

	mockUserUseCase := &mock.UserUseCasesMock{
		GetUserByIDFunc: func(ctx context.Context, id types.UserID) (*usermodel.User, error) {
			return &usermodel.User{
				ID:          id,
				SlackName:   "testuser",
				DisplayName: "Test User",
				Email:       "test@example.com",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	resolver := graphql.NewResolver(nil, mockAgentUseCase, mockUserUseCase, factory)
	mutationResolver := resolver.Mutation()

	t.Run("Valid provider and model", func(t *testing.T) {
		provider := graphqlmodel.LLMProviderOpenai
		input := graphqlmodel.CreateAgentInput{
			AgentID:      "valid-agent",
			Name:         "Valid Agent",
			Description:  stringPtr("Test description"),
			SystemPrompt: stringPtr("Test prompt"),
			LlmProvider:  provider,
			LlmModel:     "gpt-5-2025-08-07",
		}

		result, err := mutationResolver.CreateAgent(ctx, input)
		gt.NoError(t, err)
		gt.NotEqual(t, result, nil)
		gt.Equal(t, result.AgentID, "test-agent")
	})

	t.Run("Invalid model for provider", func(t *testing.T) {
		provider := graphqlmodel.LLMProviderOpenai
		input := graphqlmodel.CreateAgentInput{
			AgentID:      "invalid-model-agent",
			Name:         "Invalid Model Agent",
			Description:  stringPtr("Test description"),
			SystemPrompt: stringPtr("Test prompt"),
			LlmProvider:  provider,
			LlmModel:     "invalid-model", // This model doesn't exist for OpenAI
		}

		result, err := mutationResolver.CreateAgent(ctx, input)
		gt.NotEqual(t, err, nil)
		gt.S(t, err.Error()).Contains("invalid LLM provider/model combination")
		gt.Equal(t, result, nil)
	})

	t.Run("Valid provider switching", func(t *testing.T) {
		// Test OpenAI
		providerOpenAI := graphqlmodel.LLMProviderOpenai
		inputOpenAI := graphqlmodel.CreateAgentInput{
			AgentID:      "openai-agent",
			Name:         "OpenAI Agent",
			Description:  stringPtr("OpenAI test"),
			SystemPrompt: stringPtr("Test prompt"),
			LlmProvider:  providerOpenAI,
			LlmModel:     "gpt-5-nano-2025-08-07",
		}

		result, err := mutationResolver.CreateAgent(ctx, inputOpenAI)
		gt.NoError(t, err)
		gt.NotEqual(t, result, nil)

		// Test Gemini
		providerGemini := graphqlmodel.LLMProviderGemini
		inputGemini := graphqlmodel.CreateAgentInput{
			AgentID:      "gemini-agent",
			Name:         "Gemini Agent",
			Description:  stringPtr("Gemini test"),
			SystemPrompt: stringPtr("Test prompt"),
			LlmProvider:  providerGemini,
			LlmModel:     "gemini-2.0-flash",
		}

		result, err = mutationResolver.CreateAgent(ctx, inputGemini)
		gt.NoError(t, err)
		gt.NotEqual(t, result, nil)
	})

	t.Run("Invalid provider", func(t *testing.T) {
		provider := graphqlmodel.LLMProviderClaude // Claude is not configured in our test
		input := graphqlmodel.CreateAgentInput{
			AgentID:      "claude-agent",
			Name:         "Claude Agent",
			Description:  stringPtr("Claude test"),
			SystemPrompt: stringPtr("Test prompt"),
			LlmProvider:  provider,
			LlmModel:     "claude-sonnet-4-20250514",
		}

		result, err := mutationResolver.CreateAgent(ctx, input)
		gt.NotEqual(t, err, nil)
		gt.S(t, err.Error()).Contains("invalid LLM provider/model combination")
		gt.Equal(t, result, nil)
	})
}

func TestMutationResolver_UpdateAgent_LLMValidation(t *testing.T) {
	ctx := context.Background()

	// Setup LLM configuration
	config := &domainLLM.ProvidersConfig{
		Providers: map[string]domainLLM.Provider{
			"openai": {
				ID:          "openai",
				DisplayName: "OpenAI",
				Models: []domainLLM.Model{
					{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
				},
			},
			"gemini": {
				ID:          "gemini",
				DisplayName: "Gemini",
				Models: []domainLLM.Model{
					{ID: "gemini-2.5-flash", DisplayName: "Gemini 2.5"},
				},
			},
		},
	}

	credentials := map[types.LLMProvider]llm.Credential{
		types.LLMProviderOpenAI: {APIKey: "test-key"},
		types.LLMProviderGemini: {ProjectID: "test-project", Location: "us-central1"},
	}

	factory, err := llm.NewFactory(config, credentials)
	gt.NoError(t, err)

	agentID := types.NewUUID(ctx)

	// Setup mock use cases
	mockAgentUseCase := &mock.AgentUseCasesMock{
		UpdateAgentFunc: func(ctx context.Context, id types.UUID, req *interfaces.UpdateAgentRequest) (*agentmodel.Agent, error) {
			return &agentmodel.Agent{
				ID:          id,
				AgentID:     "updated-agent",
				Name:        "Updated Agent",
				Description: "Updated Description",
				Author:      types.NewUserID(ctx),
				Status:      agentmodel.StatusActive,
				Latest:      "2.0.0",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
		GetAgentFunc: func(ctx context.Context, id types.UUID) (*interfaces.AgentWithVersion, error) {
			return &interfaces.AgentWithVersion{
				Agent: &agentmodel.Agent{
					ID:          id,
					AgentID:     "updated-agent",
					Name:        "Updated Agent",
					Description: "Updated Description",
					Author:      types.NewUserID(ctx),
					Status:      agentmodel.StatusActive,
					Latest:      "2.0.0",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				LatestVersion: &agentmodel.AgentVersion{
					AgentUUID:    id,
					Version:      "2.0.0",
					SystemPrompt: "Updated prompt",
					LLMProvider:  "gemini",
					LLMModel:     "gemini-2.5-flash",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			}, nil
		},
	}

	mockUserUseCase := &mock.UserUseCasesMock{
		GetUserByIDFunc: func(ctx context.Context, id types.UserID) (*usermodel.User, error) {
			return &usermodel.User{
				ID:          id,
				SlackName:   "testuser",
				DisplayName: "Test User",
				Email:       "test@example.com",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	resolver := graphql.NewResolver(nil, mockAgentUseCase, mockUserUseCase, factory)
	mutationResolver := resolver.Mutation()

	t.Run("Valid provider and model update", func(t *testing.T) {
		provider := graphqlmodel.LLMProviderGemini
		model := "gemini-2.5-flash"
		input := graphqlmodel.UpdateAgentInput{
			Name:        stringPtr("Updated Name"),
			LlmProvider: &provider,
			LlmModel:    &model,
		}

		result, err := mutationResolver.UpdateAgent(ctx, agentID.String(), input)
		gt.NoError(t, err)
		gt.NotEqual(t, result, nil)
		gt.Equal(t, result.Name, "Updated Agent")
	})

	t.Run("Invalid model for provider update", func(t *testing.T) {
		provider := graphqlmodel.LLMProviderOpenai
		model := "invalid-model"
		input := graphqlmodel.UpdateAgentInput{
			Name:        stringPtr("Updated Name"),
			LlmProvider: &provider,
			LlmModel:    &model,
		}

		result, err := mutationResolver.UpdateAgent(ctx, agentID.String(), input)
		gt.NotEqual(t, err, nil)
		gt.S(t, err.Error()).Contains("invalid LLM provider/model combination")
		gt.Equal(t, result, nil)
	})

	t.Run("Update with only provider (missing model)", func(t *testing.T) {
		provider := graphqlmodel.LLMProviderOpenai
		input := graphqlmodel.UpdateAgentInput{
			Name:        stringPtr("Updated Name"),
			LlmProvider: &provider,
			// Model not provided - should not validate
		}

		result, err := mutationResolver.UpdateAgent(ctx, agentID.String(), input)
		gt.NoError(t, err)
		gt.NotEqual(t, result, nil)
	})

	t.Run("Update with only model (missing provider)", func(t *testing.T) {
		model := "gpt-5-2025-08-07"
		input := graphqlmodel.UpdateAgentInput{
			Name:     stringPtr("Updated Name"),
			LlmModel: &model,
			// Provider not provided - should not validate
		}

		result, err := mutationResolver.UpdateAgent(ctx, agentID.String(), input)
		gt.NoError(t, err)
		gt.NotEqual(t, result, nil)
	})
}