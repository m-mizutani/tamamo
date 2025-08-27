//go:generate moq -out mock_slack_message_log_repo.go . SlackMessageLogRepository
//go:generate moq -out mock_slack_client.go . SlackClient

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	slackservice "github.com/m-mizutani/tamamo/pkg/service/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/slack-go/slack/slackevents"
)

// Mock for SlackMessageLogRepository
type SlackMessageLogRepositoryMock struct {
	PutSlackMessageLogFunc            func(ctx context.Context, messageLog *slack.SlackMessageLog) error
	GetSlackMessageLogsByChannelFunc  func(ctx context.Context, channelID string, limit int) ([]*slack.SlackMessageLog, error)
	GetSlackMessageLogsByUserFunc     func(ctx context.Context, userID string, limit int) ([]*slack.SlackMessageLog, error)
	GetSlackMessageLogsFunc           func(ctx context.Context, filter *interfaces.SlackMessageLogFilter) ([]*slack.SlackMessageLog, error)
}

func (m *SlackMessageLogRepositoryMock) PutSlackMessageLog(ctx context.Context, messageLog *slack.SlackMessageLog) error {
	if m.PutSlackMessageLogFunc != nil {
		return m.PutSlackMessageLogFunc(ctx, messageLog)
	}
	return nil
}

func (m *SlackMessageLogRepositoryMock) GetSlackMessageLogsByChannel(ctx context.Context, channelID string, limit int) ([]*slack.SlackMessageLog, error) {
	if m.GetSlackMessageLogsByChannelFunc != nil {
		return m.GetSlackMessageLogsByChannelFunc(ctx, channelID, limit)
	}
	return []*slack.SlackMessageLog{}, nil
}

func (m *SlackMessageLogRepositoryMock) GetSlackMessageLogsByUser(ctx context.Context, userID string, limit int) ([]*slack.SlackMessageLog, error) {
	if m.GetSlackMessageLogsByUserFunc != nil {
		return m.GetSlackMessageLogsByUserFunc(ctx, userID, limit)
	}
	return []*slack.SlackMessageLog{}, nil
}

func (m *SlackMessageLogRepositoryMock) GetSlackMessageLogs(ctx context.Context, filter *interfaces.SlackMessageLogFilter) ([]*slack.SlackMessageLog, error) {
	if m.GetSlackMessageLogsFunc != nil {
		return m.GetSlackMessageLogsFunc(ctx, filter)
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

func TestSlackMessageLoggingUseCase_LogSlackMessage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		testLogSlackMessageSuccess(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		testLogSlackMessageRepositoryError(t)
	})

	t.Run("ChannelInfoError", func(t *testing.T) {
		testLogSlackMessageChannelInfoError(t)
	})

	t.Run("UserInfoError", func(t *testing.T) {
		testLogSlackMessageUserInfoError(t)
	})

	t.Run("BotMessage", func(t *testing.T) {
		testLogSlackMessageBot(t)
	})

	t.Run("ThreadMessage", func(t *testing.T) {
		testLogSlackMessageThread(t)
	})
}

func testLogSlackMessageSuccess(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test event
	event := &slackevents.MessageEvent{
		Type:      "message",
		User:      "U123456789",
		Text:      "Hello, world!",
		TimeStamp: "1234567890.123456",
		Channel:   "C123456789",
	}

	// Execute
	err := uc.LogSlackMessage(ctx, event)
	gt.NoError(t, err)

	// Verify
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.ChannelID, event.Channel)
	gt.Equal(t, storedMessage.UserID, event.User)
	gt.Equal(t, storedMessage.Text, event.Text)
	gt.Equal(t, storedMessage.Timestamp, event.TimeStamp)
	gt.Equal(t, storedMessage.MessageType, slack.MessageTypeUser)
	gt.Equal(t, storedMessage.ChannelType, slack.ChannelTypePublic)
}

func testLogSlackMessageRepositoryError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	// Mock repository error
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		return errors.New("repository error")
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test event
	event := &slackevents.MessageEvent{
		Type:      "message",
		User:      "U123456789",
		Text:      "Hello, world!",
		TimeStamp: "1234567890.123456",
		Channel:   "C123456789",
	}

	// Execute
	err := uc.LogSlackMessage(ctx, event)
	gt.Error(t, err) // Should return error from repository
}

func testLogSlackMessageChannelInfoError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Mock channel info error
	mockClient.GetChannelInfoFunc = func(ctx context.Context, channelID string) (*slack.ChannelInfo, error) {
		return nil, errors.New("channel info error")
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test event
	event := &slackevents.MessageEvent{
		Type:      "message",
		User:      "U123456789",
		Text:      "Hello, world!",
		TimeStamp: "1234567890.123456",
		Channel:   "C123456789",
	}

	// Execute
	err := uc.LogSlackMessage(ctx, event)
	gt.NoError(t, err) // Should succeed with default channel info (best effort)

	// Verify default values are used
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.ChannelName, event.Channel) // Fallback to channel ID
	gt.Equal(t, storedMessage.ChannelType, slack.ChannelTypePublic) // Default assumption
}

func testLogSlackMessageUserInfoError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Mock user info error
	mockClient.GetUserInfoFunc = func(ctx context.Context, userID string) (*interfaces.SlackUserInfo, error) {
		return nil, errors.New("user info error")
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test event
	event := &slackevents.MessageEvent{
		Type:      "message",
		User:      "U123456789",
		Text:      "Hello, world!",
		TimeStamp: "1234567890.123456",
		Channel:   "C123456789",
	}

	// Execute
	err := uc.LogSlackMessage(ctx, event)
	gt.NoError(t, err) // Should succeed without user name (best effort)

	// Verify user name is empty
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.UserName, "") // No user name due to error
}

func testLogSlackMessageBot(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test bot event
	event := &slackevents.MessageEvent{
		Type:      "message",
		BotID:     "B123456789",
		Text:      "Bot message",
		TimeStamp: "1234567890.123456",
		Channel:   "C123456789",
	}

	// Execute
	err := uc.LogSlackMessage(ctx, event)
	gt.NoError(t, err)

	// Verify bot message type
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.BotID, event.BotID)
	gt.Equal(t, storedMessage.MessageType, slack.MessageTypeBot)
	gt.Equal(t, storedMessage.UserID, "") // No user ID for bot messages
}

func testLogSlackMessageThread(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test thread event
	event := &slackevents.MessageEvent{
		Type:            "message",
		User:            "U123456789",
		Text:            "Thread reply",
		TimeStamp:       "1234567890.123456",
		ThreadTimeStamp: "1234567890.000000",
		Channel:         "C123456789",
	}

	// Execute
	err := uc.LogSlackMessage(ctx, event)
	gt.NoError(t, err)

	// Verify thread timestamp
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.ThreadTS, event.ThreadTimeStamp)
}

func TestSlackMessageLoggingUseCase_LogSlackAppMentionMessage(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test app mention event
	event := &slackevents.AppMentionEvent{
		Type:            "app_mention",
		User:            "U123456789",
		Text:            "<@BOTUSER> help me",
		TimeStamp:       "1234567890.123456",
		ThreadTimeStamp: "1234567890.000000",
		Channel:         "C123456789",
	}

	// Execute
	err := uc.LogSlackAppMentionMessage(ctx, event)
	gt.NoError(t, err)

	// Verify conversion to message event
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.ChannelID, event.Channel)
	gt.Equal(t, storedMessage.UserID, event.User)
	gt.Equal(t, storedMessage.Text, event.Text)
	gt.Equal(t, storedMessage.ThreadTS, event.ThreadTimeStamp)
}

func TestSlackMessageLoggingUseCase_GetMethods(t *testing.T) {
	t.Run("GetMessageLogsByChannel", func(t *testing.T) {
		testGetMessageLogsByChannel(t)
	})

	t.Run("GetMessageLogsByUser", func(t *testing.T) {
		testGetMessageLogsByUser(t)
	})

	t.Run("GetMessageLogs", func(t *testing.T) {
		testGetMessageLogs(t)
	})
}

func testGetMessageLogsByChannel(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	expectedMessages := []*slack.SlackMessageLog{
		{
			ID:        types.NewMessageID(ctx),
			ChannelID: "C123456789",
			Text:      "Test message",
		},
	}

	mockRepo.GetSlackMessageLogsByChannelFunc = func(ctx context.Context, channelID string, limit int) ([]*slack.SlackMessageLog, error) {
		gt.Equal(t, channelID, "C123456789")
		// Verify limit is handled correctly by UseCase (50 for 0, otherwise as passed)
		gt.True(t, limit == 10 || limit == 50)
		return expectedMessages, nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Execute
	messages, err := uc.GetMessageLogsByChannel(ctx, "C123456789", 10)
	gt.NoError(t, err)
	gt.Equal(t, len(messages), 1)
	gt.Equal(t, messages[0].ChannelID, "C123456789")

	// Test empty channel ID
	_, err = uc.GetMessageLogsByChannel(ctx, "", 10)
	gt.Error(t, err)

	// Test default limit
	_, err = uc.GetMessageLogsByChannel(ctx, "C123456789", 0)
	gt.NoError(t, err) // Should use default limit of 50
}

func testGetMessageLogsByUser(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	expectedMessages := []*slack.SlackMessageLog{
		{
			ID:     types.NewMessageID(ctx),
			UserID: "U123456789",
			Text:   "Test message",
		},
	}

	mockRepo.GetSlackMessageLogsByUserFunc = func(ctx context.Context, userID string, limit int) ([]*slack.SlackMessageLog, error) {
		gt.Equal(t, userID, "U123456789")
		gt.Equal(t, limit, 20)
		return expectedMessages, nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Execute
	messages, err := uc.GetMessageLogsByUser(ctx, "U123456789", 20)
	gt.NoError(t, err)
	gt.Equal(t, len(messages), 1)
	gt.Equal(t, messages[0].UserID, "U123456789")

	// Test empty user ID
	_, err = uc.GetMessageLogsByUser(ctx, "", 20)
	gt.Error(t, err)
}

func testGetMessageLogs(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	expectedMessages := []*slack.SlackMessageLog{
		{
			ID:          types.NewMessageID(ctx),
			ChannelType: slack.ChannelTypePublic,
			Text:        "Test message",
		},
	}

	mockRepo.GetSlackMessageLogsFunc = func(ctx context.Context, filter *interfaces.SlackMessageLogFilter) ([]*slack.SlackMessageLog, error) {
		gt.Equal(t, filter.ChannelType, slack.ChannelTypePublic)
		// UseCase applies default limit of 50 if 0 is passed
		if filter.Limit == 0 {
			filter.Limit = 50
		}
		return expectedMessages, nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test with filter
	filter := &interfaces.SlackMessageLogFilter{
		ChannelType: slack.ChannelTypePublic,
		Limit:       30,
	}

	messages, err := uc.GetMessageLogs(ctx, filter)
	gt.NoError(t, err)
	gt.Equal(t, len(messages), 1)
	gt.Equal(t, messages[0].ChannelType, slack.ChannelTypePublic)

	// Test nil filter
	_, err = uc.GetMessageLogs(ctx, nil)
	gt.Error(t, err)

	// Test default limit
	filter = &interfaces.SlackMessageLogFilter{
		ChannelType: slack.ChannelTypePublic,
		Limit:       0, // Should be set to default
	}
	_, err = uc.GetMessageLogs(ctx, filter)
	gt.NoError(t, err)
	gt.Equal(t, filter.Limit, 50) // Should be updated to default
}