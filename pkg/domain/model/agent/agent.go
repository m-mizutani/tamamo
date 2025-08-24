package agent

import (
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

type Agent struct {
	ID          types.UUID   `json:"id"`
	AgentID     string       `json:"agent_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Author      types.UserID `json:"author"`
	Status      Status       `json:"status"`
	Latest      string       `json:"latest"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
