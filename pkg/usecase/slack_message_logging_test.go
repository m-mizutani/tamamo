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
	slackapi "github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
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

	t.Run("MessageWithFileAttachments", func(t *testing.T) {
		testLogSlackMessageWithFiles(t)
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
	err := uc.LogSlackMessage(ctx, event, "T123456789")
	gt.NoError(t, err)

	// Verify
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.TeamID, "T123456789")
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
	err := uc.LogSlackMessage(ctx, event, "T123456789")
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
	err := uc.LogSlackMessage(ctx, event, "T123456789")
	gt.NoError(t, err) // Should succeed with default channel info (best effort)

	// Verify default values are used
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.ChannelName, event.Channel)           // Fallback to channel ID
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
	err := uc.LogSlackMessage(ctx, event, "T123456789")
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
	err := uc.LogSlackMessage(ctx, event, "T123456789")
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
	err := uc.LogSlackMessage(ctx, event, "T123456789")
	gt.NoError(t, err)

	// Verify thread timestamp
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.ThreadTS, event.ThreadTimeStamp)
}

func testLogSlackMessageWithFiles(t *testing.T) {
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

	// Test event with file attachments in Message field
	event := &slackevents.MessageEvent{
		Type:      "message",
		User:      "U123456789",
		Text:      "Check out this file!",
		TimeStamp: "1234567890.123456",
		Channel:   "C123456789",
		Message: &slackapi.Msg{
			Files: []slackapi.File{
				{
					ID:         "F123456789",
					Name:       "document.pdf",
					Mimetype:   "application/pdf",
					Filetype:   "pdf",
					URLPrivate: "https://files.slack.com/files-pri/T123456789-F123456789/document.pdf",
				},
				{
					ID:         "F987654321",
					Name:       "image.png",
					Mimetype:   "image/png",
					Filetype:   "png",
					URLPrivate: "https://files.slack.com/files-pri/T123456789-F987654321/image.png",
				},
			},
		},
	}

	// Execute
	err := uc.LogSlackMessage(ctx, event, "T123456789")
	gt.NoError(t, err)

	// Verify file attachments were processed
	gt.NotNil(t, storedMessage)
	gt.Equal(t, len(storedMessage.Attachments), 2)

	// Verify first file
	gt.Equal(t, storedMessage.Attachments[0].ID, "F123456789")
	gt.Equal(t, storedMessage.Attachments[0].Name, "document.pdf")
	gt.Equal(t, storedMessage.Attachments[0].Mimetype, "application/pdf")
	gt.Equal(t, storedMessage.Attachments[0].FileType, "pdf")
	gt.Equal(t, storedMessage.Attachments[0].URL, "https://files.slack.com/files-pri/T123456789-F123456789/document.pdf")

	// Verify second file
	gt.Equal(t, storedMessage.Attachments[1].ID, "F987654321")
	gt.Equal(t, storedMessage.Attachments[1].Name, "image.png")
	gt.Equal(t, storedMessage.Attachments[1].Mimetype, "image/png")
	gt.Equal(t, storedMessage.Attachments[1].FileType, "png")
	gt.Equal(t, storedMessage.Attachments[1].URL, "https://files.slack.com/files-pri/T123456789-F987654321/image.png")
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
	err := uc.LogSlackAppMentionMessage(ctx, event, "T123456789")
	gt.NoError(t, err)

	// Verify conversion to message event
	gt.NotNil(t, storedMessage)
	gt.Equal(t, storedMessage.TeamID, "T123456789")
	gt.Equal(t, storedMessage.ChannelID, event.Channel)
	gt.Equal(t, storedMessage.UserID, event.User)
	gt.Equal(t, storedMessage.Text, event.Text)
	gt.Equal(t, storedMessage.ThreadTS, event.ThreadTimeStamp)
}

func TestSlackMessageLoggingUseCase_GetMethods(t *testing.T) {
	t.Run("GetMessageLogs", func(t *testing.T) {
		testGetMessageLogs(t)
	})

	t.Run("GetMessageLogsByChannelFilter", func(t *testing.T) {
		testGetMessageLogsByChannelFilter(t)
	})

	t.Run("GetMessageLogsWithPagination", func(t *testing.T) {
		testGetMessageLogsWithPagination(t)
	})
}

func testGetMessageLogsByChannelFilter(t *testing.T) {
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

	mockRepo.GetSlackMessageLogsFunc = func(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error) {
		gt.Equal(t, channel, "C123456789")
		gt.Equal(t, limit, 10)
		gt.Equal(t, offset, 0)
		return expectedMessages, nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test channel filter with limit and offset
	messages, err := uc.GetMessageLogs(ctx, "C123456789", nil, nil, 10, 0)
	gt.NoError(t, err)
	gt.Equal(t, len(messages), 1)
	gt.Equal(t, messages[0].ChannelID, "C123456789")
}

func testGetMessageLogs(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	expectedMessages := []*slack.SlackMessageLog{
		{
			ID:   types.NewMessageID(ctx),
			Text: "Test message",
		},
	}

	mockRepo.GetSlackMessageLogsFunc = func(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error) {
		return expectedMessages, nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test with basic parameters
	messages, err := uc.GetMessageLogs(ctx, "", nil, nil, 0, 0)
	gt.NoError(t, err)
	gt.Equal(t, len(messages), 1)

	// Test with time range and pagination
	now := time.Now()
	from := now.Add(-time.Hour)
	to := now.Add(time.Hour)

	messages, err = uc.GetMessageLogs(ctx, "C123456789", &from, &to, 20, 10)
	gt.NoError(t, err)
	gt.Equal(t, len(messages), 1)
}

func testGetMessageLogsWithPagination(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &SlackClientMock{}

	// Create test messages
	expectedMessages := []*slack.SlackMessageLog{
		{
			ID:        types.NewMessageID(ctx),
			ChannelID: "C123456789",
			Text:      "Message 1",
		},
		{
			ID:        types.NewMessageID(ctx),
			ChannelID: "C123456789",
			Text:      "Message 2",
		},
	}

	mockRepo.GetSlackMessageLogsFunc = func(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error) {
		// Verify pagination parameters
		gt.Equal(t, limit, 5)
		gt.Equal(t, offset, 10)
		return expectedMessages, nil
	}

	// Create use case
	channelCache := slackservice.NewChannelCache(mockClient, time.Hour)
	uc := usecase.NewSlackMessageLoggingUseCase(mockRepo, mockClient, channelCache)

	// Test pagination
	messages, err := uc.GetMessageLogs(ctx, "C123456789", nil, nil, 5, 10)
	gt.NoError(t, err)
	gt.Equal(t, len(messages), 2)
	gt.Equal(t, messages[0].Text, "Message 1")
	gt.Equal(t, messages[1].Text, "Message 2")
}
