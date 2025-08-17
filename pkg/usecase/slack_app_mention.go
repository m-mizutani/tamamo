package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	pkgErrors "github.com/m-mizutani/tamamo/pkg/utils/errors"
)

// HandleSlackAppMention handles a slack app mention event with LLM integration
func (uc *Slack) HandleSlackAppMention(ctx context.Context, slackMsg slack.Message) error {
	logger := ctxlog.From(ctx)
	logger.Debug("slack app mention event",
		"slack message", slackMsg,
	)

	if uc.slackClient == nil {
		return goerr.New("slack client not configured")
	}

	// Check if LLM is configured
	if uc.geminiClient == nil {
		logger.Warn("LLM not configured, falling back to simple response")
		return uc.handleSimpleResponse(ctx, slackMsg)
	}

	// Find first bot mention
	firstBotMention := uc.findFirstBotMention(slackMsg.Mentions)
	if firstBotMention == nil {
		return nil // No bot mention found
	}

	// Process the bot mention
	return uc.processBotMention(ctx, slackMsg, firstBotMention)
}

// findFirstBotMention finds the first mention that is for the bot
func (uc *Slack) findFirstBotMention(mentions []slack.Mention) *slack.Mention {
	for _, mention := range mentions {
		if uc.slackClient.IsBotUser(mention.UserID) {
			return &mention
		}
	}
	return nil
}

// processBotMention processes a bot mention and generates a response
func (uc *Slack) processBotMention(ctx context.Context, slackMsg slack.Message, mention *slack.Mention) error {
	logger := ctxlog.From(ctx)

	// Store thread and get thread ID
	threadID := uc.storeThreadAndMessage(ctx, &slackMsg)

	// Start chat conversation
	if err := uc.chat(ctx, slackMsg, threadID, mention.Message); err != nil {
		// Log the error with context
		pkgErrors.Handle(ctx, err)

		// Notify user about the error
		errMsg := "I apologize, but I'm experiencing issues processing your request. Please try again later."
		if slackErr := uc.slackClient.PostMessage(ctx, slackMsg.Channel, slackMsg.GetThreadTS(), errMsg); slackErr != nil {
			logger.Error("failed to post error message to slack",
				"error", slackErr,
				"original_error", err,
			)
		}

		return err
	}

	return nil
}

// storeThreadAndMessage stores the thread and message if repository is available
func (uc *Slack) storeThreadAndMessage(ctx context.Context, slackMsg *slack.Message) types.ThreadID {
	logger := ctxlog.From(ctx)

	if uc.repository == nil {
		return ""
	}

	// Get or create thread atomically
	t, err := uc.repository.GetOrPutThread(ctx, slackMsg.TeamID, slackMsg.Channel, slackMsg.GetThreadTS())
	if err != nil {
		logger.Warn("failed to get or create thread",
			"error", err,
			"team_id", slackMsg.TeamID,
			"channel", slackMsg.Channel,
			"thread_ts", slackMsg.GetThreadTS(),
		)
		return ""
	}

	slackMsg.ThreadID = t.ID

	if err := uc.repository.PutThreadMessage(ctx, t.ID, slackMsg); err != nil {
		logger.Warn("failed to save message",
			"error", err,
			"thread_id", t.ID,
			"message_id", slackMsg.ID,
		)
	}

	return t.ID
}

// handleSimpleResponse provides a fallback response when LLM is not configured
func (uc *Slack) handleSimpleResponse(ctx context.Context, slackMsg slack.Message) error {
	logger := ctxlog.From(ctx)

	// Find first bot mention
	firstBotMention := uc.findFirstBotMention(slackMsg.Mentions)
	if firstBotMention == nil {
		return nil // No bot mention found
	}

	// Store thread and message
	uc.storeThreadAndMessage(ctx, &slackMsg)

	// Generate simple response
	responseText := fmt.Sprintf("Hello! You mentioned me with: %s", firstBotMention.Message)
	if firstBotMention.Message == "" {
		responseText = "Hello! How can I help you today?"
	}

	// Reply in thread
	if err := uc.slackClient.PostMessage(ctx, slackMsg.Channel, slackMsg.GetThreadTS(), responseText); err != nil {
		return goerr.Wrap(err, "failed to post message to slack")
	}

	logger.Info("responded to slack mention (simple response)",
		"channel", slackMsg.Channel,
		"thread", slackMsg.GetThreadTS(),
		"user", slackMsg.UserID,
		"message", firstBotMention.Message,
	)
	return nil
}

// chat handles the conversation with LLM and sends responses to Slack
func (uc *Slack) chat(ctx context.Context, slackMsg slack.Message, threadID types.ThreadID, userMessage string) error {
	logger := ctxlog.From(ctx)

	// Load conversation history if thread exists
	var history *gollem.History
	if threadID.IsValid() && uc.repository != nil && uc.storageRepo != nil {
		logger.Debug("attempting to load history for thread",
			"thread_id", threadID,
		)
		// Get the latest history for this thread
		latestHistory, err := uc.repository.GetLatestHistory(ctx, threadID)
		if err != nil {
			if errors.Is(err, slack.ErrHistoryNotFound) {
				// It's normal for new threads to not have history yet
				logger.Debug("no existing history for thread",
					"thread_id", threadID,
				)
			} else {
				// Log other errors as warnings, as this might indicate a problem
				logger.Warn("failed to get latest history, starting new conversation",
					"error", err,
					"thread_id", threadID,
				)
			}
		} else if latestHistory == nil {
			logger.Debug("no history found for thread",
				"thread_id", threadID,
			)
		} else {
			// Load gollem history from storage
			storedHistory, err := uc.storageRepo.LoadHistoryJSON(ctx, threadID, latestHistory.ID)
			if err != nil {
				logger.Warn("failed to load history from storage, but ignore it and start without history",
					"error", err,
					"thread_id", threadID,
					"history_id", latestHistory.ID,
				)
			} else {
				history = &storedHistory
				logger.Debug("loaded conversation history",
					"thread_id", threadID,
					"history_id", latestHistory.ID,
					"message_count", history.ToCount(),
				)
			}
		}
	} else {
		logger.Debug("conditions not met for loading history",
			"thread_id_valid", threadID.IsValid(),
			"has_repository", uc.repository != nil,
			"has_storage_repo", uc.storageRepo != nil,
		)
	}

	// Create session with history if available
	sessionOptions := []gollem.SessionOption{
		gollem.WithSessionSystemPrompt("You are a helpful Slack bot assistant. Respond concisely and helpfully to user questions."),
	}

	if history != nil {
		sessionOptions = append(sessionOptions, gollem.WithSessionHistory(history))
	}

	// Create a new session for this conversation
	session, err := uc.geminiClient.NewSession(ctx, sessionOptions...)
	if err != nil {
		return goerr.Wrap(err, "failed to create LLM session",
			goerr.V("thread_id", threadID),
			goerr.V("channel", slackMsg.Channel),
			goerr.V("user", slackMsg.UserID),
		)
	}

	// Generate content directly through session
	resp, err := session.GenerateContent(ctx, gollem.Text(userMessage))
	if err != nil {
		return goerr.Wrap(err, "failed to generate content with LLM",
			goerr.V("thread_id", threadID),
			goerr.V("message", userMessage),
			goerr.V("channel", slackMsg.Channel),
		)
	}

	var responseText string
	if resp != nil && len(resp.Texts) > 0 {
		responseText = resp.Texts[0]
	}

	// If no response was captured, use a fallback
	if responseText == "" {
		responseText = "(no response)"
	}

	// Send response to Slack
	if err := uc.slackClient.PostMessage(ctx, slackMsg.Channel, slackMsg.GetThreadTS(), responseText); err != nil {
		return goerr.Wrap(err, "failed to post message to slack")
	}

	logger.Info("responded to slack mention with LLM",
		"channel", slackMsg.Channel,
		"thread", slackMsg.GetThreadTS(),
		"user", slackMsg.UserID,
		"message", userMessage,
	)

	// Save updated history to storage for future use
	if threadID.IsValid() && uc.repository != nil && uc.storageRepo != nil && session != nil {
		// Get the session's history
		updatedHistory := session.History()
		if updatedHistory != nil && updatedHistory.ToCount() > 0 {
			// Create history record with consistent ID
			historyRecord := slack.NewHistory(ctx, threadID)
			historyID := historyRecord.ID

			// Save gollem history to storage
			if err := uc.storageRepo.SaveHistoryJSON(ctx, threadID, historyID, updatedHistory); err != nil {
				logger.Warn("failed to save history to storage",
					"error", err,
					"thread_id", threadID,
					"history_id", historyID,
				)
			} else {
				// Save history record to repository
				if err := uc.repository.PutHistory(ctx, historyRecord); err != nil {
					logger.Warn("failed to save history record",
						"error", err,
						"thread_id", threadID,
						"history_id", historyID,
					)
				} else {
					logger.Debug("saved session history",
						"thread_id", threadID,
						"history_id", historyID,
						"message_count", updatedHistory.ToCount(),
						"created_at", historyRecord.CreatedAt,
					)
				}
			}
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
