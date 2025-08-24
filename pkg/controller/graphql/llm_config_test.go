package graphql_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/controller/graphql"
	graphqlmodel "github.com/m-mizutani/tamamo/pkg/domain/model/graphql"
	domainLLM "github.com/m-mizutani/tamamo/pkg/domain/model/llm"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/service/llm"
)

func TestQueryResolver_LLMConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("Get LLM configuration with multiple providers", func(t *testing.T) {
		// Setup test configuration
		config := &domainLLM.ProvidersConfig{
			Providers: map[string]domainLLM.Provider{
				"openai": {
					ID:          "openai",
					DisplayName: "OpenAI",
					Models: []domainLLM.Model{
						{
							ID:          "gpt-5-2025-08-07",
							DisplayName: "GPT-5",
							Description: "Latest OpenAI model",
						},
						{
							ID:          "gpt-5-nano-2025-08-07",
							DisplayName: "GPT-5 Nano",
							Description: "Lightweight variant",
						},
					},
				},
				"claude": {
					ID:          "claude",
					DisplayName: "Claude (Anthropic)",
					Models: []domainLLM.Model{
						{
							ID:          "claude-sonnet-4-20250514",
							DisplayName: "Claude Sonnet 4",
							Description: "Advanced reasoning",
						},
					},
				},
				"gemini": {
					ID:          "gemini",
					DisplayName: "Google Gemini",
					Models: []domainLLM.Model{
						{
							ID:          "gemini-2.5-flash",
							DisplayName: "Gemini 2.5 Flash",
							Description: "Fast and efficient",
						},
						{
							ID:          "gemini-2.0-flash",
							DisplayName: "Gemini 2.0 Flash",
							Description: "Previous generation",
						},
					},
				},
			},
			Defaults: domainLLM.DefaultConfig{
				Provider: "gemini",
				Model:    "gemini-2.0-flash",
			},
			Fallback: domainLLM.FallbackConfig{
				Enabled:  true,
				Provider: "openai",
				Model:    "gpt-5-nano-2025-08-07",
			},
		}

		// Create factory with test configuration
		credentials := map[types.LLMProvider]llm.Credential{
			types.LLMProviderOpenAI: {APIKey: "test-key"},
			types.LLMProviderClaude: {APIKey: "test-key"},
			types.LLMProviderGemini: {ProjectID: "test-project", Location: "us-central1"},
		}
		factory, err := llm.NewFactory(config, credentials)
		gt.NoError(t, err)

		// Create resolver with factory
		resolver := graphql.NewResolver(nil, nil, nil, factory)
		queryResolver := resolver.Query()

		// Execute query
		result, err := queryResolver.LlmConfig(ctx)
		gt.NoError(t, err)
		gt.NotEqual(t, result, nil)

		// Verify providers
		gt.Equal(t, len(result.Providers), 3)

		// Find and verify OpenAI provider
		var openaiProvider *graphqlmodel.LLMProviderInfo
		for _, p := range result.Providers {
			if p.ID == "openai" {
				openaiProvider = p
				break
			}
		}
		gt.NotEqual(t, openaiProvider, nil)
		gt.Equal(t, openaiProvider.DisplayName, "OpenAI")
		gt.Equal(t, len(openaiProvider.Models), 2)
		gt.Equal(t, openaiProvider.Models[0].ID, "gpt-5-2025-08-07")
		gt.Equal(t, openaiProvider.Models[0].DisplayName, "GPT-5")
		gt.Equal(t, openaiProvider.Models[0].Description, "Latest OpenAI model")

		// Find and verify Claude provider
		var claudeProvider *graphqlmodel.LLMProviderInfo
		for _, p := range result.Providers {
			if p.ID == "claude" {
				claudeProvider = p
				break
			}
		}
		gt.NotEqual(t, claudeProvider, nil)
		gt.Equal(t, claudeProvider.DisplayName, "Claude (Anthropic)")
		gt.Equal(t, len(claudeProvider.Models), 1)

		// Find and verify Gemini provider
		var geminiProvider *graphqlmodel.LLMProviderInfo
		for _, p := range result.Providers {
			if p.ID == "gemini" {
				geminiProvider = p
				break
			}
		}
		gt.NotEqual(t, geminiProvider, nil)
		gt.Equal(t, geminiProvider.DisplayName, "Google Gemini")
		gt.Equal(t, len(geminiProvider.Models), 2)

		// Verify defaults
		gt.Equal(t, result.DefaultProvider, "gemini")
		gt.Equal(t, result.DefaultModel, "gemini-2.0-flash")

		// Verify fallback
		gt.Equal(t, result.FallbackEnabled, true)
		gt.Equal(t, result.FallbackProvider, "openai")
		gt.Equal(t, result.FallbackModel, "gpt-5-nano-2025-08-07")
	})

	t.Run("Get LLM configuration without factory", func(t *testing.T) {
		// Create resolver without factory
		resolver := graphql.NewResolver(nil, nil, nil, nil)
		queryResolver := resolver.Query()

		// Execute query
		result, err := queryResolver.LlmConfig(ctx)
		gt.NotEqual(t, err, nil)
		gt.S(t, err.Error()).Contains("LLM configuration not available")
		gt.Equal(t, result, nil)
	})

	t.Run("Get LLM configuration with no defaults", func(t *testing.T) {
		// Setup test configuration without defaults
		config := &domainLLM.ProvidersConfig{
			Providers: map[string]domainLLM.Provider{
				"openai": {
					ID:          "openai",
					DisplayName: "OpenAI",
					Models: []domainLLM.Model{
						{
							ID:          "gpt-5-2025-08-07",
							DisplayName: "GPT-5",
						},
					},
				},
			},
			// No defaults or fallback
		}

		// Create factory
		credentials := map[types.LLMProvider]llm.Credential{
			types.LLMProviderOpenAI: {APIKey: "test-key"},
		}
		factory, err := llm.NewFactory(config, credentials)
		gt.NoError(t, err)

		// Create resolver with factory
		resolver := graphql.NewResolver(nil, nil, nil, factory)
		queryResolver := resolver.Query()

		// Execute query
		result, err := queryResolver.LlmConfig(ctx)
		gt.NoError(t, err)
		gt.NotEqual(t, result, nil)

		// Verify providers exist
		gt.Equal(t, len(result.Providers), 1)

		// Verify no defaults
		gt.Equal(t, result.DefaultProvider, "")
		gt.Equal(t, result.DefaultModel, "")

		// Verify fallback disabled
		gt.Equal(t, result.FallbackEnabled, false)
		gt.Equal(t, result.FallbackProvider, "")
		gt.Equal(t, result.FallbackModel, "")
	})
}
