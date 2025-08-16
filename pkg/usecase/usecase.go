package usecase

import (
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

// Slack holds all use cases
type Slack struct {
	slackClient interfaces.SlackClient
}

// SlackOption is a functional option for Slack
type SlackOption func(*Slack)

// WithSlackClient sets the Slack client
func WithSlackClient(client interfaces.SlackClient) SlackOption {
	return func(uc *Slack) {
		uc.slackClient = client
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
