package graphql

import (
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	threadRepo   interfaces.ThreadRepository
	agentUseCase interfaces.AgentUseCases
}

// NewResolver creates a new resolver instance
func NewResolver(threadRepo interfaces.ThreadRepository, agentUseCase interfaces.AgentUseCases) *Resolver {
	return &Resolver{
		threadRepo:   threadRepo,
		agentUseCase: agentUseCase,
	}
}
