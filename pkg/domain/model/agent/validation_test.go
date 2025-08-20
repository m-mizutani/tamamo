package agent_test

import (
	"strings"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

func TestValidateAgent(t *testing.T) {
	now := time.Now()
	validAgent := &agent.Agent{
		ID:          types.UUID("01234567-abcd-1234-abcd-0123456789ab"),
		AgentID:     "test-agent",
		Name:        "Test Agent",
		Description: "A test agent",
		Author:      "Test Author",
		Status:      agent.StatusActive,
		Latest:      "1.0.0",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	testCases := []struct {
		name      string
		agent     *agent.Agent
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "valid agent with active status",
			agent:     validAgent,
			shouldErr: false,
		},
		{
			name: "valid agent with archived status",
			agent: &agent.Agent{
				ID:          types.UUID("01234567-abcd-1234-abcd-0123456789ab"),
				AgentID:     "test-agent",
				Name:        "Test Agent",
				Description: "A test agent",
				Author:      "Test Author",
				Status:      agent.StatusArchived,
				Latest:      "1.0.0",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			shouldErr: false,
		},
		{
			name: "invalid agent with empty status",
			agent: &agent.Agent{
				ID:          types.UUID("01234567-abcd-1234-abcd-0123456789ab"),
				AgentID:     "test-agent",
				Name:        "Test Agent",
				Description: "A test agent",
				Author:      "Test Author",
				Status:      "",
				Latest:      "1.0.0",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			shouldErr: true,
			errMsg:    "invalid agent status",
		},
		{
			name: "invalid agent with unknown status",
			agent: &agent.Agent{
				ID:          types.UUID("01234567-abcd-1234-abcd-0123456789ab"),
				AgentID:     "test-agent",
				Name:        "Test Agent",
				Description: "A test agent",
				Author:      "Test Author",
				Status:      "unknown",
				Latest:      "1.0.0",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			shouldErr: true,
			errMsg:    "invalid agent status",
		},
		{
			name: "invalid agent with empty name",
			agent: &agent.Agent{
				ID:          types.UUID("01234567-abcd-1234-abcd-0123456789ab"),
				AgentID:     "test-agent",
				Name:        "",
				Description: "A test agent",
				Author:      "Test Author",
				Status:      agent.StatusActive,
				Latest:      "1.0.0",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			shouldErr: true,
			errMsg:    "agent name cannot be empty",
		},
		{
			name: "invalid agent with empty agent ID",
			agent: &agent.Agent{
				ID:          types.UUID("01234567-abcd-1234-abcd-0123456789ab"),
				AgentID:     "",
				Name:        "Test Agent",
				Description: "A test agent",
				Author:      "Test Author",
				Status:      agent.StatusActive,
				Latest:      "1.0.0",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			shouldErr: true,
			errMsg:    "invalid agent ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := agent.ValidateAgent(tc.agent)
			if tc.shouldErr {
				gt.Error(t, err)
				if tc.errMsg != "" {
					gt.True(t, strings.Contains(err.Error(), tc.errMsg))
				}
			} else {
				gt.NoError(t, err)
			}
		})
	}
}

func TestValidateAgentID(t *testing.T) {
	testCases := []struct {
		name      string
		agentID   string
		shouldErr bool
	}{
		{
			name:      "valid simple agent ID",
			agentID:   "agent1",
			shouldErr: false,
		},
		{
			name:      "valid agent ID with dash",
			agentID:   "agent-1",
			shouldErr: false,
		},
		{
			name:      "valid agent ID with underscore",
			agentID:   "agent_1",
			shouldErr: false,
		},
		{
			name:      "valid agent ID with dot",
			agentID:   "agent.1",
			shouldErr: false,
		},
		{
			name:      "valid complex agent ID",
			agentID:   "agent-1.2_3",
			shouldErr: false,
		},
		{
			name:      "empty agent ID should be invalid",
			agentID:   "",
			shouldErr: true,
		},
		{
			name:      "agent ID starting with dash should be invalid",
			agentID:   "-agent",
			shouldErr: true,
		},
		{
			name:      "agent ID ending with dash should be invalid",
			agentID:   "agent-",
			shouldErr: true,
		},
		{
			name:      "agent ID with special characters should be invalid",
			agentID:   "agent@test",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := agent.ValidateAgentID(tc.agentID)
			if tc.shouldErr {
				gt.Error(t, err)
			} else {
				gt.NoError(t, err)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	testCases := []struct {
		name      string
		version   string
		shouldErr bool
	}{
		{
			name:      "valid semantic version",
			version:   "1.0.0",
			shouldErr: false,
		},
		{
			name:      "valid complex version",
			version:   "10.20.30",
			shouldErr: false,
		},
		{
			name:      "empty version should be invalid",
			version:   "",
			shouldErr: true,
		},
		{
			name:      "invalid version format",
			version:   "1.0",
			shouldErr: true,
		},
		{
			name:      "invalid version with text",
			version:   "1.0.0-beta",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := agent.ValidateVersion(tc.version)
			if tc.shouldErr {
				gt.Error(t, err)
			} else {
				gt.NoError(t, err)
			}
		})
	}
}
