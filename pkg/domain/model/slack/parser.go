package slack

import (
	"regexp"
	"strings"
)

var mentionPattern = regexp.MustCompile(`<@([A-Z0-9]+)>`)

// AgentMention represents an agent mention with agent ID
type AgentMention struct {
	UserID  string
	AgentID string // Agent ID (empty for general mode)
	Message string
}

// ParseMention extracts mention information from Slack message text
func ParseMention(text string) []Mention {
	matches := mentionPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}

	mentions := make([]Mention, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		userID := match[1]

		// Extract message after mention
		message := text
		mentionStr := match[0]
		idx := strings.Index(message, mentionStr)
		if idx >= 0 {
			message = strings.TrimSpace(message[idx+len(mentionStr):])
		}

		mentions = append(mentions, Mention{
			UserID:  userID,
			Message: message,
		})
	}

	return mentions
}

// ParseAgentMention extracts agent mention information from Slack message text
func ParseAgentMention(text string) []AgentMention {
	matches := mentionPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}

	mentions := make([]AgentMention, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		userID := match[1]

		// Extract message after mention
		message := text
		mentionStr := match[0]
		idx := strings.Index(message, mentionStr)
		if idx >= 0 {
			message = strings.TrimSpace(message[idx+len(mentionStr):])
		}

		// Parse agent ID from the beginning of the message
		agentID := ""
		if message != "" {
			parts := strings.Fields(message)
			if len(parts) > 0 {
				// Check if first word looks like an agent ID
				// Agent IDs should contain dashes or be specifically formatted alphanumeric
				firstWord := parts[0]
				if isValidAgentID(firstWord) {
					agentID = firstWord
					// Remove agent ID from message
					if len(parts) > 1 {
						message = strings.TrimSpace(strings.Join(parts[1:], " "))
					} else {
						message = ""
					}
				}
			}
		}

		mentions = append(mentions, AgentMention{
			UserID:  userID,
			AgentID: agentID,
			Message: message,
		})
	}

	return mentions
}

// isValidAgentID checks if a string is a valid agent ID
// Valid agent IDs are:
// - 2+ characters long AND contain only alphanumeric characters and dashes
// - OR contain at least one dash (any length)
func isValidAgentID(s string) bool {
	if s == "" {
		return false
	}

	// Must contain only alphanumeric characters and dashes
	if !isAlphanumeric(s) {
		return false
	}

	// Valid if either:
	// 1. Contains at least one dash (kebab-case style: code-helper, a-b, etc.)
	// 2. Is 2+ characters long (allows simple names like: aoko, gpt, claude, etc.)
	containsDash := strings.Contains(s, "-")
	isLongEnough := len(s) >= 2

	return containsDash || isLongEnough
}

// isAlphanumeric checks if a string contains only alphanumeric characters and dashes
func isAlphanumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return true
}

// isNumeric checks if a string contains only numeric characters
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}
