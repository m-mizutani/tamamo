package config_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/cli/config"
	"github.com/urfave/cli/v3"
)

func TestLLMConfig_LoadAndValidate(t *testing.T) {
	t.Run("Load valid YAML configuration", func(t *testing.T) {
		// Use testdata file
		configPath := "testdata/valid_providers.yaml"

		// Load configuration
		llmConfig := &config.LLMConfig{
			ProvidersFile: configPath,
			// Set credentials for providers used in defaults/fallback
			GeminiProject:  "test-project",
			GeminiLocation: "us-central1",
			OpenAIAPIKey:   "test-openai-key",
		}
		providersConfig, err := llmConfig.LoadAndValidate()
		gt.NoError(t, err)
		gt.NotEqual(t, providersConfig, nil)

		// Verify providers
		gt.Equal(t, len(providersConfig.Providers), 3)

		// Check provider directly from map first
		openaiFromMap, hasOpenAI := providersConfig.Providers["openai"]
		gt.Equal(t, hasOpenAI, true)
		gt.Equal(t, openaiFromMap.DisplayName, "OpenAI")

		// Verify OpenAI provider via GetProvider
		openai, exists := providersConfig.GetProvider("openai")
		gt.Equal(t, exists, true)
		gt.NotEqual(t, openai, nil)
		gt.Equal(t, openai.DisplayName, "OpenAI")
		gt.Equal(t, len(openai.Models), 2)

		// Verify defaults
		gt.Equal(t, providersConfig.Defaults.Provider, "gemini")
		gt.Equal(t, providersConfig.Defaults.Model, "gemini-2.0-flash")

		// Verify fallback
		gt.Equal(t, providersConfig.Fallback.Enabled, true)
		gt.Equal(t, providersConfig.Fallback.Provider, "openai")
		gt.Equal(t, providersConfig.Fallback.Model, "gpt-5-nano-2025-08-07")
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		llmConfig := &config.LLMConfig{
			ProvidersFile: "/non/existent/file.yaml",
		}
		providersConfig, err := llmConfig.LoadAndValidate()
		gt.NotEqual(t, err, nil)
		gt.Equal(t, providersConfig, nil)
	})

	t.Run("Load invalid YAML", func(t *testing.T) {
		// Use testdata file
		configPath := "testdata/invalid_providers.yaml"

		llmConfig := &config.LLMConfig{
			ProvidersFile: configPath,
		}
		providersConfig, err := llmConfig.LoadAndValidate()
		gt.NotEqual(t, err, nil)
		gt.Equal(t, providersConfig, nil)
	})

	t.Run("Empty configuration file", func(t *testing.T) {
		// Use testdata file
		configPath := "testdata/empty_providers.yaml"

		llmConfig := &config.LLMConfig{
			ProvidersFile: configPath,
		}
		providersConfig, err := llmConfig.LoadAndValidate()
		gt.NoError(t, err)
		gt.NotEqual(t, providersConfig, nil)
		gt.Equal(t, len(providersConfig.Providers), 0)
	})
}

func TestLLMConfig_BuildFactory(t *testing.T) {
	t.Run("Build factory with all credentials", func(t *testing.T) {
		// Setup environment variables
		t.Setenv("TAMAMO_OPENAI_API_KEY", "test-openai-key")
		t.Setenv("TAMAMO_CLAUDE_API_KEY", "test-claude-key")
		t.Setenv("TAMAMO_GEMINI_PROJECT_ID", "test-project")
		t.Setenv("TAMAMO_GEMINI_LOCATION", "us-central1")

		// Use testdata file
		configPath := "testdata/valid_providers.yaml"

		// Load configuration and build factory
		llmConfig := &config.LLMConfig{
			ProvidersFile: configPath,
			// Set credentials for all providers
			OpenAIAPIKey:   "test-openai-key",
			ClaudeAPIKey:   "test-claude-key",
			GeminiProject:  "test-project",
			GeminiLocation: "us-central1",
		}
		providersConfig, err := llmConfig.LoadAndValidate()
		gt.NoError(t, err)

		ctx := context.Background()
		factory, err := llmConfig.BuildFactory(ctx, providersConfig)
		gt.NoError(t, err)
		gt.NotEqual(t, factory, nil)

		// Verify factory has configuration
		gt.NotEqual(t, factory.GetConfig(), nil)
		gt.NotEqual(t, factory.GetDefaultClient(), nil)
	})

	t.Run("Build factory with missing credentials", func(t *testing.T) {
		// Clear environment variables by not setting them
		// t.Setenv automatically cleans up after test

		// Use testdata file
		configPath := "testdata/valid_providers.yaml"

		// Load configuration
		llmConfig := &config.LLMConfig{
			ProvidersFile: configPath,
			// Set credentials for providers used in defaults/fallback
			GeminiProject:  "test-project",
			GeminiLocation: "us-central1",
			OpenAIAPIKey:   "test-openai-key",
		}
		providersConfig, err := llmConfig.LoadAndValidate()
		gt.NoError(t, err)

		// Build factory should succeed even without credentials
		// (credentials are only checked when creating clients)
		ctx := context.Background()
		factory, err := llmConfig.BuildFactory(ctx, providersConfig)
		gt.NoError(t, err)
		gt.NotEqual(t, factory, nil)
	})
}

func TestLLMConfig_Flags(t *testing.T) {
	llmConfig := &config.LLMConfig{}
	flags := llmConfig.Flags()

	gt.NotEqual(t, flags, nil)
	gt.Equal(t, len(flags), 9) // Now we have 9 flags

	// Check that the first flag is the llm-config flag
	flag := flags[0]
	gt.NotEqual(t, flag, nil)

	// Check that it's a StringFlag
	stringFlag, ok := flag.(*cli.StringFlag)
	gt.Equal(t, ok, true)
	gt.NotEqual(t, stringFlag, nil)

	// Verify all flags are StringFlags
	for _, f := range flags {
		_, ok := f.(*cli.StringFlag)
		gt.Equal(t, ok, true)
	}
}

func TestLLMConfig_ValidateProviderModel(t *testing.T) {
	// Use testdata file
	configPath := "testdata/valid_providers.yaml"

	llmConfig := &config.LLMConfig{
		ProvidersFile: configPath,
		// Provide credentials for default provider (Gemini)
		GeminiProject:  "test-project",
		GeminiLocation: "us-central1",
		// Also provide OpenAI credentials for fallback
		OpenAIAPIKey: "test-key",
	}
	providersConfig, err := llmConfig.LoadAndValidate()
	gt.NoError(t, err)

	t.Run("Valid provider and model", func(t *testing.T) {
		valid := providersConfig.ValidateProviderModel("openai", "gpt-5-2025-08-07")
		gt.Equal(t, valid, true)

		valid = providersConfig.ValidateProviderModel("gemini", "gemini-2.5-flash")
		gt.Equal(t, valid, true)

		// Test Claude provider from testdata
		valid = providersConfig.ValidateProviderModel("claude", "claude-sonnet-4-20250514")
		gt.Equal(t, valid, true)
	})

	t.Run("Invalid provider", func(t *testing.T) {
		// Use a provider that doesn't exist in testdata
		valid := providersConfig.ValidateProviderModel("anthropic", "some-model")
		gt.Equal(t, valid, false)
	})

	t.Run("Invalid model for valid provider", func(t *testing.T) {
		valid := providersConfig.ValidateProviderModel("openai", "gpt-4")
		gt.Equal(t, valid, false)
	})
}

func TestLLMConfig_DefaultConfiguration(t *testing.T) {
	t.Run("Use embedded default when no file specified", func(t *testing.T) {
		llmConfig := &config.LLMConfig{
			// No ProvidersFile specified, should use embedded default
			// Set credentials for default provider (Gemini)
			GeminiProject:  "test-project",
			GeminiLocation: "us-central1",
		}

		// When no path is provided, it should use the embedded default
		providersConfig, err := llmConfig.LoadAndValidate()
		gt.NoError(t, err)
		gt.NotEqual(t, providersConfig, nil)

		// Should have default providers
		gt.NotEqual(t, len(providersConfig.Providers), 0)

		// Check that default providers exist
		_, hasOpenAI := providersConfig.Providers["openai"]
		_, hasClaude := providersConfig.Providers["claude"]
		_, hasGemini := providersConfig.Providers["gemini"]

		gt.Equal(t, hasOpenAI, true)
		gt.Equal(t, hasClaude, true)
		gt.Equal(t, hasGemini, true)
	})
}
