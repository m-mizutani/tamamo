package usecase

import (
	"context"
	"fmt"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
)

// HandleSlackAppMention handles a slack app mention event
func (uc *Slack) HandleSlackAppMention(ctx context.Context, slackMsg slack.Message) error {
	ctxlog.From(ctx).Debug("slack app mention event",
		"slack message", slackMsg,
	)

	if uc.slackClient == nil {
		return goerr.New("slack client not configured")
	}

	// Process each mention
	for i, mention := range slackMsg.Mentions {
		if !uc.slackClient.IsBotUser(mention.UserID) {
			continue
		}

		// First mention to the bot
		if i == 0 {
			// Store thread and message if repository is available
			if uc.repository != nil {
				// Get or create thread atomically
				t, err := uc.repository.GetOrPutThread(ctx, slackMsg.TeamID, slackMsg.Channel, slackMsg.GetThreadTS())
				if err != nil {
					ctxlog.From(ctx).Warn("failed to get or create thread",
						"error", err,
						"team_id", slackMsg.TeamID,
						"channel", slackMsg.Channel,
						"thread_ts", slackMsg.GetThreadTS(),
					)
				} else {
					// Save message - populate ThreadID for persistence
					slackMsg.ThreadID = t.ID

					if err := uc.repository.PutThreadMessage(ctx, t.ID, &slackMsg); err != nil {
						ctxlog.From(ctx).Warn("failed to save message",
							"error", err,
							"thread_id", t.ID,
							"message_id", slackMsg.ID,
						)
					}
				}
			}

			// Simple response for now
			responseText := fmt.Sprintf("Hello! You mentioned me with: %s", mention.Message)
			if mention.Message == "" {
				responseText = "Hello! How can I help you today?"
			}

			// Reply in thread
			if err := uc.slackClient.PostMessage(ctx, slackMsg.Channel, slackMsg.GetThreadTS(), responseText); err != nil {
				return goerr.Wrap(err, "failed to post message to slack")
			}

			ctxlog.From(ctx).Info("responded to slack mention",
				"channel", slackMsg.Channel,
				"thread", slackMsg.GetThreadTS(),
				"user", slackMsg.UserID,
				"message", mention.Message,
			)
		}
	}

	return nil
}

// HandleSlackMessage handles a slack message event
func (uc *Slack) HandleSlackMessage(ctx context.Context, slackMsg slack.Message) error {
	ctxlog.From(ctx).Debug("slack message event",
		"channel", slackMsg.Channel,
		"thread", slackMsg.GetThreadTS(),
		"text", slackMsg.Text,
	)

	// If repository is available, check if this is in a participating thread
	if uc.repository != nil && slackMsg.ThreadTS != "" {
		// Check if we have this thread in our database (meaning we're participating)
		thread, err := uc.repository.GetThreadByTS(ctx, slackMsg.Channel, slackMsg.ThreadTS)
		if err == nil {
			// This is a participating thread, record the message
			slackMsg.ThreadID = thread.ID

			if err := uc.repository.PutThreadMessage(ctx, thread.ID, &slackMsg); err != nil {
				ctxlog.From(ctx).Warn("failed to save message in participating thread",
					"error", err,
					"thread_id", thread.ID,
					"thread_ts", slackMsg.ThreadTS,
					"message_id", slackMsg.ID,
				)
			} else {
				ctxlog.From(ctx).Debug("recorded message in participating thread",
					"thread_id", thread.ID,
					"thread_ts", slackMsg.ThreadTS,
					"user_id", slackMsg.UserID,
				)
			}
		}
		// If thread not found, just ignore (not a participating thread)
	}

	return nil
}
