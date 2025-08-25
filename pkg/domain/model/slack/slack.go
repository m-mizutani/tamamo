package slack

import (
	"context"
	"time"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/slack-go/slack/slackevents"
)

// User represents a Slack user from events
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Mention represents a mention in a Slack message
type Mention struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

// Message represents a Slack message (both from events and for persistence)
type Message struct {
	// Core message fields
	ID        types.MessageID `json:"id"`
	Text      string          `json:"text"`
	UserID    string          `json:"user_id"`
	UserName  string          `json:"user_name"`
	BotID     string          `json:"bot_id,omitempty"` // Bot ID for bot messages
	Timestamp string          `json:"timestamp"`
	CreatedAt time.Time       `json:"created_at"`

	// Thread association (for persistence)
	ThreadID types.ThreadID `json:"thread_id,omitempty"` // Only set for persistence

	// Slack event fields
	ThreadTS string    `json:"thread_ts,omitempty"` // From Slack events
	Channel  string    `json:"channel,omitempty"`   // From Slack events
	TeamID   string    `json:"team_id,omitempty"`   // From Slack events
	Mentions []Mention `json:"mentions,omitempty"`  // From Slack events
}

// GetThreadTS returns the thread timestamp for this message
// If the message is not in a thread, returns the message timestamp
func (x *Message) GetThreadTS() string {
	// For Slack events, use ThreadTS if available
	if x.ThreadTS != "" {
		return x.ThreadTS
	}
	// Fallback to message timestamp (for starting new threads)
	return x.Timestamp
}

// InThread returns true if the message is in a thread
func (x *Message) InThread() bool {
	return x.ThreadTS != "" || x.ThreadID != ""
}

// Validate checks if the message has valid fields (for persistence)
func (m *Message) Validate() error {
	if m.ID == "" {
		return ErrEmptyMessageID
	}
	if m.ID != "" && !m.ID.IsValid() {
		return ErrInvalidMessageID
	}
	if m.ThreadID != "" && !m.ThreadID.IsValid() {
		return ErrInvalidThreadID
	}
	// Either UserID or BotID must be present
	if m.UserID == "" && m.BotID == "" {
		return ErrEmptyUserID
	}
	if m.Text == "" {
		return ErrEmptyText
	}
	if m.Timestamp == "" {
		return ErrEmptyTimestamp
	}
	return nil
}

// NewMessage creates a Message from Slack events
func NewMessage(ctx context.Context, ev *slackevents.EventsAPIEvent) *Message {
	// Generate MessageID for the message
	msgID := types.NewMessageID(ctx)

	switch inEv := ev.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		return &Message{
			ID:        msgID,
			ThreadTS:  inEv.ThreadTimeStamp,
			Channel:   inEv.Channel,
			TeamID:    ev.TeamID,
			UserID:    inEv.User,
			UserName:  getUserDisplayName(inEv.User, ""), // App mentions are always from users
			Text:      inEv.Text,
			Timestamp: inEv.TimeStamp,
			Mentions:  ParseMention(inEv.Text),
			CreatedAt: time.Now(),
		}

	case *slackevents.MessageEvent:
		return &Message{
			ID:        msgID,
			ThreadTS:  inEv.ThreadTimeStamp,
			Channel:   inEv.Channel,
			TeamID:    ev.TeamID,
			UserID:    inEv.User,                                 // User ID (empty for bot messages)
			BotID:     inEv.BotID,                                // Bot ID (only for bot messages)
			UserName:  getUserDisplayName(inEv.User, inEv.BotID), // Will implement this
			Text:      inEv.Text,
			Timestamp: inEv.TimeStamp,
			Mentions:  ParseMention(inEv.Text),
			CreatedAt: time.Now(),
		}

	default:
		ctxlog.From(ctx).Warn("unknown event type", "event", inEv)
		return nil
	}
}

// getUserDisplayName returns a display name for the user or bot
// For now, returns the ID as placeholder - TODO: implement proper Slack API integration
func getUserDisplayName(userID, botID string) string {
	if userID != "" {
		// TODO: Fetch user display name from Slack API
		return userID // Placeholder: return user ID for now
	}
	if botID != "" {
		// TODO: Fetch bot display name from Slack API
		return botID // Placeholder: return bot ID for now
	}
	return "unknown"
}
