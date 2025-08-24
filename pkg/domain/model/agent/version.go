package agent

import (
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

type AgentVersion struct {
	AgentUUID    types.UUID        `json:"agent_uuid"`
	Version      string            `json:"version"`
	SystemPrompt string            `json:"system_prompt"`
	LLMProvider  types.LLMProvider `json:"llm_provider"`
	LLMModel     string            `json:"llm_model"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}
