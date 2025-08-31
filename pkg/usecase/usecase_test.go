//go:generate moq -out mock_slack_message_log_repo.go . SlackMessageLogRepository
//go:generate moq -out mock_slack_client.go . SlackClient

package usecase_test

import (
	"context"
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
)

// Mock for SlackMessageLogRepository
type SlackMessageLogRepositoryMock struct {
	PutSlackMessageLogFunc  func(ctx context.Context, messageLog *slack.SlackMessageLog) error
	GetSlackMessageLogsFunc func(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error)
}

func (m *SlackMessageLogRepositoryMock) PutSlackMessageLog(ctx context.Context, messageLog *slack.SlackMessageLog) error {
	if m.PutSlackMessageLogFunc != nil {
		return m.PutSlackMessageLogFunc(ctx, messageLog)
	}
	return nil
}

func (m *SlackMessageLogRepositoryMock) GetSlackMessageLogs(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error) {
	if m.GetSlackMessageLogsFunc != nil {
		return m.GetSlackMessageLogsFunc(ctx, channel, from, to, limit, offset)
	}
	return []*slack.SlackMessageLog{}, nil
}

// Mock for SlackClient
type SlackClientMock struct {
	GetChannelInfoFunc func(ctx context.Context, channelID string) (*slack.ChannelInfo, error)
	GetUserInfoFunc    func(ctx context.Context, userID string) (*interfaces.SlackUserInfo, error)
}

func (m *SlackClientMock) PostMessage(ctx context.Context, channelID, threadTS, text string) error {
	return nil
}

func (m *SlackClientMock) PostMessageWithOptions(ctx context.Context, channelID, threadTS, text string, options *interfaces.SlackMessageOptions) error {
	return nil
}

func (m *SlackClientMock) IsBotUser(userID string) bool {
	return false
}

func (m *SlackClientMock) GetUserProfile(ctx context.Context, userID string) (*interfaces.SlackUserProfile, error) {
	return nil, nil
}

func (m *SlackClientMock) GetUserInfo(ctx context.Context, userID string) (*interfaces.SlackUserInfo, error) {
	if m.GetUserInfoFunc != nil {
		return m.GetUserInfoFunc(ctx, userID)
	}
	return &interfaces.SlackUserInfo{
		ID:          userID,
		Name:        "testuser",
		DisplayName: "Test User",
		RealName:    "Test User",
	}, nil
}

func (m *SlackClientMock) GetBotInfo(ctx context.Context, botID string) (*interfaces.SlackBotInfo, error) {
	return nil, nil
}

func (m *SlackClientMock) GetChannelInfo(ctx context.Context, channelID string) (*slack.ChannelInfo, error) {
	if m.GetChannelInfoFunc != nil {
		return m.GetChannelInfoFunc(ctx, channelID)
	}
	return &slack.ChannelInfo{
		ID:        channelID,
		Name:      "general",
		Type:      slack.ChannelTypePublic,
		IsPrivate: false,
		UpdatedAt: time.Now(),
	}, nil
}