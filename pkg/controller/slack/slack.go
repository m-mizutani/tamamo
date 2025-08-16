package slack

import (
	"context"

	"github.com/m-mizutani/ctxlog"
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
	ctxlog.From(ctx).Debug("handling slack app mention",
		"event_ts", event.EventTimeStamp,
		"channel", event.Channel,
	)

	slackMsg := slack_model.NewMessage(ctx, apiEvent)
	if slackMsg == nil {
		return nil
	}

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

	// Process the message event
	return x.event.HandleSlackMessage(ctx, *slackMsg)
}
