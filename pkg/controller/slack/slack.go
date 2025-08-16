package slack

import (
	"context"
	"log/slog"

	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	slack_model "github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/slack-go/slack/slackevents"
)

// Controller handles Slack events
type Controller struct {
	event interfaces.SlackEventUseCases
}

// New creates a new Slack controller
func New(event interfaces.SlackEventUseCases) *Controller {
	return &Controller{
		event: event,
	}
}

// HandleSlackAppMention handles Slack app mention events
func (x *Controller) HandleSlackAppMention(ctx context.Context, apiEvent *slackevents.EventsAPIEvent, event *slackevents.AppMentionEvent) error {
	logger := slog.With("event_ts", event.EventTimeStamp, "channel", event.Channel)
	logger.DebugContext(ctx, "handling slack app mention")

	slackMsg := slack_model.NewMessage(ctx, apiEvent)
	if slackMsg == nil {
		return nil
	}

	// Process the mention event
	return x.event.HandleSlackAppMention(ctx, *slackMsg)
}

// HandleSlackMessage handles Slack message events
func (x *Controller) HandleSlackMessage(ctx context.Context, apiEvent *slackevents.EventsAPIEvent, event *slackevents.MessageEvent) error {
	logger := slog.With("event_ts", event.EventTimeStamp, "channel", event.Channel)
	logger.DebugContext(ctx, "handling slack message")

	slackMsg := slack_model.NewMessage(ctx, apiEvent)
	if slackMsg == nil {
		return nil
	}

	// Process the message event
	return x.event.HandleSlackMessage(ctx, *slackMsg)
}
