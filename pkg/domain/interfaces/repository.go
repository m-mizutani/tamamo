package interfaces

import (
	"context"

	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// ThreadRepository manages thread, message and history persistence
type ThreadRepository interface {
	// Thread operations
	GetThread(ctx context.Context, id types.ThreadID) (*slack.Thread, error)
	GetThreadByTS(ctx context.Context, channelID, threadTS string) (*slack.Thread, error)
	GetOrPutThread(ctx context.Context, teamID, channelID, threadTS string) (*slack.Thread, error)

	// Message operations
	PutThreadMessage(ctx context.Context, threadID types.ThreadID, message *slack.Message) error
	GetThreadMessages(ctx context.Context, threadID types.ThreadID) ([]*slack.Message, error)

	// History operations
	PutHistory(ctx context.Context, history *slack.History) error
	GetLatestHistory(ctx context.Context, threadID types.ThreadID) (*slack.History, error)
	GetHistoryByID(ctx context.Context, id types.HistoryID) (*slack.History, error)
}

// HistoryRepository is deprecated - use ThreadRepository instead
// Kept for backward compatibility
type HistoryRepository interface {
	// History operations
	PutHistory(ctx context.Context, history *slack.History) error
	GetLatestHistory(ctx context.Context, threadID types.ThreadID) (*slack.History, error)
	GetHistoryByID(ctx context.Context, id types.HistoryID) (*slack.History, error)
}
