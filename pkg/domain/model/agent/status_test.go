package agent_test

import (
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
)

func TestStatus_IsValid(t *testing.T) {
	testCases := []struct {
		name     string
		status   agent.Status
		expected bool
	}{
		{
			name:     "active status should be valid",
			status:   agent.StatusActive,
			expected: true,
		},
		{
			name:     "archived status should be valid",
			status:   agent.StatusArchived,
			expected: true,
		},
		{
			name:     "empty status should be invalid",
			status:   "",
			expected: false,
		},
		{
			name:     "unknown status should be invalid",
			status:   "unknown",
			expected: false,
		},
		{
			name:     "invalid case should be invalid",
			status:   "ACTIVE",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.status.IsValid()
			gt.Equal(t, result, tc.expected)
		})
	}
}

func TestStatus_String(t *testing.T) {
	testCases := []struct {
		name     string
		status   agent.Status
		expected string
	}{
		{
			name:     "active status string representation",
			status:   agent.StatusActive,
			expected: "active",
		},
		{
			name:     "archived status string representation",
			status:   agent.StatusArchived,
			expected: "archived",
		},
		{
			name:     "custom status string representation",
			status:   "custom",
			expected: "custom",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.status.String()
			gt.Equal(t, result, tc.expected)
		})
	}
}

func TestValidateStatus(t *testing.T) {
	testCases := []struct {
		name      string
		status    agent.Status
		shouldErr bool
	}{
		{
			name:      "valid active status",
			status:    agent.StatusActive,
			shouldErr: false,
		},
		{
			name:      "valid archived status",
			status:    agent.StatusArchived,
			shouldErr: false,
		},
		{
			name:      "invalid empty status",
			status:    "",
			shouldErr: true,
		},
		{
			name:      "invalid unknown status",
			status:    "unknown",
			shouldErr: true,
		},
		{
			name:      "invalid case status",
			status:    "ACTIVE",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := agent.ValidateStatus(tc.status)
			if tc.shouldErr {
				gt.Error(t, err)
			} else {
				gt.NoError(t, err)
			}
		})
	}
}

func TestStatusConstants(t *testing.T) {
	// Ensure constants have expected values
	gt.Equal(t, string(agent.StatusActive), "active")
	gt.Equal(t, string(agent.StatusArchived), "archived")
}
