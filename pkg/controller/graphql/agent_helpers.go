package graphql

import (
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	agentmodel "github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	graphql1 "github.com/m-mizutani/tamamo/pkg/domain/model/graphql"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// convertAgentToGraphQL converts domain Agent to GraphQL Agent
func convertAgentToGraphQL(a *agentmodel.Agent, latestVersion *agentmodel.AgentVersion) *graphql1.Agent {
	if a == nil {
		return nil
	}

	result := &graphql1.Agent{
		ID:          a.ID.String(),
		AgentID:     a.AgentID,
		Name:        a.Name,
		Description: a.Description,
		Author:      a.Author,
		Latest:      a.Latest,
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

	return &graphql1.AgentVersion{
		AgentUUID:    v.AgentUUID.String(),
		Version:      v.Version,
		SystemPrompt: v.SystemPrompt,
		LlmProvider:  convertLLMProviderToGraphQL(v.LLMProvider),
		LlmModel:     v.LLMModel,
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
	}
}

// convertLLMProviderToGraphQL converts domain LLMProvider to GraphQL LLMProvider
func convertLLMProviderToGraphQL(p agentmodel.LLMProvider) graphql1.LLMProvider {
	switch p {
	case agentmodel.LLMProviderOpenAI:
		return graphql1.LLMProviderOpenai
	case agentmodel.LLMProviderClaude:
		return graphql1.LLMProviderClaude
	case agentmodel.LLMProviderGemini:
		return graphql1.LLMProviderGemini
	default:
		return graphql1.LLMProviderOpenai // Default fallback
	}
}

// convertGraphQLLLMProviderToDomain converts GraphQL LLMProvider to domain LLMProvider
func convertGraphQLLLMProviderToDomain(p graphql1.LLMProvider) agentmodel.LLMProvider {
	switch p {
	case graphql1.LLMProviderOpenai:
		return agentmodel.LLMProviderOpenAI
	case graphql1.LLMProviderClaude:
		return agentmodel.LLMProviderClaude
	case graphql1.LLMProviderGemini:
		return agentmodel.LLMProviderGemini
	default:
		return agentmodel.LLMProviderOpenAI // Default fallback
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
