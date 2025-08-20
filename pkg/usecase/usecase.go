package usecase

import (
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/repository/storage"
)

// Slack holds all use cases
type Slack struct {
	slackClient     interfaces.SlackClient
	repository      interfaces.ThreadRepository
	agentRepository interfaces.AgentRepository
	storageRepo     *storage.Client
	llmClient       gollem.LLMClient
	llmModel        string
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

// WithStorageRepository sets the storage repository
func WithStorageRepository(repo *storage.Client) SlackOption {
	return func(uc *Slack) {
		uc.storageRepo = repo
	}
}

// WithLLMClient sets the LLM client
func WithLLMClient(client gollem.LLMClient) SlackOption {
	return func(uc *Slack) {
		uc.llmClient = client
	}
}

// WithLLMModel sets the LLM model
func WithLLMModel(model string) SlackOption {
	return func(uc *Slack) {
		uc.llmModel = model
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

// New creates a new Slack instance
func New(opts ...SlackOption) *Slack {
	uc := &Slack{}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// Ensure Slack implements required interfaces
var _ interfaces.SlackEventUseCases = (*Slack)(nil)
