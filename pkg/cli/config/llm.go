package config

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/llm"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	llmService "github.com/m-mizutani/tamamo/pkg/service/llm"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

//go:embed templates/llm.yaml
var defaultProvidersConfig string

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	ProvidersFile string // YAML file path

	// Claude settings
	ClaudeAPIKey         string
	ClaudeVertexProject  string
	ClaudeVertexLocation string

	// OpenAI settings
	OpenAIAPIKey string

	// Gemini settings (existing)
	GeminiProject  string
	GeminiLocation string

	// Default settings (override)
	DefaultProvider string
	DefaultModel    string
}

// LoadAndValidate reads the providers config file and validates credentials
func (c *LLMConfig) LoadAndValidate() (*llm.ProvidersConfig, error) {
	var config llm.ProvidersConfig

	if c.ProvidersFile != "" {
		// Load from file
		data, err := os.ReadFile(c.ProvidersFile)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to read providers config file", goerr.V("file", c.ProvidersFile))
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, goerr.Wrap(err, "failed to parse providers config", goerr.V("file", c.ProvidersFile))
		}
	} else {
		// Use default config
		if err := yaml.Unmarshal([]byte(defaultProvidersConfig), &config); err != nil {
			return nil, goerr.Wrap(err, "failed to parse default providers config")
		}
	}

	// Override defaults if specified
	if c.DefaultProvider != "" {
		config.Defaults.Provider = c.DefaultProvider
	}
	if c.DefaultModel != "" {
		config.Defaults.Model = c.DefaultModel
	}

	// Validate that we have credentials for configured providers
	if err := c.validateCredentials(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateCredentials checks if required credentials are present for configured providers
func (c *LLMConfig) validateCredentials(config *llm.ProvidersConfig) error {
	requiredProviders := make(map[string]bool)

	// Check default provider
	if config.Defaults.Provider != "" {
		requiredProviders[config.Defaults.Provider] = true
	}

	// Check fallback provider
	if config.Fallback.Enabled && config.Fallback.Provider != "" {
		requiredProviders[config.Fallback.Provider] = true
	}

	// Validate credentials for required providers
	for provider := range requiredProviders {
		switch provider {
		case "gemini":
			if c.GeminiProject == "" {
				return goerr.New("Gemini provider requires project ID", goerr.V("provider", provider))
			}
		case "claude":
			hasDirectAPI := c.ClaudeAPIKey != ""
			hasVertexAI := c.ClaudeVertexProject != "" && c.ClaudeVertexLocation != ""
			if !hasDirectAPI && !hasVertexAI {
				return goerr.New("Claude provider requires either API key or VertexAI project/location", goerr.V("provider", provider))
			}
		case "openai":
			if c.OpenAIAPIKey == "" {
				return goerr.New("OpenAI provider requires API key", goerr.V("provider", provider))
			}
		}
	}

	return nil
}

// BuildFactory creates and configures the LLM Factory with all providers
func (c *LLMConfig) BuildFactory(ctx context.Context, providersConfig *llm.ProvidersConfig) (*llmService.Factory, error) {
	credentials := make(map[types.LLMProvider]llmService.Credential)

	// Set up Gemini credentials
	if c.GeminiProject != "" {
		credentials[types.LLMProviderGemini] = llmService.Credential{
			ProjectID: c.GeminiProject,
			Location:  c.GeminiLocation,
		}
		if c.GeminiLocation == "" {
			credentials[types.LLMProviderGemini] = llmService.Credential{
				ProjectID: c.GeminiProject,
				Location:  "us-central1", // Default location
			}
		}
	}

	// Set up Claude credentials
	if c.ClaudeAPIKey != "" || (c.ClaudeVertexProject != "" && c.ClaudeVertexLocation != "") {
		credentials[types.LLMProviderClaude] = llmService.Credential{
			APIKey:    c.ClaudeAPIKey,
			ProjectID: c.ClaudeVertexProject,
			Location:  c.ClaudeVertexLocation,
		}
	}

	// Set up OpenAI credentials
	if c.OpenAIAPIKey != "" {
		credentials[types.LLMProviderOpenAI] = llmService.Credential{
			APIKey: c.OpenAIAPIKey,
		}
	}

	return llmService.NewFactory(providersConfig, credentials)
}

// Flags returns CLI flags for LLM configuration
func (c *LLMConfig) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "llm-config",
			Sources:     cli.EnvVars("TAMAMO_LLM_CONFIG"),
			Usage:       "Path to LLM providers configuration file",
			Destination: &c.ProvidersFile,
		},
		&cli.StringFlag{
			Name:        "claude-api-key",
			Sources:     cli.EnvVars("TAMAMO_CLAUDE_API_KEY"),
			Usage:       "Claude API key for direct API access",
			Destination: &c.ClaudeAPIKey,
		},
		&cli.StringFlag{
			Name:        "claude-vertex-project",
			Sources:     cli.EnvVars("TAMAMO_CLAUDE_VERTEX_PROJECT"),
			Usage:       "GCP project ID for Claude via VertexAI",
			Destination: &c.ClaudeVertexProject,
		},
		&cli.StringFlag{
			Name:        "claude-vertex-location",
			Sources:     cli.EnvVars("TAMAMO_CLAUDE_VERTEX_LOCATION"),
			Usage:       "GCP location for Claude via VertexAI",
			Destination: &c.ClaudeVertexLocation,
		},
		&cli.StringFlag{
			Name:        "openai-api-key",
			Sources:     cli.EnvVars("TAMAMO_OPENAI_API_KEY"),
			Usage:       "OpenAI API key",
			Destination: &c.OpenAIAPIKey,
		},
		&cli.StringFlag{
			Name:        "gemini-project-id",
			Sources:     cli.EnvVars("TAMAMO_GEMINI_PROJECT_ID"),
			Usage:       "Google Cloud Project ID for Gemini API",
			Destination: &c.GeminiProject,
		},
		&cli.StringFlag{
			Name:        "gemini-location",
			Sources:     cli.EnvVars("TAMAMO_GEMINI_LOCATION"),
			Usage:       "Google Cloud location for Gemini API",
			Value:       "us-central1",
			Destination: &c.GeminiLocation,
		},
		&cli.StringFlag{
			Name:        "llm-default-provider",
			Sources:     cli.EnvVars("TAMAMO_LLM_DEFAULT_PROVIDER"),
			Usage:       "Default LLM provider (overrides config file)",
			Destination: &c.DefaultProvider,
		},
		&cli.StringFlag{
			Name:        "llm-default-model",
			Sources:     cli.EnvVars("TAMAMO_LLM_DEFAULT_MODEL"),
			Usage:       "Default LLM model (overrides config file)",
			Destination: &c.DefaultModel,
		},
	}
}

// GetDefaultProvidersConfig returns the default providers configuration template
func GetDefaultProvidersConfig() string {
	return defaultProvidersConfig
}

// GenerateConfigFile writes the default configuration to a file
func GenerateConfigFile(outputPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0750); err != nil { // #nosec G301 - 0750 is appropriate for config directories
		return goerr.Wrap(err, "failed to create directory", goerr.V("dir", dir))
	}

	// Write default config
	if err := os.WriteFile(outputPath, []byte(defaultProvidersConfig), 0600); err != nil { // #nosec G306 - 0600 is appropriate for config files
		return goerr.Wrap(err, "failed to write config file", goerr.V("path", outputPath))
	}

	return nil
}
