package interfaces

import (
	"context"

	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// ThreadRepository manages thread and message persistence
type ThreadRepository interface {
	// Thread operations
	GetThread(ctx context.Context, id types.ThreadID) (*slack.Thread, error)
	GetThreadByTS(ctx context.Context, channelID, threadTS string) (*slack.Thread, error)
	GetOrPutThread(ctx context.Context, teamID, channelID, threadTS string) (*slack.Thread, error)

	// Message operations
	PutThreadMessage(ctx context.Context, threadID types.ThreadID, message *slack.Message) error
	GetThreadMessages(ctx context.Context, threadID types.ThreadID) ([]*slack.Message, error)
}
