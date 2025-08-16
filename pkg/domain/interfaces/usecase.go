package interfaces

import (
	"context"

	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
)

type SlackEventUseCases interface {
	HandleSlackAppMention(ctx context.Context, slackMsg slack.Message) error
	HandleSlackMessage(ctx context.Context, slackMsg slack.Message) error
}
