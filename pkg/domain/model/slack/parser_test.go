package slack_test

import (
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
)

func TestParseAgentMention(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []slack.AgentMention
	}{
		{
			name: "agent with message",
			text: "<@U123456> code-helper please help me debug this",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "code-helper",
					Message: "please help me debug this",
				},
			},
		},
		{
			name: "agent with dashes in ID",
			text: "<@U123456> data-analysis-pro analyze this dataset",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "data-analysis-pro",
					Message: "analyze this dataset",
				},
			},
		},
		{
			name: "agent ID only (no message)",
			text: "<@U123456> code-helper",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "code-helper",
					Message: "",
				},
			},
		},
		{
			name: "2+ character word (valid agent ID)",
			text: "<@U123456> hello how are you?",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "hello",
					Message: "how are you?",
				},
			},
		},
		{
			name: "mention only (no text)",
			text: "<@U123456>",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "",
					Message: "",
				},
			},
		},
		{
			name: "mention with whitespace",
			text: "<@U123456>   code-helper   analyze this code   ",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "code-helper",
					Message: "analyze this code",
				},
			},
		},
		{
			name: "alphanumeric without dash (2+ chars, valid agent ID)",
			text: "<@U123456> agent123 help me",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "agent123",
					Message: "help me",
				},
			},
		},
		{
			name: "single character with dash (valid agent ID)",
			text: "<@U123456> a-b run command",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "a-b",
					Message: "run command",
				},
			},
		},
		{
			name: "mixed alphanumeric with dashes",
			text: "<@U123456> agent-v2-pro do something",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "agent-v2-pro",
					Message: "do something",
				},
			},
		},
		{
			name: "multiple mentions",
			text: "<@U123456> code-helper debug this <@U789012> data-helper analyze",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "code-helper",
					Message: "debug this <@U789012> data-helper analyze",
				},
				{
					UserID:  "U789012",
					AgentID: "data-helper",
					Message: "analyze",
				},
			},
		},
		{
			name: "invalid agent ID (contains special characters)",
			text: "<@U123456> code@helper invalid agent",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "",
					Message: "code@helper invalid agent",
				},
			},
		},
		{
			name: "numeric string that looks like agent ID",
			text: "<@U123456> 123456 process this",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "123456",
					Message: "process this",
				},
			},
		},
		{
			name: "single character (not agent ID - no dash)",
			text: "<@U123456> a run command",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "",
					Message: "a run command",
				},
			},
		},
		{
			name: "general conversation (no agent ID pattern)",
			text: "<@U123456> !help with this code",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "",
					Message: "!help with this code",
				},
			},
		},
		{
			name:     "empty text",
			text:     "",
			expected: nil,
		},
		{
			name:     "no mentions",
			text:     "hello world",
			expected: nil,
		},
		{
			name: "message starting with punctuation (not agent ID)",
			text: "<@U123456> !help me",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "",
					Message: "!help me",
				},
			},
		},
		{
			name: "message starting with space then agent ID",
			text: "<@U123456>  code-helper debug",
			expected: []slack.AgentMention{
				{
					UserID:  "U123456",
					AgentID: "code-helper",
					Message: "debug",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slack.ParseAgentMention(tt.text)

			if tt.expected == nil {
				gt.V(t, result).Equal(nil)
				return
			}

			gt.V(t, len(result)).Equal(len(tt.expected))

			for i, expected := range tt.expected {
				gt.V(t, result[i].UserID).Equal(expected.UserID)
				gt.V(t, result[i].AgentID).Equal(expected.AgentID)
				gt.V(t, result[i].Message).Equal(expected.Message)
			}
		})
	}
}

func TestIsValidAgentID(t *testing.T) {
	// Note: isValidAgentID is not exported, so we test it indirectly through ParseAgentMention
	tests := []struct {
		name            string
		text            string
		shouldBeAgentID bool
	}{
		{
			name:            "alphanumeric without dash (2+ chars, valid)",
			text:            "<@U123456> abc123 message",
			shouldBeAgentID: true,
		},
		{
			name:            "valid with dashes",
			text:            "<@U123456> code-helper-v2 message",
			shouldBeAgentID: true,
		},
		{
			name:            "purely numeric (valid)",
			text:            "<@U123456> 123456 message",
			shouldBeAgentID: true,
		},
		{
			name:            "invalid with special chars",
			text:            "<@U123456> code@helper message",
			shouldBeAgentID: false,
		},
		{
			name:            "invalid with spaces (first word valid)",
			text:            "<@U123456> code helper message",
			shouldBeAgentID: true, // "code" is 2+ chars and valid, spaces are in message part
		},
		{
			name:            "invalid with dots",
			text:            "<@U123456> code.helper message",
			shouldBeAgentID: false,
		},
		{
			name:            "empty string",
			text:            "<@U123456> ",
			shouldBeAgentID: false,
		},
		{
			name:            "single character (invalid)",
			text:            "<@U123456> a message",
			shouldBeAgentID: false,
		},
		{
			name:            "single character with dash (valid)",
			text:            "<@U123456> a-1 message",
			shouldBeAgentID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slack.ParseAgentMention(tt.text)
			gt.V(t, len(result)).Equal(1)

			if tt.shouldBeAgentID {
				gt.V(t, result[0].AgentID).NotEqual("")
			} else {
				gt.V(t, result[0].AgentID).Equal("")
			}
		})
	}
}
