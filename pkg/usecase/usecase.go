package usecase

import (
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/repository/storage"
)

// Slack holds all use cases
type Slack struct {
	slackClient  interfaces.SlackClient
	repository   interfaces.ThreadRepository
	storageRepo  *storage.Client
	geminiClient gollem.LLMClient
	geminiModel  string
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

// WithStorageRepository sets the storage repository
func WithStorageRepository(repo *storage.Client) SlackOption {
	return func(uc *Slack) {
		uc.storageRepo = repo
	}
}

// WithGeminiClient sets the Gemini client
func WithGeminiClient(client gollem.LLMClient) SlackOption {
	return func(uc *Slack) {
		uc.geminiClient = client
	}
}

// WithGeminiModel sets the Gemini model
func WithGeminiModel(model string) SlackOption {
	return func(uc *Slack) {
		uc.geminiModel = model
	}
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
