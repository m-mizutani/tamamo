package usecase

import (
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

// Slack holds all use cases
type Slack struct {
	slackClient interfaces.SlackClient
	repository  interfaces.ThreadRepository
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
