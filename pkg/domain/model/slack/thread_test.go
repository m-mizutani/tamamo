package slack_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

func TestNewThread(t *testing.T) {
	ctx := context.Background()
	teamID := "T123456"
	channelID := "C123456"
	threadTS := "1234567890.123456"

	thread := slack.NewThread(ctx, teamID, channelID, threadTS)

	gt.V(t, thread).NotEqual(nil)
	gt.V(t, thread.ID.IsValid()).Equal(true)
	gt.V(t, thread.TeamID).Equal(teamID)
	gt.V(t, thread.ChannelID).Equal(channelID)
	gt.V(t, thread.ThreadTS).Equal(threadTS)
	gt.V(t, thread.AgentUUID).Equal(nil) // Should be nil for regular threads
	gt.V(t, thread.AgentVersion).Equal("")
	gt.V(t, thread.CreatedAt).NotEqual(time.Time{})
	gt.V(t, thread.UpdatedAt).Equal(thread.CreatedAt)
}

func TestNewThreadWithAgent(t *testing.T) {
	ctx := context.Background()
	teamID := "T123456"
	channelID := "C123456"
	threadTS := "1234567890.123456"
	agentUUID := types.NewUUID(ctx)
	agentVersion := "v1.0.0"

	thread := slack.NewThreadWithAgent(ctx, teamID, channelID, threadTS, &agentUUID, agentVersion)

	gt.V(t, thread).NotEqual(nil)
	gt.V(t, thread.ID.IsValid()).Equal(true)
	gt.V(t, thread.TeamID).Equal(teamID)
	gt.V(t, thread.ChannelID).Equal(channelID)
	gt.V(t, thread.ThreadTS).Equal(threadTS)
	gt.V(t, thread.AgentUUID).NotEqual(nil)
	gt.V(t, *thread.AgentUUID).Equal(agentUUID)
	gt.V(t, thread.AgentVersion).Equal(agentVersion)
	gt.V(t, thread.CreatedAt).NotEqual(time.Time{})
	gt.V(t, thread.UpdatedAt).Equal(thread.CreatedAt)
}

func TestNewThreadWithAgent_GeneralMode(t *testing.T) {
	ctx := context.Background()
	teamID := "T123456"
	channelID := "C123456"
	threadTS := "1234567890.123456"
	generalModeUUID := types.UUID("00000000-0000-0000-0000-000000000000")
	agentVersion := "general-v1"

	thread := slack.NewThreadWithAgent(ctx, teamID, channelID, threadTS, &generalModeUUID, agentVersion)

	gt.V(t, thread).NotEqual(nil)
	gt.V(t, thread.AgentUUID).NotEqual(nil)
	gt.V(t, *thread.AgentUUID).Equal(generalModeUUID)
	gt.V(t, thread.AgentVersion).Equal(agentVersion)
}

func TestNewThreadWithAgent_NilUUID(t *testing.T) {
	ctx := context.Background()
	teamID := "T123456"
	channelID := "C123456"
	threadTS := "1234567890.123456"
	agentVersion := "v1.0.0"

	thread := slack.NewThreadWithAgent(ctx, teamID, channelID, threadTS, nil, agentVersion)

	gt.V(t, thread).NotEqual(nil)
	gt.V(t, thread.AgentUUID).Equal(nil)
	gt.V(t, thread.AgentVersion).Equal(agentVersion)
}

func TestThreadValidate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		thread    func() *slack.Thread
		expectErr bool
		errorType error
	}{
		{
			name: "valid thread without agent",
			thread: func() *slack.Thread {
				return slack.NewThread(ctx, "T123456", "C123456", "1234567890.123456")
			},
			expectErr: false,
		},
		{
			name: "valid thread with agent",
			thread: func() *slack.Thread {
				agentUUID := types.NewUUID(ctx)
				return slack.NewThreadWithAgent(ctx, "T123456", "C123456", "1234567890.123456", &agentUUID, "v1.0.0")
			},
			expectErr: false,
		},
		{
			name: "valid thread with general mode UUID",
			thread: func() *slack.Thread {
				generalModeUUID := types.UUID("00000000-0000-0000-0000-000000000000")
				return slack.NewThreadWithAgent(ctx, "T123456", "C123456", "1234567890.123456", &generalModeUUID, "general-v1")
			},
			expectErr: false,
		},
		{
			name: "valid thread with nil agent UUID",
			thread: func() *slack.Thread {
				return slack.NewThreadWithAgent(ctx, "T123456", "C123456", "1234567890.123456", nil, "v1.0.0")
			},
			expectErr: false,
		},
		{
			name: "valid thread with empty threadTS",
			thread: func() *slack.Thread {
				return slack.NewThread(ctx, "T123456", "C123456", "")
			},
			expectErr: false,
		},
		{
			name: "invalid thread - empty team ID",
			thread: func() *slack.Thread {
				thread := slack.NewThread(ctx, "", "C123456", "1234567890.123456")
				return thread
			},
			expectErr: true,
			errorType: slack.ErrEmptyTeamID,
		},
		{
			name: "invalid thread - empty channel ID",
			thread: func() *slack.Thread {
				thread := slack.NewThread(ctx, "T123456", "", "1234567890.123456")
				return thread
			},
			expectErr: true,
			errorType: slack.ErrEmptyChannelID,
		},
		{
			name: "invalid thread - invalid thread ID",
			thread: func() *slack.Thread {
				thread := slack.NewThread(ctx, "T123456", "C123456", "1234567890.123456")
				thread.ID = types.ThreadID("") // Invalid ID
				return thread
			},
			expectErr: true,
			errorType: slack.ErrInvalidThreadID,
		},
		{
			name: "invalid thread - invalid agent UUID",
			thread: func() *slack.Thread {
				invalidUUID := types.UUID("invalid-uuid")
				thread := slack.NewThreadWithAgent(ctx, "T123456", "C123456", "1234567890.123456", &invalidUUID, "v1.0.0")
				return thread
			},
			expectErr: true,
			errorType: slack.ErrInvalidAgentUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thread := tt.thread()
			err := thread.Validate()

			if tt.expectErr {
				gt.V(t, err).NotEqual(nil)
				if tt.errorType != nil {
					gt.V(t, errors.Is(err, tt.errorType) || strings.Contains(err.Error(), tt.errorType.Error())).Equal(true)
				}
			} else {
				gt.V(t, err).Equal(nil)
			}
		})
	}
}

func TestThreadTimestamps(t *testing.T) {
	ctx := context.Background()

	// Test that CreatedAt and UpdatedAt are set correctly
	before := time.Now()
	thread := slack.NewThread(ctx, "T123456", "C123456", "1234567890.123456")
	after := time.Now()

	gt.V(t, thread.CreatedAt.After(before) || thread.CreatedAt.Equal(before)).Equal(true)
	gt.V(t, thread.CreatedAt.Before(after) || thread.CreatedAt.Equal(after)).Equal(true)
	gt.V(t, thread.UpdatedAt).Equal(thread.CreatedAt)
}

func TestThreadWithAgentTimestamps(t *testing.T) {
	ctx := context.Background()
	agentUUID := types.NewUUID(ctx)

	// Test that CreatedAt and UpdatedAt are set correctly for threads with agents
	before := time.Now()
	thread := slack.NewThreadWithAgent(ctx, "T123456", "C123456", "1234567890.123456", &agentUUID, "v1.0.0")
	after := time.Now()

	gt.V(t, thread.CreatedAt.After(before) || thread.CreatedAt.Equal(before)).Equal(true)
	gt.V(t, thread.CreatedAt.Before(after) || thread.CreatedAt.Equal(after)).Equal(true)
	gt.V(t, thread.UpdatedAt).Equal(thread.CreatedAt)
}

func TestThreadIDUniqueness(t *testing.T) {
	ctx := context.Background()

	// Create multiple threads and ensure they have unique IDs
	threads := make([]*slack.Thread, 10)
	for i := 0; i < 10; i++ {
		threads[i] = slack.NewThread(ctx, "T123456", "C123456", "1234567890.123456")
	}

	// Check that all IDs are unique
	seen := make(map[types.ThreadID]bool)
	for _, thread := range threads {
		gt.V(t, seen[thread.ID]).Equal(false) // Should not have seen this ID before
		seen[thread.ID] = true
		gt.V(t, thread.ID.IsValid()).Equal(true)
	}
}

func TestThreadCompatibility(t *testing.T) {
	ctx := context.Background()

	t.Run("thread without agent can be validated", func(t *testing.T) {
		thread := slack.NewThread(ctx, "T123456", "C123456", "1234567890.123456")
		err := thread.Validate()
		gt.V(t, err).Equal(nil)

		// Ensure agent fields are properly initialized
		gt.V(t, thread.AgentUUID).Equal(nil)
		gt.V(t, thread.AgentVersion).Equal("")
	})

	t.Run("thread with agent can be validated", func(t *testing.T) {
		agentUUID := types.NewUUID(ctx)
		thread := slack.NewThreadWithAgent(ctx, "T123456", "C123456", "1234567890.123456", &agentUUID, "v1.0.0")
		err := thread.Validate()
		gt.V(t, err).Equal(nil)

		// Ensure agent fields are properly set
		gt.V(t, thread.AgentUUID).NotEqual(nil)
		gt.V(t, *thread.AgentUUID).Equal(agentUUID)
		gt.V(t, thread.AgentVersion).Equal("v1.0.0")
	})
}
