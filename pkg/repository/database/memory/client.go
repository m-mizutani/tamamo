package memory

import (
	"context"
	"sync"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// Client is an in-memory implementation of ThreadRepository and HistoryRepository
type Client struct {
	mu        sync.RWMutex
	threads   map[types.ThreadID]*slack.Thread
	messages  map[types.ThreadID][]*slack.Message
	histories map[types.HistoryID]*slack.History
}

// New creates a new in-memory client
func New() *Client {
	return &Client{
		threads:   make(map[types.ThreadID]*slack.Thread),
		messages:  make(map[types.ThreadID][]*slack.Message),
		histories: make(map[types.HistoryID]*slack.History),
	}
}

// GetOrPutThread gets an existing thread or creates a new one atomically
func (c *Client) GetOrPutThread(ctx context.Context, teamID, channelID, threadTS string) (*slack.Thread, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// First try to find existing thread
	for _, t := range c.threads {
		if t.ChannelID == channelID && t.ThreadTS == threadTS {
			// Return a copy to avoid external modifications
			threadCopy := *t
			return &threadCopy, nil
		}
	}

	// Thread not found, create new one
	t := slack.NewThread(ctx, teamID, channelID, threadTS)
	if err := t.Validate(); err != nil {
		return nil, goerr.Wrap(err, "invalid thread", goerr.V("thread_id", t.ID))
	}

	// Deep copy to avoid external modifications
	threadCopy := *t
	c.threads[t.ID] = &threadCopy

	return &threadCopy, nil
}

// GetThread retrieves a thread from memory
func (c *Client) GetThread(ctx context.Context, id types.ThreadID) (*slack.Thread, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	t, exists := c.threads[id]
	if !exists {
		return nil, goerr.Wrap(slack.ErrThreadNotFound, "thread not found", goerr.V("thread_id", id))
	}

	// Return a copy to avoid external modifications
	threadCopy := *t
	return &threadCopy, nil
}

// GetThreadByTS retrieves a thread by channel ID and thread timestamp
func (c *Client) GetThreadByTS(ctx context.Context, channelID, threadTS string) (*slack.Thread, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, t := range c.threads {
		if t.ChannelID == channelID && t.ThreadTS == threadTS {
			// Return a copy to avoid external modifications
			threadCopy := *t
			return &threadCopy, nil
		}
	}

	return nil, goerr.Wrap(slack.ErrThreadNotFound, "thread not found",
		goerr.V("channel_id", channelID),
		goerr.V("thread_ts", threadTS))
}

// PutThreadMessage stores a message in a thread
func (c *Client) PutThreadMessage(ctx context.Context, threadID types.ThreadID, msg *slack.Message) error {
	if err := msg.Validate(); err != nil {
		return goerr.Wrap(err, "invalid message", goerr.V("thread_id", threadID), goerr.V("message_id", msg.ID))
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if thread exists
	if _, exists := c.threads[threadID]; !exists {
		return goerr.Wrap(slack.ErrThreadNotFound, "thread not found", goerr.V("thread_id", threadID))
	}

	// Deep copy to avoid external modifications
	msgCopy := *msg
	c.messages[threadID] = append(c.messages[threadID], &msgCopy)

	return nil
}

// GetThreadMessages retrieves all messages in a thread
func (c *Client) GetThreadMessages(ctx context.Context, threadID types.ThreadID) ([]*slack.Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if thread exists
	if _, exists := c.threads[threadID]; !exists {
		return nil, goerr.Wrap(slack.ErrThreadNotFound, "thread not found", goerr.V("thread_id", threadID))
	}

	msgs := c.messages[threadID]
	if msgs == nil {
		return []*slack.Message{}, nil
	}

	// Return copies to avoid external modifications
	result := make([]*slack.Message, len(msgs))
	for i, msg := range msgs {
		msgCopy := *msg
		result[i] = &msgCopy
	}

	return result, nil
}

// PutHistory stores a history record
func (c *Client) PutHistory(ctx context.Context, history *slack.History) error {
	if err := history.Validate(); err != nil {
		return goerr.Wrap(err, "invalid history", goerr.V("history_id", history.ID))
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if thread exists
	if _, exists := c.threads[history.ThreadID]; !exists {
		return goerr.Wrap(slack.ErrThreadNotFound, "thread not found", goerr.V("thread_id", history.ThreadID))
	}

	// Deep copy to avoid external modifications
	historyCopy := *history
	c.histories[history.ID] = &historyCopy

	return nil
}

// GetLatestHistory retrieves the most recent history for a thread
func (c *Client) GetLatestHistory(ctx context.Context, threadID types.ThreadID) (*slack.History, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if thread exists
	if _, exists := c.threads[threadID]; !exists {
		return nil, goerr.Wrap(slack.ErrThreadNotFound, "thread not found", goerr.V("thread_id", threadID))
	}

	var latestHistory *slack.History
	for _, h := range c.histories {
		if h.ThreadID == threadID {
			if latestHistory == nil || h.CreatedAt.After(latestHistory.CreatedAt) {
				latestHistory = h
			}
		}
	}

	if latestHistory == nil {
		return nil, goerr.Wrap(slack.ErrHistoryNotFound, "no history found for thread", goerr.V("thread_id", threadID))
	}

	// Return a copy to avoid external modifications
	historyCopy := *latestHistory
	return &historyCopy, nil
}

// GetHistoryByID retrieves a specific history record by ID
func (c *Client) GetHistoryByID(ctx context.Context, id types.HistoryID) (*slack.History, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	h, exists := c.histories[id]
	if !exists {
		return nil, goerr.Wrap(slack.ErrHistoryNotFound, "history not found", goerr.V("history_id", id))
	}

	// Return a copy to avoid external modifications
	historyCopy := *h
	return &historyCopy, nil
}
