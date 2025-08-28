package slack

import (
	"context"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	slack_model "github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/m-mizutani/tamamo/pkg/utils/async"
	"github.com/slack-go/slack/slackevents"
)

// Controller handles Slack events
type Controller struct {
	event       interfaces.SlackEventUseCases
	slackClient interfaces.SlackClient
	useCases    *usecase.Slack // Add reference to use cases for message logging
}

// New creates a new Slack controller
func New(event interfaces.SlackEventUseCases, slackClient interfaces.SlackClient) *Controller {
	return &Controller{
		event:       event,
		slackClient: slackClient,
	}
}

// NewWithUseCases creates a new Slack controller with use cases for message logging
func NewWithUseCases(event interfaces.SlackEventUseCases, slackClient interfaces.SlackClient, useCases *usecase.Slack) *Controller {
	return &Controller{
		event:       event,
		slackClient: slackClient,
		useCases:    useCases,
	}
}

// enrichMessageWithUserInfo fetches user display name from Slack API and updates the message
func (x *Controller) enrichMessageWithUserInfo(ctx context.Context, slackMsg *slack_model.Message) error {
	logger := ctxlog.From(ctx)

	if x.slackClient == nil {
		return nil // No client available, skip enrichment
	}

	// For user messages, fetch user info
	if slackMsg.UserID != "" {
		userInfo, err := x.slackClient.GetUserInfo(ctx, slackMsg.UserID)
		if err != nil {
			logger.Warn("failed to get user info from Slack API",
				"user_id", slackMsg.UserID,
				"error", err,
			)
			// Don't return error, just use the existing UserName
			return nil
		}

		// Update with display name from Slack
		if userInfo.RealName != "" {
			slackMsg.UserName = userInfo.RealName
		} else if userInfo.DisplayName != "" {
			slackMsg.UserName = userInfo.DisplayName
		} else if userInfo.Name != "" {
			slackMsg.UserName = userInfo.Name
		}

		logger.Debug("enriched user message with display name",
			"user_id", slackMsg.UserID,
			"user_name", slackMsg.UserName,
		)
	}

	// For bot messages, fetch bot info
	if slackMsg.BotID != "" {
		botInfo, err := x.slackClient.GetBotInfo(ctx, slackMsg.BotID)
		if err != nil {
			logger.Warn("failed to get bot info from Slack API",
				"bot_id", slackMsg.BotID,
				"error", err,
			)
			// Don't return error, just use the existing UserName
			return nil
		}

		// Update with bot name from Slack
		if botInfo.Name != "" {
			slackMsg.UserName = botInfo.Name
		}

		logger.Debug("enriched bot message with display name",
			"bot_id", slackMsg.BotID,
			"user_name", slackMsg.UserName,
		)
	}

	return nil
}

// logMessageAsync handles asynchronous message logging for any Slack event
func (x *Controller) logMessageAsync(ctx context.Context, logFunc func(context.Context) error) {
	if x.useCases != nil && x.useCases.GetSlackMessageLogging() != nil {
		async.Dispatch(ctx, logFunc)
	}
}

// HandleSlackAppMention handles Slack app mention events
func (x *Controller) HandleSlackAppMention(ctx context.Context, apiEvent *slackevents.EventsAPIEvent, event *slackevents.AppMentionEvent) error {
	ctxlog.From(ctx).Debug("handling slack app mention",
		"event_ts", event.EventTimeStamp,
		"channel", event.Channel,
	)

	slackMsg := slack_model.NewMessage(ctx, apiEvent)
	if slackMsg == nil {
		return nil
	}

	// Fetch user display name if needed
	if err := x.enrichMessageWithUserInfo(ctx, slackMsg); err != nil {
		ctxlog.From(ctx).Warn("failed to enrich message with user info", "error", err)
		// Continue processing even if user info fetch fails
	}

	// Log message asynchronously (if message logging is available)
	x.logMessageAsync(ctx, func(ctx context.Context) error {
		return x.useCases.GetSlackMessageLogging().LogSlackAppMentionMessage(ctx, event)
	})

	// Process the mention event
	return x.event.HandleSlackAppMention(ctx, *slackMsg)
}

// HandleSlackMessage handles Slack message events
func (x *Controller) HandleSlackMessage(ctx context.Context, apiEvent *slackevents.EventsAPIEvent, event *slackevents.MessageEvent) error {
	ctxlog.From(ctx).Debug("handling slack message",
		"event_ts", event.EventTimeStamp,
		"channel", event.Channel,
	)

	slackMsg := slack_model.NewMessage(ctx, apiEvent)
	if slackMsg == nil {
		return nil
	}

	// Fetch user display name if needed
	if err := x.enrichMessageWithUserInfo(ctx, slackMsg); err != nil {
		ctxlog.From(ctx).Warn("failed to enrich message with user info", "error", err)
		// Continue processing even if user info fetch fails
	}

	// Log message asynchronously (if message logging is available)
	x.logMessageAsync(ctx, func(ctx context.Context) error {
		return x.useCases.GetSlackMessageLogging().LogSlackMessage(ctx, event)
	})

	// Process the message event
	return x.event.HandleSlackMessage(ctx, *slackMsg)
}
