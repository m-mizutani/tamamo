package agent

import (
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

type Agent struct {
	ID          types.UUID `json:"id"`
	AgentID     string     `json:"agent_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Author      string     `json:"author"`
	Latest      string     `json:"latest"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type LLMProvider string

const (
	LLMProviderOpenAI LLMProvider = "OpenAI"
	LLMProviderClaude LLMProvider = "Claude"
	LLMProviderGemini LLMProvider = "Gemini"
)

// String returns the string representation of LLMProvider
func (p LLMProvider) String() string {
	return string(p)
}

// IsValid checks if the LLMProvider is valid
func (p LLMProvider) IsValid() bool {
	switch p {
	case LLMProviderOpenAI, LLMProviderClaude, LLMProviderGemini:
		return true
	default:
		return false
	}
}
