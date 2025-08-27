package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
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
func (s *slackMessageLogStorage) GetSlackMessageLogs(ctx context.Context, filter *interfaces.SlackMessageLogFilter) ([]*slack.SlackMessageLog, error) {
	if filter == nil {
		return nil, goerr.Wrap(ErrNilPointer, "filter cannot be nil")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*slack.SlackMessageLog

	for _, log := range s.logs {
		if s.matchesFilter(log, filter) {
			// Create a copy to avoid external modifications
			logCopy := *log
			results = append(results, &logCopy)
		}
	}

	// Sort by CreatedAt descending (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	// Apply limit
	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return results, nil
}

// GetSlackMessageLogsByChannel retrieves message logs for a specific channel
func (s *slackMessageLogStorage) GetSlackMessageLogsByChannel(ctx context.Context, channelID string, limit int) ([]*slack.SlackMessageLog, error) {
	if channelID == "" {
		return nil, goerr.Wrap(ErrInvalidInput, "channelID cannot be empty")
	}

	filter := &interfaces.SlackMessageLogFilter{
		ChannelID: channelID,
		Limit:     limit,
	}

	return s.GetSlackMessageLogs(ctx, filter)
}

// GetSlackMessageLogsByUser retrieves message logs for a specific user
func (s *slackMessageLogStorage) GetSlackMessageLogsByUser(ctx context.Context, userID string, limit int) ([]*slack.SlackMessageLog, error) {
	if userID == "" {
		return nil, goerr.Wrap(ErrInvalidInput, "userID cannot be empty")
	}

	filter := &interfaces.SlackMessageLogFilter{
		UserID: userID,
		Limit:  limit,
	}

	return s.GetSlackMessageLogs(ctx, filter)
}

// matchesFilter checks if a log entry matches the given filter criteria
func (s *slackMessageLogStorage) matchesFilter(log *slack.SlackMessageLog, filter *interfaces.SlackMessageLogFilter) bool {
	// Channel ID filter
	if filter.ChannelID != "" && log.ChannelID != filter.ChannelID {
		return false
	}

	// User ID filter
	if filter.UserID != "" && log.UserID != filter.UserID {
		return false
	}

	// Channel type filter
	if filter.ChannelType != "" && log.ChannelType != filter.ChannelType {
		return false
	}

	// Message type filter
	if filter.MessageType != "" && log.MessageType != filter.MessageType {
		return false
	}

	// Time range filters
	if filter.FromTime != nil && log.CreatedAt.Before(*filter.FromTime) {
		return false
	}
	if filter.ToTime != nil && log.CreatedAt.After(*filter.ToTime) {
		return false
	}

	return true
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
func (c *Client) GetSlackMessageLogs(ctx context.Context, filter *interfaces.SlackMessageLogFilter) ([]*slack.SlackMessageLog, error) {
	return c.slackMessageLogStorage().GetSlackMessageLogs(ctx, filter)
}

// GetSlackMessageLogsByChannel implements SlackMessageLogRepository
func (c *Client) GetSlackMessageLogsByChannel(ctx context.Context, channelID string, limit int) ([]*slack.SlackMessageLog, error) {
	return c.slackMessageLogStorage().GetSlackMessageLogsByChannel(ctx, channelID, limit)
}

// GetSlackMessageLogsByUser implements SlackMessageLogRepository
func (c *Client) GetSlackMessageLogsByUser(ctx context.Context, userID string, limit int) ([]*slack.SlackMessageLog, error) {
	return c.slackMessageLogStorage().GetSlackMessageLogsByUser(ctx, userID, limit)
}
