package graphql

import (
	"context"
	"log/slog"

	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	agentmodel "github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	graphql1 "github.com/m-mizutani/tamamo/pkg/domain/model/graphql"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/utils/logging"
)

// convertAgentToGraphQL converts domain Agent to GraphQL Agent
func convertAgentToGraphQL(ctx context.Context, a *agentmodel.Agent, latestVersion *agentmodel.AgentVersion, userUseCase interfaces.UserUseCases) *graphql1.Agent {
	if a == nil {
		return nil
	}

	// Fetch the author user data
	var authorUser *user.User
	if userUseCase != nil {
		u, err := userUseCase.GetUserByID(ctx, a.Author)
		if err != nil {
			// Log the error but don't fail the entire conversion
			logging.Default().Warn("Failed to fetch user data for agent author",
				slog.String("agent_id", a.ID.String()),
				slog.String("author_id", a.Author.String()),
				slog.String("error", err.Error()))
			// Create a fallback user with minimal data
			authorUser = &user.User{
				ID:          a.Author,
				SlackName:   "unknown",
				DisplayName: "Unknown User",
				Email:       "",
				CreatedAt:   a.CreatedAt,
				UpdatedAt:   a.UpdatedAt,
			}
		} else {
			authorUser = u
		}
	} else {
		// If userUseCase is nil, create a fallback user
		authorUser = &user.User{
			ID:          a.Author,
			SlackName:   "unknown",
			DisplayName: "Unknown User",
			Email:       "",
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		}
	}

	// Handle Latest field - convert to pointer
	var latest *string
	if a.Latest != "" {
		latest = &a.Latest
	}
	
	result := &graphql1.Agent{
		ID:          a.ID.String(),
		AgentID:     a.AgentID,
		Name:        a.Name,
		Description: a.Description,
		Author:      authorUser,
		Status:      convertAgentStatusToGraphQL(a.Status),
		Latest:      latest,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}

	if latestVersion != nil {
		result.LatestVersion = convertAgentVersionToGraphQL(latestVersion)
	}

	return result
}

// convertAgentVersionToGraphQL converts domain AgentVersion to GraphQL AgentVersion
func convertAgentVersionToGraphQL(v *agentmodel.AgentVersion) *graphql1.AgentVersion {
	if v == nil {
		return nil
	}

	var llmProvider *graphql1.LLMProvider
	var llmModel *string
	
	// Only set LLM fields if they have values
	if v.LLMProvider != "" {
		provider := convertLLMProviderToGraphQL(v.LLMProvider)
		llmProvider = &provider
	}
	if v.LLMModel != "" {
		llmModel = &v.LLMModel
	}
	
	return &graphql1.AgentVersion{
		AgentUUID:    v.AgentUUID.String(),
		Version:      v.Version,
		SystemPrompt: v.SystemPrompt,
		LlmProvider:  llmProvider,
		LlmModel:     llmModel,
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
	}
}

// convertAgentStatusToGraphQL converts domain Agent Status to GraphQL AgentStatus
func convertAgentStatusToGraphQL(s agentmodel.Status) graphql1.AgentStatus {
	switch s {
	case agentmodel.StatusActive:
		return graphql1.AgentStatusActive
	case agentmodel.StatusArchived:
		return graphql1.AgentStatusArchived
	default:
		logging.Default().Warn("Unknown agent status type, falling back to Active",
			slog.String("status", string(s)),
			slog.String("fallback", "Active"))
		return graphql1.AgentStatusActive // Default fallback
	}
}

// convertLLMProviderToGraphQL converts domain LLMProvider to GraphQL LLMProvider
func convertLLMProviderToGraphQL(p types.LLMProvider) graphql1.LLMProvider {
	switch p {
	case types.LLMProviderOpenAI:
		return graphql1.LLMProviderOpenai
	case types.LLMProviderClaude:
		return graphql1.LLMProviderClaude
	case types.LLMProviderGemini:
		return graphql1.LLMProviderGemini
	default:
		// Should not happen if data is properly normalized
		logging.Default().Error("Invalid LLM provider type",
			slog.String("provider", string(p)))
		return graphql1.LLMProviderOpenai // Emergency fallback
	}
}

// convertGraphQLLLMProviderToDomain converts GraphQL LLMProvider to domain LLMProvider
func convertGraphQLLLMProviderToDomain(p graphql1.LLMProvider) types.LLMProvider {
	switch p {
	case graphql1.LLMProviderOpenai:
		return types.LLMProviderOpenAI
	case graphql1.LLMProviderClaude:
		return types.LLMProviderClaude
	case graphql1.LLMProviderGemini:
		return types.LLMProviderGemini
	default:
		logging.Default().Warn("Unknown GraphQL LLM provider type, falling back to openai",
			slog.String("provider", string(p)),
			slog.String("fallback", "openai"))
		return types.LLMProviderOpenAI // Default fallback
	}
}

// convertLLMProviderToString converts GraphQL LLMProvider to string for validation
func convertLLMProviderToString(p graphql1.LLMProvider) string {
	switch p {
	case graphql1.LLMProviderOpenai:
		return "openai"
	case graphql1.LLMProviderClaude:
		return "claude"
	case graphql1.LLMProviderGemini:
		return "gemini"
	default:
		return ""
	}
}

// convertCreateAgentInputToRequest converts GraphQL input to use case request
func convertCreateAgentInputToRequest(input graphql1.CreateAgentInput) *interfaces.CreateAgentRequest {
	version := "1.0.0"
	if input.Version != nil {
		version = *input.Version
	}

	return &interfaces.CreateAgentRequest{
		AgentID:      input.AgentID,
		Name:         input.Name,
		Description:  input.Description,
		SystemPrompt: input.SystemPrompt,
		LLMProvider:  convertGraphQLLLMProviderToDomain(input.LlmProvider),
		LLMModel:     input.LlmModel,
		Version:      version,
	}
}

// convertUpdateAgentInputToRequest converts GraphQL input to use case request
func convertUpdateAgentInputToRequest(input graphql1.UpdateAgentInput) *interfaces.UpdateAgentRequest {
	req := &interfaces.UpdateAgentRequest{
		AgentID:      input.AgentID,
		Name:         input.Name,
		Description:  input.Description,
		SystemPrompt: input.SystemPrompt,
		LLMModel:     input.LlmModel,
	}

	// Convert LLM provider if provided
	if input.LlmProvider != nil {
		domainProvider := convertGraphQLLLMProviderToDomain(*input.LlmProvider)
		req.LLMProvider = &domainProvider
	}

	return req
}

// convertCreateAgentVersionInputToRequest converts GraphQL input to use case request
func convertCreateAgentVersionInputToRequest(input graphql1.CreateAgentVersionInput) *interfaces.CreateVersionRequest {
	return &interfaces.CreateVersionRequest{
		AgentUUID:    types.UUID(input.AgentUUID),
		Version:      input.Version,
		SystemPrompt: input.SystemPrompt,
		LLMProvider:  convertGraphQLLLMProviderToDomain(input.LlmProvider),
		LLMModel:     input.LlmModel,
	}
}
