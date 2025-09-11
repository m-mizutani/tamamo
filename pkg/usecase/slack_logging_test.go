package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	slackapi "github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

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
	mockClient := &mock.SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Create use case
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

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
	mockClient := &mock.SlackClientMock{}

	// Mock repository error
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		return errors.New("repository error")
	}

	// Create use case
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

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
	mockClient := &mock.SlackClientMock{}

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
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

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
	mockClient := &mock.SlackClientMock{}

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
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

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
	mockClient := &mock.SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Create use case
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

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
	mockClient := &mock.SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Create use case
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

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
	mockClient := &mock.SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Create use case
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

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
	mockClient := &mock.SlackClientMock{}

	var storedMessage *slack.SlackMessageLog
	mockRepo.PutSlackMessageLogFunc = func(ctx context.Context, messageLog *slack.SlackMessageLog) error {
		storedMessage = messageLog
		return nil
	}

	// Create use case
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

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
	mockClient := &mock.SlackClientMock{}

	// Prepare test data
	testLogs := []*slack.SlackMessageLog{
		{
			ChannelID: "C123456789",
			Text:      "Message 1",
			CreatedAt: time.Now(),
		},
		{
			ChannelID: "C987654321", // Different channel
			Text:      "Message 2",
			CreatedAt: time.Now(),
		},
	}

	mockRepo.GetSlackMessageLogsFunc = func(ctx context.Context, channel string, from, to *time.Time, limit, offset int) ([]*slack.SlackMessageLog, error) {
		if channel == "C123456789" {
			return []*slack.SlackMessageLog{testLogs[0]}, nil
		}
		return []*slack.SlackMessageLog{}, nil
	}

	// Create use case
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

	// Execute
	logs, err := uc.GetMessageLogs(ctx, "C123456789", nil, nil, 10, 0)
	gt.NoError(t, err)

	// Verify filtering
	gt.Equal(t, len(logs), 1)
	gt.Equal(t, logs[0].ChannelID, "C123456789")
	gt.Equal(t, logs[0].Text, "Message 1")
}

func testGetMessageLogs(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &mock.SlackClientMock{}

	// Prepare test data
	testLogs := []*slack.SlackMessageLog{
		{
			ChannelID: "C123456789",
			Text:      "Message 1",
			CreatedAt: time.Now(),
		},
		{
			ChannelID: "C123456789",
			Text:      "Message 2",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	mockRepo.GetSlackMessageLogsFunc = func(ctx context.Context, channel string, from, to *time.Time, limit, offset int) ([]*slack.SlackMessageLog, error) {
		return testLogs, nil
	}

	// Create use case
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

	// Execute
	logs, err := uc.GetMessageLogs(ctx, "", nil, nil, 10, 0)
	gt.NoError(t, err)

	// Verify
	gt.Equal(t, len(logs), 2)
}

func testGetMessageLogsWithPagination(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRepo := &SlackMessageLogRepositoryMock{}
	mockClient := &mock.SlackClientMock{}

	// Prepare test data
	allLogs := make([]*slack.SlackMessageLog, 5)
	for i := 0; i < 5; i++ {
		allLogs[i] = &slack.SlackMessageLog{
			ChannelID: "C123456789",
			Text:      "Message " + string(rune('A'+i)),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}

	mockRepo.GetSlackMessageLogsFunc = func(ctx context.Context, channel string, from, to *time.Time, limit, offset int) ([]*slack.SlackMessageLog, error) {
		end := offset + limit
		if end > len(allLogs) {
			end = len(allLogs)
		}
		if offset >= len(allLogs) {
			return []*slack.SlackMessageLog{}, nil
		}
		return allLogs[offset:end], nil
	}

	// Create use case
	uc := usecase.New(
		usecase.WithSlackClient(mockClient),
		usecase.WithSlackMessageLogRepository(mockRepo),
	)

	// Test first page
	page1, err := uc.GetMessageLogs(ctx, "C123456789", nil, nil, 2, 0)
	gt.NoError(t, err)
	gt.Equal(t, len(page1), 2)
	gt.Equal(t, page1[0].Text, "Message A")
	gt.Equal(t, page1[1].Text, "Message B")

	// Test second page
	page2, err := uc.GetMessageLogs(ctx, "C123456789", nil, nil, 2, 2)
	gt.NoError(t, err)
	gt.Equal(t, len(page2), 2)
	gt.Equal(t, page2[0].Text, "Message C")
	gt.Equal(t, page2[1].Text, "Message D")

	// Test third page (partial)
	page3, err := uc.GetMessageLogs(ctx, "C123456789", nil, nil, 2, 4)
	gt.NoError(t, err)
	gt.Equal(t, len(page3), 1)
	gt.Equal(t, page3[0].Text, "Message E")
}
