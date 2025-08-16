package memory

import (
	"context"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
)

// TestHelpers provides additional methods for testing
// These methods are exported but should only be used in tests

// GetAllThreadsForTest returns all threads in memory for testing purposes
func (c *Client) GetAllThreadsForTest() []*slack.Thread {
	c.mu.RLock()
	defer c.mu.RUnlock()

	threads := make([]*slack.Thread, 0, len(c.threads))
	for _, t := range c.threads {
		// Return a copy
		threadCopy := *t
		threads = append(threads, &threadCopy)
	}
	return threads
}

// GetThreadByChannelAndTSForTest finds a thread by channel ID and thread timestamp for testing
func (c *Client) GetThreadByChannelAndTSForTest(ctx context.Context, channelID, threadTS string) (*slack.Thread, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, t := range c.threads {
		if t.ChannelID == channelID && t.ThreadTS == threadTS {
			// Return a copy
			threadCopy := *t
			return &threadCopy, nil
		}
	}

	return nil, goerr.Wrap(slack.ErrThreadNotFound, "thread not found by channel and TS",
		goerr.V("channel_id", channelID),
		goerr.V("thread_ts", threadTS))
}
