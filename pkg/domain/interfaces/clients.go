package interfaces

import (
	"context"
)

type SlackClient interface {
	PostMessage(ctx context.Context, channelID, threadTS, text string) error
	IsBotUser(userID string) bool
}
