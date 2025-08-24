package llm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/gollem/llm/claude"
	"github.com/m-mizutani/gollem/llm/gemini"
	"github.com/m-mizutani/gollem/llm/openai"
	"github.com/m-mizutani/tamamo/pkg/domain/model/llm"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/utils/logging"
)

// Credential holds authentication information for LLM providers
type Credential struct {
	APIKey    string
	ProjectID string // For Gemini/VertexAI
	Location  string // For Gemini/VertexAI
}

// Factory creates and manages LLM clients
type Factory struct {
	config        *llm.ProvidersConfig
	credentials   map[types.LLMProvider]Credential
	defaultClient gollem.LLMClient
	clients       map[string]gollem.LLMClient // Cache for created clients
}

// NewFactory creates a new LLM factory
func NewFactory(config *llm.ProvidersConfig, credentials map[types.LLMProvider]Credential) (*Factory, error) {
	logger := logging.Default()

	f := &Factory{
		config:      config,
		credentials: credentials,
		clients:     make(map[string]gollem.LLMClient),
	}

	// Log which providers have credentials configured
	for providerType, cred := range credentials {
		hasCredentials := false
		switch providerType {
		case types.LLMProviderOpenAI:
			hasCredentials = cred.APIKey != ""
		case types.LLMProviderClaude:
			hasCredentials = cred.APIKey != ""
		case types.LLMProviderGemini:
			hasCredentials = cred.ProjectID != "" && cred.Location != ""
		}

		if hasCredentials {
			logger.Info("LLM provider credentials configured",
				slog.String("provider", string(providerType)),
				slog.Bool("ready", true),
			)
		} else {
			logger.Warn("LLM provider credentials missing or incomplete",
				slog.String("provider", string(providerType)),
			)
		}
	}

	// Create default client
	if config.Defaults.Provider != "" && config.Defaults.Model != "" {
		logger.Info("Creating default LLM client",
			slog.String("provider", config.Defaults.Provider),
			slog.String("model", config.Defaults.Model),
		)

		ctx := context.Background()
		defaultClient, err := f.CreateClient(ctx, config.Defaults.Provider, config.Defaults.Model)
		if err != nil {
			logger.Error("Failed to create default LLM client",
				slog.String("provider", config.Defaults.Provider),
				slog.String("model", config.Defaults.Model),
				slog.String("error", err.Error()),
			)
			return nil, goerr.Wrap(err, "failed to create default LLM client")
		}
		f.defaultClient = defaultClient

		logger.Info("Default LLM client created successfully",
			slog.String("provider", config.Defaults.Provider),
			slog.String("model", config.Defaults.Model),
		)
	}

	return f, nil
}

// CreateClient creates an LLM client based on provider and model
func (f *Factory) CreateClient(ctx context.Context, provider, model string) (gollem.LLMClient, error) {
	// Validate provider and model
	if !f.config.ValidateProviderModel(provider, model) {
		return nil, goerr.New("invalid provider/model combination", goerr.V("provider", provider), goerr.V("model", model))
	}

	// Check cache
	cacheKey := fmt.Sprintf("%s:%s", provider, model)
	if client, exists := f.clients[cacheKey]; exists {
		return client, nil
	}

	// Get credentials for provider
	providerType := types.LLMProviderFromString(provider)
	cred, exists := f.credentials[providerType]
	if !exists {
		return nil, goerr.New("no credentials configured for provider", goerr.V("provider", provider))
	}

	var client gollem.LLMClient
	var err error

	switch providerType {
	case types.LLMProviderGemini:
		if cred.ProjectID == "" {
			return nil, goerr.New("Gemini requires project ID")
		}
		client, err = gemini.New(ctx, cred.ProjectID, cred.Location, gemini.WithModel(model))
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create Gemini client")
		}

	case types.LLMProviderClaude:
		if cred.APIKey != "" {
			// Use direct Claude API
			client, err = claude.New(ctx, cred.APIKey, claude.WithModel(model))
			if err != nil {
				return nil, goerr.Wrap(err, "failed to create Claude client")
			}
		} else {
			// For now, VertexAI Claude support needs to be implemented in gollem
			// Fallback to error
			return nil, goerr.New("Claude requires API key (VertexAI support not yet available)")
		}

	case types.LLMProviderOpenAI:
		if cred.APIKey == "" {
			return nil, goerr.New("OpenAI requires API key")
		}
		client, err = openai.New(ctx, cred.APIKey, openai.WithModel(model))
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create OpenAI client")
		}

	default:
		return nil, goerr.New("unsupported provider", goerr.V("provider", provider))
	}

	// Cache the client
	f.clients[cacheKey] = client

	// Log successful client creation
	logging.Default().Info("LLM client created",
		slog.String("provider", provider),
		slog.String("model", model),
		slog.String("cache_key", cacheKey),
	)

	return client, nil
}

// GetDefaultClient returns the default LLM client
func (f *Factory) GetDefaultClient() gollem.LLMClient {
	return f.defaultClient
}

// GetFallbackClient returns the fallback LLM client if enabled
func (f *Factory) GetFallbackClient(ctx context.Context) (gollem.LLMClient, error) {
	if !f.config.Fallback.Enabled {
		return nil, goerr.New("fallback is not enabled")
	}

	if f.config.Fallback.Provider == "" || f.config.Fallback.Model == "" {
		return nil, goerr.New("fallback provider/model not configured")
	}

	return f.CreateClient(ctx, f.config.Fallback.Provider, f.config.Fallback.Model)
}

// GetConfig returns the providers configuration
func (f *Factory) GetConfig() *llm.ProvidersConfig {
	return f.config
}
