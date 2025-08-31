package usecase

import (
	"time"

	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/repository/storage"
	"github.com/m-mizutani/tamamo/pkg/service/llm"
	slackservice "github.com/m-mizutani/tamamo/pkg/service/slack"
)

// Slack holds all use cases
type Slack struct {
	slackClient         interfaces.SlackClient
	repository          interfaces.ThreadRepository
	agentRepository     interfaces.AgentRepository
	agentImageRepo      interfaces.AgentImageRepository
	slackMessageLogRepo interfaces.SlackMessageLogRepository
	storageRepo         *storage.Client
	llmClient           gollem.LLMClient // Deprecated: use llmFactory instead
	llmModel            string           // Deprecated: use llmFactory instead
	llmFactory          *llm.Factory
	serverBaseURL       string // Base URL for constructing image URLs
	channelCache        *slackservice.ChannelCache
}

// SlackOption is a functional option for Slack
type SlackOption func(*Slack)

// WithSlackClient sets the Slack client
func WithSlackClient(client interfaces.SlackClient) SlackOption {
	return func(uc *Slack) {
		uc.slackClient = client
	}
}

// WithRepository sets the repository
func WithRepository(repo interfaces.ThreadRepository) SlackOption {
	return func(uc *Slack) {
		uc.repository = repo
	}
}

// WithAgentRepository sets the agent repository
func WithAgentRepository(repo interfaces.AgentRepository) SlackOption {
	return func(uc *Slack) {
		uc.agentRepository = repo
	}
}

// WithAgentImageRepository sets the agent image repository
func WithAgentImageRepository(repo interfaces.AgentImageRepository) SlackOption {
	return func(uc *Slack) {
		uc.agentImageRepo = repo
	}
}

// WithStorageRepository sets the storage repository
func WithStorageRepository(repo *storage.Client) SlackOption {
	return func(uc *Slack) {
		uc.storageRepo = repo
	}
}

// WithLLMClient sets the LLM client (deprecated: use WithLLMFactory)
func WithLLMClient(client gollem.LLMClient) SlackOption {
	return func(uc *Slack) {
		uc.llmClient = client
	}
}

// WithLLMModel sets the LLM model (deprecated: use WithLLMFactory)
func WithLLMModel(model string) SlackOption {
	return func(uc *Slack) {
		uc.llmModel = model
	}
}

// WithLLMFactory sets the LLM factory for multi-provider support
func WithLLMFactory(factory *llm.Factory) SlackOption {
	return func(uc *Slack) {
		uc.llmFactory = factory
		// Also set default client for backward compatibility
		if factory != nil && factory.GetDefaultClient() != nil {
			uc.llmClient = factory.GetDefaultClient()
		}
	}
}

// WithGeminiClient is deprecated. Use WithLLMClient instead.
func WithGeminiClient(client gollem.LLMClient) SlackOption {
	return WithLLMClient(client)
}

// WithGeminiModel is deprecated. Use WithLLMModel instead.
func WithGeminiModel(model string) SlackOption {
	return WithLLMModel(model)
}

// WithSlackMessageLogRepository sets the Slack message log repository
func WithSlackMessageLogRepository(repo interfaces.SlackMessageLogRepository) SlackOption {
	return func(uc *Slack) {
		uc.slackMessageLogRepo = repo
	}
}

// WithServerBaseURL sets the server base URL for constructing image URLs
func WithServerBaseURL(baseURL string) SlackOption {
	return func(uc *Slack) {
		uc.serverBaseURL = baseURL
	}
}

// New creates a new Slack instance
func New(opts ...SlackOption) *Slack {
	uc := &Slack{}
	for _, opt := range opts {
		opt(uc)
	}

	// Initialize channel cache if slack client is available
	if uc.slackClient != nil {
		uc.channelCache = slackservice.NewChannelCache(uc.slackClient, time.Hour)
	}

	return uc
}

// Ensure Slack implements required interfaces
var _ interfaces.SlackEventUseCases = (*Slack)(nil)
