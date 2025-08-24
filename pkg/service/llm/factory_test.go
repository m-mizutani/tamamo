package llm_test

import (
	"context"
	"os"
	"testing"

	"github.com/m-mizutani/gt"
	domainLLM "github.com/m-mizutani/tamamo/pkg/domain/model/llm"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/service/llm"
)

func TestFactory_ValidateProviderModel(t *testing.T) {
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
					{ID: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash"},
				},
			},
		},
		Defaults: domainLLM.DefaultConfig{
			Provider: "openai",
			Model:    "gpt-5-2025-08-07",
		},
	}

	credentials := map[types.LLMProvider]llm.Credential{
		types.LLMProviderOpenAI: {APIKey: "test-openai-key"},
		types.LLMProviderGemini: {ProjectID: "test-project", Location: "us-central1"},
	}

	t.Run("Invalid provider", func(t *testing.T) {
		factory, err := llm.NewFactory(config, credentials)
		gt.Value(t, err).Equal(nil)
		gt.Value(t, factory).NotEqual(nil)

		ctx := context.Background()
		client, err := factory.CreateClient(ctx, "invalid-provider", "some-model")
		gt.Value(t, client).Equal(nil)
		gt.Value(t, err).NotEqual(nil)
		gt.S(t, err.Error()).Contains("invalid provider/model combination")
	})

	t.Run("Invalid model for valid provider", func(t *testing.T) {
		factory, err := llm.NewFactory(config, credentials)
		gt.Value(t, err).Equal(nil)
		gt.Value(t, factory).NotEqual(nil)

		ctx := context.Background()
		client, err := factory.CreateClient(ctx, "openai", "invalid-model")
		gt.Value(t, client).Equal(nil)
		gt.Value(t, err).NotEqual(nil)
		gt.S(t, err.Error()).Contains("invalid provider/model combination")
	})

	t.Run("Missing credentials", func(t *testing.T) {
		// Create config without default provider to avoid error during factory creation
		configWithoutDefault := &domainLLM.ProvidersConfig{
			Providers: map[string]domainLLM.Provider{
				"openai": {
					ID:          "openai",
					DisplayName: "OpenAI",
					Models: []domainLLM.Model{
						{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
					},
				},
			},
			// No defaults set
		}

		// Create factory with no credentials
		emptyCredentials := map[types.LLMProvider]llm.Credential{}
		factory, err := llm.NewFactory(configWithoutDefault, emptyCredentials)
		gt.Value(t, err).Equal(nil)
		gt.Value(t, factory).NotEqual(nil)

		ctx := context.Background()
		client, err := factory.CreateClient(ctx, "openai", "gpt-5-2025-08-07")
		gt.Value(t, client).Equal(nil)
		gt.Value(t, err).NotEqual(nil)
		gt.S(t, err.Error()).Contains("no credentials configured for provider")
	})
}

func TestFactory_GetDefaultClient(t *testing.T) {
	t.Run("With default configuration", func(t *testing.T) {
		config := &domainLLM.ProvidersConfig{
			Providers: map[string]domainLLM.Provider{
				"openai": {
					ID:          "openai",
					DisplayName: "OpenAI",
					Models: []domainLLM.Model{
						{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
					},
				},
			},
			Defaults: domainLLM.DefaultConfig{
				Provider: "openai",
				Model:    "gpt-5-2025-08-07",
			},
		}

		credentials := map[types.LLMProvider]llm.Credential{
			types.LLMProviderOpenAI: {APIKey: "test-key"},
		}

		factory, err := llm.NewFactory(config, credentials)
		gt.Value(t, err).Equal(nil)

		client := factory.GetDefaultClient()
		gt.Value(t, client).NotEqual(nil)
	})

	t.Run("Without default configuration", func(t *testing.T) {
		config := &domainLLM.ProvidersConfig{
			Providers: map[string]domainLLM.Provider{
				"openai": {
					ID:          "openai",
					DisplayName: "OpenAI",
					Models: []domainLLM.Model{
						{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
					},
				},
			},
			// No defaults set
		}

		credentials := map[types.LLMProvider]llm.Credential{
			types.LLMProviderOpenAI: {APIKey: "test-key"},
		}

		factory, err := llm.NewFactory(config, credentials)
		gt.Value(t, err).Equal(nil)

		client := factory.GetDefaultClient()
		gt.Value(t, client).Equal(nil)
	})
}

func TestFactory_GetFallbackClient(t *testing.T) {
	t.Run("Fallback enabled", func(t *testing.T) {
		config := &domainLLM.ProvidersConfig{
			Providers: map[string]domainLLM.Provider{
				"openai": {
					ID:          "openai",
					DisplayName: "OpenAI",
					Models: []domainLLM.Model{
						{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
					},
				},
			},
			Fallback: domainLLM.FallbackConfig{
				Enabled:  true,
				Provider: "openai",
				Model:    "gpt-5-2025-08-07",
			},
		}

		credentials := map[types.LLMProvider]llm.Credential{
			types.LLMProviderOpenAI: {APIKey: "test-api-key"},
		}

		factory, err := llm.NewFactory(config, credentials)
		gt.Value(t, err).Equal(nil)

		ctx := context.Background()
		client, err := factory.GetFallbackClient(ctx)
		gt.Value(t, err).Equal(nil)
		gt.Value(t, client).NotEqual(nil)
	})

	t.Run("Fallback disabled", func(t *testing.T) {
		config := &domainLLM.ProvidersConfig{
			Providers: map[string]domainLLM.Provider{
				"gemini": {
					ID:          "gemini",
					DisplayName: "Gemini",
					Models: []domainLLM.Model{
						{ID: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash"},
					},
				},
			},
			Fallback: domainLLM.FallbackConfig{
				Enabled: false,
			},
		}

		credentials := map[types.LLMProvider]llm.Credential{
			types.LLMProviderGemini: {ProjectID: "test-project", Location: "us-central1"},
		}

		factory, err := llm.NewFactory(config, credentials)
		gt.Value(t, err).Equal(nil)

		ctx := context.Background()
		client, err := factory.GetFallbackClient(ctx)
		gt.Value(t, err).NotEqual(nil)
		gt.S(t, err.Error()).Contains("fallback is not enabled")
		gt.Value(t, client).Equal(nil)
	})

	t.Run("Fallback enabled but not configured", func(t *testing.T) {
		config := &domainLLM.ProvidersConfig{
			Providers: map[string]domainLLM.Provider{
				"gemini": {
					ID:          "gemini",
					DisplayName: "Gemini",
					Models: []domainLLM.Model{
						{ID: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash"},
					},
				},
			},
			Fallback: domainLLM.FallbackConfig{
				Enabled: true,
				// Provider and Model not set
			},
		}

		credentials := map[types.LLMProvider]llm.Credential{
			types.LLMProviderGemini: {ProjectID: "test-project", Location: "us-central1"},
		}

		factory, err := llm.NewFactory(config, credentials)
		gt.Value(t, err).Equal(nil)

		ctx := context.Background()
		client, err := factory.GetFallbackClient(ctx)
		gt.Value(t, err).NotEqual(nil)
		gt.S(t, err.Error()).Contains("fallback provider/model not configured")
		gt.Value(t, client).Equal(nil)
	})
}

func TestFactory_GetConfig(t *testing.T) {
	config := &domainLLM.ProvidersConfig{
		Providers: map[string]domainLLM.Provider{
			"openai": {
				ID:          "openai",
				DisplayName: "OpenAI",
				Models: []domainLLM.Model{
					{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
				},
			},
		},
	}

	credentials := map[types.LLMProvider]llm.Credential{
		types.LLMProviderOpenAI: {APIKey: "test-key"},
	}

	factory, err := llm.NewFactory(config, credentials)
	gt.Value(t, err).Equal(nil)

	returnedConfig := factory.GetConfig()
	gt.Value(t, returnedConfig).Equal(config)
}

// TestFactory_RealGeminiClient tests creating a real Gemini client with actual credentials
func TestFactory_RealGeminiClient(t *testing.T) {
	// Skip if Gemini credentials are not set
	projectID := os.Getenv("TEST_GEMINI_PROJECT_ID")
	location := os.Getenv("TEST_GEMINI_LOCATION")
	if projectID == "" || location == "" {
		t.Skip("Skipping Gemini client test: TEST_GEMINI_PROJECT_ID and TEST_GEMINI_LOCATION not set")
	}

	config := &domainLLM.ProvidersConfig{
		Providers: map[string]domainLLM.Provider{
			"gemini": {
				ID:          "gemini",
				DisplayName: "Gemini",
				Models: []domainLLM.Model{
					{ID: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash"},
				},
			},
		},
		Defaults: domainLLM.DefaultConfig{
			Provider: "gemini",
			Model:    "gemini-2.0-flash",
		},
	}

	credentials := map[types.LLMProvider]llm.Credential{
		types.LLMProviderGemini: {
			ProjectID: projectID,
			Location:  location,
		},
	}

	factory, err := llm.NewFactory(config, credentials)
	gt.Value(t, err).Equal(nil)
	gt.Value(t, factory.GetDefaultClient()).NotEqual(nil)

	// Test creating client
	ctx := context.Background()
	client, err := factory.CreateClient(ctx, "gemini", "gemini-2.0-flash")
	gt.Value(t, err).Equal(nil)
	gt.Value(t, client).NotEqual(nil)
}

// TestFactory_RealOpenAIClient tests creating a real OpenAI client with actual credentials
func TestFactory_RealOpenAIClient(t *testing.T) {
	// Skip if OpenAI API key is not set
	apiKey := os.Getenv("TEST_OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI client test: TEST_OPENAI_API_KEY not set")
	}

	config := &domainLLM.ProvidersConfig{
		Providers: map[string]domainLLM.Provider{
			"openai": {
				ID:          "openai",
				DisplayName: "OpenAI",
				Models: []domainLLM.Model{
					{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
				},
			},
		},
		Defaults: domainLLM.DefaultConfig{
			Provider: "openai",
			Model:    "gpt-5-2025-08-07",
		},
	}

	credentials := map[types.LLMProvider]llm.Credential{
		types.LLMProviderOpenAI: {APIKey: apiKey},
	}

	factory, err := llm.NewFactory(config, credentials)
	gt.Value(t, err).Equal(nil)
	gt.Value(t, factory.GetDefaultClient()).NotEqual(nil)

	// Test creating client
	ctx := context.Background()
	client, err := factory.CreateClient(ctx, "openai", "gpt-5-2025-08-07")
	gt.Value(t, err).Equal(nil)
	gt.Value(t, client).NotEqual(nil)
}

// TestFactory_RealClaudeClient tests creating a real Claude client with actual credentials
func TestFactory_RealClaudeClient(t *testing.T) {
	// Skip if Claude API key is not set
	apiKey := os.Getenv("TEST_CLAUDE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping Claude client test: TEST_CLAUDE_API_KEY not set")
	}

	config := &domainLLM.ProvidersConfig{
		Providers: map[string]domainLLM.Provider{
			"claude": {
				ID:          "claude",
				DisplayName: "Claude",
				Models: []domainLLM.Model{
					{ID: "claude-sonnet-4-20250514", DisplayName: "Claude Sonnet 4"},
				},
			},
		},
		Defaults: domainLLM.DefaultConfig{
			Provider: "claude",
			Model:    "claude-sonnet-4-20250514",
		},
	}

	credentials := map[types.LLMProvider]llm.Credential{
		types.LLMProviderClaude: {APIKey: apiKey},
	}

	factory, err := llm.NewFactory(config, credentials)
	gt.Value(t, err).Equal(nil)
	gt.Value(t, factory.GetDefaultClient()).NotEqual(nil)

	// Test creating client
	ctx := context.Background()
	client, err := factory.CreateClient(ctx, "claude", "claude-sonnet-4-20250514")
	gt.Value(t, err).Equal(nil)
	gt.Value(t, client).NotEqual(nil)
}
