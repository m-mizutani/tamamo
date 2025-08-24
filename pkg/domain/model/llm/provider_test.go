package llm_test

import (
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/llm"
)

func TestProvidersConfig_ValidateProviderModel(t *testing.T) {
	config := &llm.ProvidersConfig{
		Providers: map[string]llm.Provider{
			"openai": {
				ID:          "openai",
				DisplayName: "OpenAI",
				Models: []llm.Model{
					{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
					{ID: "gpt-5-nano-2025-08-07", DisplayName: "GPT-5 Nano"},
				},
			},
			"claude": {
				ID:          "claude",
				DisplayName: "Claude",
				Models: []llm.Model{
					{ID: "claude-sonnet-4-20250514", DisplayName: "Claude Sonnet 4"},
					{ID: "claude-3-7-sonnet-20250219", DisplayName: "Claude 3.7 Sonnet"},
				},
			},
			"gemini": {
				ID:          "gemini",
				DisplayName: "Gemini",
				Models: []llm.Model{
					{ID: "gemini-2.5-flash", DisplayName: "Gemini 2.5 Flash"},
					{ID: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash"},
				},
			},
		},
	}

	t.Run("Valid provider and model", func(t *testing.T) {
		gt.Value(t, config.ValidateProviderModel("openai", "gpt-5-2025-08-07")).Equal(true)
		gt.Value(t, config.ValidateProviderModel("claude", "claude-sonnet-4-20250514")).Equal(true)
		gt.Value(t, config.ValidateProviderModel("gemini", "gemini-2.5-flash")).Equal(true)
	})

	t.Run("Invalid provider", func(t *testing.T) {
		gt.Value(t, config.ValidateProviderModel("invalid", "gpt-5-2025-08-07")).Equal(false)
		gt.Value(t, config.ValidateProviderModel("", "gpt-5-2025-08-07")).Equal(false)
	})

	t.Run("Invalid model for valid provider", func(t *testing.T) {
		gt.Value(t, config.ValidateProviderModel("openai", "invalid-model")).Equal(false)
		gt.Value(t, config.ValidateProviderModel("claude", "gpt-5-2025-08-07")).Equal(false)
		gt.Value(t, config.ValidateProviderModel("gemini", "claude-sonnet-4-20250514")).Equal(false)
	})

	t.Run("Empty model", func(t *testing.T) {
		gt.Value(t, config.ValidateProviderModel("openai", "")).Equal(false)
	})

	t.Run("Case sensitivity", func(t *testing.T) {
		// Provider names are case-sensitive
		gt.Value(t, config.ValidateProviderModel("OpenAI", "gpt-5-2025-08-07")).Equal(false)
		gt.Value(t, config.ValidateProviderModel("CLAUDE", "claude-sonnet-4-20250514")).Equal(false)
	})
}

func TestProvidersConfig_GetProvider(t *testing.T) {
	config := &llm.ProvidersConfig{
		Providers: map[string]llm.Provider{
			"openai": {
				ID:          "openai",
				DisplayName: "OpenAI",
				Models: []llm.Model{
					{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5"},
				},
			},
		},
	}

	t.Run("Existing provider", func(t *testing.T) {
		provider, exists := config.GetProvider("openai")
		gt.Value(t, exists).Equal(true)
		gt.Value(t, provider).NotEqual(nil)
		gt.Value(t, provider.ID).Equal("openai")
		gt.Value(t, provider.DisplayName).Equal("OpenAI")
	})

	t.Run("Non-existing provider", func(t *testing.T) {
		provider, exists := config.GetProvider("claude")
		gt.Value(t, exists).Equal(false)
		gt.Value(t, provider).Equal(nil)
	})
}

func TestProvider_GetModel(t *testing.T) {
	provider := llm.Provider{
		ID:          "openai",
		DisplayName: "OpenAI",
		Models: []llm.Model{
			{ID: "gpt-5-2025-08-07", DisplayName: "GPT-5", Description: "Latest model"},
			{ID: "gpt-5-nano-2025-08-07", DisplayName: "GPT-5 Nano", Description: "Lightweight model"},
		},
	}

	t.Run("Existing model", func(t *testing.T) {
		// Find model by iterating through Models slice
		var foundModel llm.Model
		var found bool
		for _, m := range provider.Models {
			if m.ID == "gpt-5-2025-08-07" {
				foundModel = m
				found = true
				break
			}
		}
		gt.Value(t, found).Equal(true)
		gt.Value(t, foundModel.ID).Equal("gpt-5-2025-08-07")
		gt.Value(t, foundModel.DisplayName).Equal("GPT-5")
		gt.Value(t, foundModel.Description).Equal("Latest model")
	})

	t.Run("Non-existing model", func(t *testing.T) {
		// Find model by iterating through Models slice
		var found bool
		for _, m := range provider.Models {
			if m.ID == "gpt-4" {
				found = true
				break
			}
		}
		gt.Value(t, found).Equal(false)
	})

	t.Run("Empty model ID", func(t *testing.T) {
		// Find model by iterating through Models slice
		var found bool
		for _, m := range provider.Models {
			if m.ID == "" {
				found = true
				break
			}
		}
		gt.Value(t, found).Equal(false)
	})
}

func TestDefaultConfig(t *testing.T) {
	t.Run("Default config with values", func(t *testing.T) {
		config := llm.DefaultConfig{
			Provider: "openai",
			Model:    "gpt-5-2025-08-07",
		}
		gt.Value(t, config.Provider).Equal("openai")
		gt.Value(t, config.Model).Equal("gpt-5-2025-08-07")
	})

	t.Run("Empty default config", func(t *testing.T) {
		config := llm.DefaultConfig{}
		gt.Value(t, config.Provider).Equal("")
		gt.Value(t, config.Model).Equal("")
	})
}

func TestFallbackConfig(t *testing.T) {
	t.Run("Fallback enabled", func(t *testing.T) {
		config := llm.FallbackConfig{
			Enabled:  true,
			Provider: "gemini",
			Model:    "gemini-2.0-flash",
		}
		gt.Value(t, config.Enabled).Equal(true)
		gt.Value(t, config.Provider).Equal("gemini")
		gt.Value(t, config.Model).Equal("gemini-2.0-flash")
	})

	t.Run("Fallback disabled", func(t *testing.T) {
		config := llm.FallbackConfig{
			Enabled: false,
		}
		gt.Value(t, config.Enabled).Equal(false)
		gt.Value(t, config.Provider).Equal("")
		gt.Value(t, config.Model).Equal("")
	})
}