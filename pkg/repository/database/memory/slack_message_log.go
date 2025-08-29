package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// slackMessageLogStorage handles in-memory storage for Slack message logs
type slackMessageLogStorage struct {
	mu   sync.RWMutex
	logs map[types.MessageID]*slack.SlackMessageLog
}

// newSlackMessageLogStorage creates a new slack message log storage
func newSlackMessageLogStorage() *slackMessageLogStorage {
	return &slackMessageLogStorage{
		logs: make(map[types.MessageID]*slack.SlackMessageLog),
	}
}

// PutSlackMessageLog stores a Slack message log entry
func (s *slackMessageLogStorage) PutSlackMessageLog(ctx context.Context, messageLog *slack.SlackMessageLog) error {
	if messageLog == nil {
		return goerr.Wrap(ErrNilPointer, "messageLog cannot be nil")
	}
	if messageLog.ID == "" {
		return goerr.Wrap(ErrInvalidInput, "messageLog.ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy to avoid external modifications
	logCopy := *messageLog
	s.logs[messageLog.ID] = &logCopy

	return nil
}

// GetSlackMessageLogs retrieves message logs with filtering
func (s *slackMessageLogStorage) GetSlackMessageLogs(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Apply default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	var results []*slack.SlackMessageLog

	for _, log := range s.logs {
		// Channel filter
		if channel != "" && log.ChannelID != channel {
			continue
		}

		// Time range filters
		if from != nil && log.CreatedAt.Before(*from) {
			continue
		}
		if to != nil && log.CreatedAt.After(*to) {
			continue
		}

		// Create a copy to avoid external modifications
		logCopy := *log
		results = append(results, &logCopy)
	}

	// Sort by CreatedAt descending (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	// Apply offset
	if offset > 0 {
		if offset >= len(results) {
			return []*slack.SlackMessageLog{}, nil
		}
		results = results[offset:]
	}

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// Extend Client to implement SlackMessageLogRepository
func (c *Client) slackMessageLogStorage() *slackMessageLogStorage {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize slack message log storage if not already done
	if c.slackMsgLogs == nil {
		c.slackMsgLogs = newSlackMessageLogStorage()
	}
	return c.slackMsgLogs
}

// PutSlackMessageLog implements SlackMessageLogRepository
func (c *Client) PutSlackMessageLog(ctx context.Context, messageLog *slack.SlackMessageLog) error {
	return c.slackMessageLogStorage().PutSlackMessageLog(ctx, messageLog)
}

// GetSlackMessageLogs implements SlackMessageLogRepository
func (c *Client) GetSlackMessageLogs(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error) {
	return c.slackMessageLogStorage().GetSlackMessageLogs(ctx, channel, from, to, limit, offset)
}
