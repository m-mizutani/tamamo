package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
)

// HandleSlackAppMention handles a slack app mention event
func (uc *Slack) HandleSlackAppMention(ctx context.Context, slackMsg slack.Message) error {
	logger := slog.With("channel", slackMsg.ChannelID(), "thread", slackMsg.ThreadID())
	logger.DebugContext(ctx, "slack app mention event", "mentions", slackMsg.Mention())

	if uc.slackClient == nil {
		return goerr.New("slack client not configured")
	}

	// Process each mention
	for i, mention := range slackMsg.Mention() {
		if !uc.slackClient.IsBotUser(mention.UserID) {
			continue
		}

		// First mention to the bot
		if i == 0 {
			// Get the thread to reply to
			thread := slackMsg.Thread()

			// Simple response for now
			responseText := fmt.Sprintf("Hello! You mentioned me with: %s", mention.Message)
			if mention.Message == "" {
				responseText = "Hello! How can I help you today?"
			}

			// Reply in thread
			if err := uc.slackClient.PostMessage(ctx, thread.ChannelID, thread.ThreadID, responseText); err != nil {
				return goerr.Wrap(err, "failed to post message to slack")
			}

			logger.InfoContext(ctx, "responded to slack mention",
				"user", slackMsg.User(),
				"message", mention.Message,
			)
		}
	}

	return nil
}

// HandleSlackMessage handles a slack message event
func (uc *Slack) HandleSlackMessage(ctx context.Context, slackMsg slack.Message) error {
	// For now, we don't process regular messages, only mentions
	logger := slog.With("channel", slackMsg.ChannelID(), "thread", slackMsg.ThreadID())
	logger.DebugContext(ctx, "slack message event (ignored)", "text", slackMsg.Text())
	return nil
}
