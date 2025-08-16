package slack

import (
	"regexp"
	"strings"
)

var mentionPattern = regexp.MustCompile(`<@([A-Z0-9]+)>`)

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
