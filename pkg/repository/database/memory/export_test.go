package memory

import (
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
)

// GetAllThreads returns all threads in memory for testing purposes
func (c *Client) GetAllThreads() []*slack.Thread {
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
