package usecase_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
	llm_mock "github.com/m-mizutani/gollem/mock"
	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/repository/storage"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/sashabaranov/go-openai"
	"github.com/slack-go/slack/slackevents"
)

// mockStorageAdapter for testing
type mockStorageAdapter struct {
	storage map[string][]byte
}

func newMockStorageAdapter() *mockStorageAdapter {
	return &mockStorageAdapter{
		storage: make(map[string][]byte),
	}
}

func (m *mockStorageAdapter) Put(ctx context.Context, key string, data []byte) error {
	m.storage[key] = data
	return nil
}

func (m *mockStorageAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	data, exists := m.storage[key]
	if !exists {
		return nil, interfaces.ErrStorageKeyNotFound
	}
	return data, nil
}

func TestHandleSlackAppMention(t *testing.T) {
	botUserID := "U12345BOT"
	userID := "U67890USER"
	channelID := "C11111"
	threadTS := "1234567890.123456"

	t.Run("responds to bot mention with message", func(t *testing.T) {
		// Setup mock
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				// Verify the message is posted to the correct thread
				gt.Equal(t, channelID, "C11111")
				gt.Equal(t, threadTS, "1234567890.123456") // Should reply in thread
				gt.S(t, text).Contains("LLM not configured")
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase with mock
		uc := usecase.New(usecase.WithSlackClient(mockClient))

		// Create test message with bot mention
		ev := &slackevents.EventsAPIEvent{
			TeamID: "T12345",
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> help me",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: threadTS,
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, threadTS)
		gt.S(t, postMessageCall.Text).Contains("LLM not configured")
		gt.S(t, postMessageCall.Text).Contains("Please contact your administrator")

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)
	})

	t.Run("responds to bot mention without message", func(t *testing.T) {
		// Setup mock
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				gt.S(t, text).Contains("LLM not configured")
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase with mock
		uc := usecase.New(usecase.WithSlackClient(mockClient))

		// Create test message with bot mention only
		ev := &slackevents.EventsAPIEvent{
			TeamID: "T12345",
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT>",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: "", // No thread, should use message TS
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, "1234567890.123456") // Should use message TS as thread
		gt.S(t, postMessageCall.Text).Contains("LLM not configured")

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)
	})

	t.Run("ignores non-bot mentions", func(t *testing.T) {
		// Setup mock
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				t.Fatal("should not post message for non-bot mention")
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID // Different from mentioned user
			},
		}

		// Create usecase with mock
		uc := usecase.New(usecase.WithSlackClient(mockClient))

		// Create test message with non-bot mention
		ev := &slackevents.EventsAPIEvent{
			TeamID: "T12345",
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U99999OTHER> hello",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: threadTS,
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 0)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, "U99999OTHER")
	})

	t.Run("ensures reply is in thread", func(t *testing.T) {
		// Test both with and without existing thread
		testCases := []struct {
			name            string
			threadTimeStamp string
			expectedTS      string
		}{
			{
				name:            "existing thread",
				threadTimeStamp: "1234567890.111111",
				expectedTS:      "1234567890.111111",
			},
			{
				name:            "no thread (use message TS)",
				threadTimeStamp: "",
				expectedTS:      "1234567890.123456", // Message TS
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup mock to verify thread reply for this specific test case
				mockClient := &mock.SlackClientMock{
					PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
						// Must have thread TS
						gt.NotEqual(t, threadTS, "")
						gt.Equal(t, channelID, "C11111")
						gt.Equal(t, threadTS, tc.expectedTS) // Check expected TS
						return nil
					},
					IsBotUserFunc: func(uid string) bool {
						return uid == botUserID
					},
				}

				// Create usecase with mock
				uc := usecase.New(usecase.WithSlackClient(mockClient))

				ev := &slackevents.EventsAPIEvent{
					TeamID: "T12345",
					InnerEvent: slackevents.EventsAPIInnerEvent{
						Data: &slackevents.AppMentionEvent{
							User:            userID,
							Text:            "<@U12345BOT> test",
							TimeStamp:       "1234567890.123456",
							Channel:         channelID,
							ThreadTimeStamp: tc.threadTimeStamp,
						},
					},
				}
				msg := slack.NewMessage(context.Background(), ev)

				err := uc.HandleSlackAppMention(context.Background(), *msg)
				gt.NoError(t, err)

				// Verify mock call counts
				gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
				gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

				// Verify PostMessage call details
				postMessageCall := mockClient.PostMessageCalls()[0]
				gt.Equal(t, postMessageCall.ChannelID, channelID)
				gt.Equal(t, postMessageCall.ThreadTS, tc.expectedTS)
				gt.S(t, postMessageCall.Text).Contains("LLM not configured")

				// Verify IsBotUser call details
				isBotUserCall := mockClient.IsBotUserCalls()[0]
				gt.Equal(t, isBotUserCall.UserID, botUserID)
			})
		}
	})
}

func TestHandleSlackAppMentionWithRepository(t *testing.T) {
	botUserID := "U12345BOT"
	userID := "U67890USER"
	// userName := "testuser" // Not used - slack.Message uses userID as userName
	channelID := "C11111"
	teamID := "T12345"

	t.Run("records new thread and message on first mention", func(t *testing.T) {
		// Setup repository
		repo := memory.New()

		// Setup mock slack client
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase with repository
		uc := usecase.New(
			usecase.WithSlackClient(mockClient),
			usecase.WithRepository(repo),
		)

		// Create test message - first mention in a new thread
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> help me with testing",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000", // Thread TS
				},
			},
		}

		// Create message
		ctx := context.Background()
		msg := slack.NewMessage(ctx, ev)

		// Execute
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify thread was created using test helpers
		threads := repo.GetAllThreadsForTest()
		gt.A(t, threads).Length(1)

		createdThread := threads[0]
		gt.Equal(t, createdThread.TeamID, teamID)
		gt.Equal(t, createdThread.ChannelID, channelID)
		gt.Equal(t, createdThread.ThreadTS, "1234567890.100000")

		// Verify message was recorded
		messages, err := repo.GetThreadMessages(context.Background(), createdThread.ID)
		gt.NoError(t, err)
		gt.A(t, messages).Length(1)

		recordedMsg := messages[0]
		gt.Equal(t, recordedMsg.UserID, userID)
		gt.Equal(t, recordedMsg.UserName, userID) // Currently user name is same as user ID
		gt.Equal(t, recordedMsg.Text, "<@U12345BOT> help me with testing")
		gt.Equal(t, recordedMsg.Timestamp, "1234567890.123456")
		gt.Equal(t, recordedMsg.ThreadID, createdThread.ID)
	})

	t.Run("records message in existing thread", func(t *testing.T) {
		// Setup repository with existing thread
		repo := memory.New()
		ctx := context.Background()

		// Create existing thread using GetOrPutThread
		existingThread, err := repo.GetOrPutThread(ctx, teamID, channelID, "1234567890.100000")
		gt.NoError(t, err)

		// Add an existing message to the thread
		existingMsg := &slack.Message{
			ID:        types.NewMessageID(ctx),
			ThreadID:  existingThread.ID,
			UserID:    "U99999OTHER",
			UserName:  "otheruser",
			Text:      "Previous message",
			Timestamp: "1234567890.099999",
			CreatedAt: time.Now(),
		}
		err = repo.PutThreadMessage(ctx, existingThread.ID, existingMsg)
		gt.NoError(t, err)

		// Setup mock slack client
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase with repository
		uc := usecase.New(
			usecase.WithSlackClient(mockClient),
			usecase.WithRepository(repo),
		)

		// Create test message - mention in existing thread
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> continue the conversation",
					TimeStamp:       "1234567890.200000",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000", // Same thread TS
				},
			},
		}

		msg := slack.NewMessage(ctx, ev)

		// Execute
		err = uc.HandleSlackAppMention(ctx, *msg)
		gt.NoError(t, err)

		// The implementation should use GetOrPutThread, which will find the existing thread
		// since the threadTS is the same ("1234567890.100000")
		// Should record the message in the existing thread
		messages, err := repo.GetThreadMessages(ctx, existingThread.ID)
		gt.NoError(t, err)
		gt.A(t, messages).Length(2) // 1 existing + 1 new

		// Check the newly added message (should be the second one)
		newMsg := messages[1]
		gt.Equal(t, newMsg.UserID, userID)
		gt.Equal(t, newMsg.UserName, userID) // Currently user name is same as user ID
		gt.Equal(t, newMsg.Text, "<@U12345BOT> continue the conversation")
		gt.Equal(t, newMsg.Timestamp, "1234567890.200000")
	})

	t.Run("records message in channel-level mention (no thread)", func(t *testing.T) {
		// Setup repository
		repo := memory.New()

		// Setup mock slack client
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase with repository
		uc := usecase.New(
			usecase.WithSlackClient(mockClient),
			usecase.WithRepository(repo),
		)

		// Create test message - channel-level mention (no thread)
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> start new conversation",
					TimeStamp:       "1234567890.300000",
					Channel:         channelID,
					ThreadTimeStamp: "", // No thread - channel-level message
				},
			},
		}

		ctx := context.Background()
		msg := slack.NewMessage(ctx, ev)

		// Execute
		err := uc.HandleSlackAppMention(ctx, *msg)
		gt.NoError(t, err)

		// Verify thread was created with message TS as thread TS
		threads := repo.GetAllThreadsForTest()
		gt.A(t, threads).Length(1)

		createdThread := threads[0]
		gt.Equal(t, createdThread.TeamID, teamID)
		gt.Equal(t, createdThread.ChannelID, channelID)
		gt.Equal(t, createdThread.ThreadTS, "1234567890.300000") // Uses message TS when no thread

		// Verify message was recorded
		messages, err := repo.GetThreadMessages(ctx, createdThread.ID)
		gt.NoError(t, err)
		gt.A(t, messages).Length(1)

		recordedMsg := messages[0]
		gt.Equal(t, recordedMsg.UserID, userID)
		gt.Equal(t, recordedMsg.UserName, userID) // Currently user name is same as user ID
		gt.Equal(t, recordedMsg.Text, "<@U12345BOT> start new conversation")
		gt.Equal(t, recordedMsg.Timestamp, "1234567890.300000")
	})

	t.Run("does not record when repository is nil", func(t *testing.T) {
		// Setup mock slack client without repository
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase WITHOUT repository
		uc := usecase.New(usecase.WithSlackClient(mockClient))

		// Create test message
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> test without repo",
					TimeStamp:       "1234567890.400000",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000",
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute - should not panic or error
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, "1234567890.100000")
		gt.S(t, postMessageCall.Text).Contains("LLM not configured")

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)
	})

	t.Run("ignores non-bot mentions and does not record", func(t *testing.T) {
		// Setup repository
		repo := memory.New()

		// Setup mock slack client
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				t.Fatal("should not post message for non-bot mention")
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase with repository
		uc := usecase.New(
			usecase.WithSlackClient(mockClient),
			usecase.WithRepository(repo),
		)

		// Create test message with non-bot mention
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U99999OTHER> hello",
					TimeStamp:       "1234567890.500000",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000",
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 0)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, "U99999OTHER")

		// Verify no thread was created
		threads := repo.GetAllThreadsForTest()
		gt.A(t, threads).Length(0)
	})
}

// MockSession implements gollem.Session for testing
type MockSession struct {
	history             *gollem.History
	generateContentFunc func(ctx context.Context, input ...gollem.Input) (*gollem.Response, error)
	messageCount        int // Track how many messages have been generated
}

func (m *MockSession) GenerateContent(ctx context.Context, input ...gollem.Input) (*gollem.Response, error) {
	// Increment message count to simulate history growth
	m.messageCount++

	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, input...)
	}
	return &gollem.Response{
		Texts: []string{"Mock response"},
	}, nil
}

func (m *MockSession) GenerateStream(ctx context.Context, input ...gollem.Input) (<-chan *gollem.Response, error) {
	ch := make(chan *gollem.Response, 1)
	resp, err := m.GenerateContent(ctx, input...)
	if err != nil {
		close(ch)
		return ch, err
	}
	ch <- resp
	close(ch)
	return ch, nil
}

func (m *MockSession) History() *gollem.History {
	if m.history == nil {
		// Create empty Gemini history
		m.history = &gollem.History{
			LLType:  "Gemini",
			Version: 1,
		}
	}

	// Simulate that history has been populated based on message count
	// For testing, we just need ToCount() to return > 0
	if m.messageCount > 0 {
		// Set a dummy OpenAI history to make ToCount() return non-zero
		m.history.OpenAI = []openai.ChatCompletionMessage{
			{Role: "user", Content: "test"},
			{Role: "assistant", Content: "response"},
		}
	}

	return m.history
}

func TestHandleSlackAppMentionWithLLM(t *testing.T) {
	botUserID := "U12345BOT"
	userID := "U67890USER"
	channelID := "C11111"
	teamID := "T12345"

	t.Run("responds with LLM when configured", func(t *testing.T) {
		// Create mock LLM client
		mockSession := &MockSession{
			generateContentFunc: func(ctx context.Context, input ...gollem.Input) (*gollem.Response, error) {
				return &gollem.Response{
					Texts: []string{"This is an AI-powered response"},
				}, nil
			},
		}

		mockLLMClient := &llm_mock.LLMClientMock{
			NewSessionFunc: func(ctx context.Context, options ...gollem.SessionOption) (gollem.Session, error) {
				return mockSession, nil
			},
		}

		// Setup mock Slack client
		var capturedResponse string
		mockSlackClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				capturedResponse = text
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create repositories
		repo := memory.New()
		agentRepo := memory.NewAgentMemoryClient()
		storageAdapter := newMockStorageAdapter()
		storageRepo := storage.New(storageAdapter)

		// Create usecase with LLM and agent repository
		uc := usecase.New(
			usecase.WithSlackClient(mockSlackClient),
			usecase.WithRepository(repo),
			usecase.WithAgentRepository(agentRepo),
			usecase.WithStorageRepository(storageRepo),
			usecase.WithLLMClient(mockLLMClient),
		)

		// Create test message
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> !help with quantum computing",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000",
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify LLM mock call counts
		gt.Equal(t, len(mockLLMClient.NewSessionCalls()), 1)

		// Verify Slack mock call counts
		gt.Equal(t, len(mockSlackClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockSlackClient.IsBotUserCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockSlackClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, "1234567890.100000")
		gt.S(t, capturedResponse).Contains("AI-powered response")

		// Verify IsBotUser call details
		isBotUserCall := mockSlackClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)
	})

	t.Run("maintains conversation history", func(t *testing.T) {
		// Track history updates - simulate that history is populated on each call
		mockSession := &MockSession{
			history: &gollem.History{
				LLType:  "Gemini",
				Version: 1,
			},
			generateContentFunc: func(ctx context.Context, input ...gollem.Input) (*gollem.Response, error) {
				return &gollem.Response{
					Texts: []string{"Response with history"},
				}, nil
			},
		}

		mockLLMClient := &llm_mock.LLMClientMock{
			NewSessionFunc: func(ctx context.Context, options ...gollem.SessionOption) (gollem.Session, error) {
				return mockSession, nil
			},
		}

		// Setup mock Slack client
		mockSlackClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create repositories
		repo := memory.New()
		agentRepo := memory.NewAgentMemoryClient()
		storageAdapter := newMockStorageAdapter()
		storageRepo := storage.New(storageAdapter)

		// Create usecase with LLM and agent repository
		uc := usecase.New(
			usecase.WithSlackClient(mockSlackClient),
			usecase.WithRepository(repo),
			usecase.WithAgentRepository(agentRepo),
			usecase.WithStorageRepository(storageRepo),
			usecase.WithLLMClient(mockLLMClient),
		)

		// First message
		ev1 := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> !help with my question",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000",
				},
			},
		}
		msg1 := slack.NewMessage(context.Background(), ev1)

		err := uc.HandleSlackAppMention(context.Background(), *msg1)
		gt.NoError(t, err)

		// Check that history was saved by attempting to get latest history
		thread, err := repo.GetThreadByTS(context.Background(), channelID, "1234567890.100000")
		gt.NoError(t, err)
		latestHistory, err := repo.GetLatestHistory(context.Background(), thread.ID)
		gt.NoError(t, err)
		gt.NotNil(t, latestHistory)

		// Second message in same thread
		ev2 := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> ?please continue",
					TimeStamp:       "1234567890.234567",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000", // Same thread
				},
			},
		}
		msg2 := slack.NewMessage(context.Background(), ev2)

		err = uc.HandleSlackAppMention(context.Background(), *msg2)
		gt.NoError(t, err)

		// History should be updated - check we can get it
		latestHistory2, err := repo.GetLatestHistory(context.Background(), thread.ID)
		gt.NoError(t, err)
		gt.NotNil(t, latestHistory2)
		// The ID should be different from the first one
		gt.NotEqual(t, latestHistory.ID, latestHistory2.ID)
	})

	t.Run("falls back gracefully when LLM fails", func(t *testing.T) {
		// Mock LLM that fails
		mockLLMClient := &llm_mock.LLMClientMock{
			NewSessionFunc: func(ctx context.Context, options ...gollem.SessionOption) (gollem.Session, error) {
				return nil, goerr.New("LLM service unavailable")
			},
		}

		// Setup mock Slack client
		var capturedResponse string
		mockSlackClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				capturedResponse = text
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create repositories
		repo := memory.New()
		agentRepo := memory.NewAgentMemoryClient()
		storageAdapter := newMockStorageAdapter()
		storageRepo := storage.New(storageAdapter)

		// Create usecase with failing LLM and repositories
		uc := usecase.New(
			usecase.WithSlackClient(mockSlackClient),
			usecase.WithRepository(repo),
			usecase.WithAgentRepository(agentRepo),
			usecase.WithStorageRepository(storageRepo),
			usecase.WithLLMClient(mockLLMClient),
		)

		// Create test message
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> ?test",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: "",
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute - should return error but also send fallback response
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.Error(t, err) // Now we expect an error to be returned
		gt.S(t, err.Error()).Contains("failed to create LLM session")

		// Verify Slack mock call counts
		gt.Equal(t, len(mockSlackClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockSlackClient.IsBotUserCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockSlackClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, "1234567890.123456")
		gt.S(t, capturedResponse).Contains("experiencing issues")

		// Verify IsBotUser call details
		isBotUserCall := mockSlackClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)
	})
}

// TestAnalyzeThreadContext tests the thread context analysis
func TestAnalyzeThreadContext(t *testing.T) {
	botUserID := "U12345BOT"
	userID := "U67890USER"
	channelID := "C11111"
	teamID := "T12345"

	t.Run("new thread without repository", func(t *testing.T) {
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase without repository
		uc := usecase.New(usecase.WithSlackClient(mockClient))

		// Create test message
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> code-helper debug this",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000",
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute - this should work without repository
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, "1234567890.100000")
		gt.S(t, postMessageCall.Text).Contains("LLM not configured")

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)
	})

	t.Run("new thread with repository", func(t *testing.T) {
		repo := memory.New()
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase with repository
		uc := usecase.New(
			usecase.WithSlackClient(mockClient),
			usecase.WithRepository(repo),
		)

		// Create test message for new thread
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> code-helper debug this",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000",
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, "1234567890.100000")
		gt.S(t, postMessageCall.Text).Contains("LLM not configured")

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)

		// Verify thread was created
		threads := repo.GetAllThreadsForTest()
		gt.A(t, threads).Length(1)

		createdThread := threads[0]
		gt.V(t, createdThread.TeamID).Equal(teamID)
		gt.V(t, createdThread.ChannelID).Equal(channelID)
		gt.V(t, createdThread.ThreadTS).Equal("1234567890.100000")
	})

	t.Run("existing thread", func(t *testing.T) {
		repo := memory.New()
		ctx := context.Background()

		// Create existing thread
		existingThread, err := repo.GetOrPutThread(ctx, teamID, channelID, "1234567890.100000")
		gt.NoError(t, err)

		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create usecase with repository
		uc := usecase.New(
			usecase.WithSlackClient(mockClient),
			usecase.WithRepository(repo),
		)

		// Create test message for existing thread
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> continue conversation",
					TimeStamp:       "1234567890.200000",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000", // Same thread TS
				},
			},
		}
		msg := slack.NewMessage(ctx, ev)

		// Execute
		err = uc.HandleSlackAppMention(ctx, *msg)
		gt.NoError(t, err)

		// Verify mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, "1234567890.100000")
		gt.S(t, postMessageCall.Text).Contains("LLM not configured")

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)

		// Should still have only one thread
		threads := repo.GetAllThreadsForTest()
		gt.A(t, threads).Length(1)
		gt.V(t, threads[0].ID).Equal(existingThread.ID)
	})
}

// TestResolveAgent tests agent resolution functionality
func TestResolveAgent(t *testing.T) {
	botUserID := "U12345BOT"
	userID := "U67890USER"
	channelID := "C11111"
	teamID := "T12345"

	t.Run("general mode with agent repository", func(t *testing.T) {
		repo := memory.New()
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Mock LLM to enable agent functionality
		mockSession := &MockSession{
			generateContentFunc: func(ctx context.Context, input ...gollem.Input) (*gollem.Response, error) {
				return &gollem.Response{
					Texts: []string{"Welcome to Tamamo general mode!"},
				}, nil
			},
		}

		mockLLMClient := &llm_mock.LLMClientMock{
			NewSessionFunc: func(ctx context.Context, options ...gollem.SessionOption) (gollem.Session, error) {
				return mockSession, nil
			},
		}

		// Create repositories
		storageAdapter := newMockStorageAdapter()
		storageRepo := storage.New(storageAdapter)

		// Create agent repository for modern implementation (now required)
		agentRepo := memory.NewAgentMemoryClient()

		// Create usecase with all required repositories and LLM
		uc := usecase.New(
			usecase.WithSlackClient(mockClient),
			usecase.WithRepository(repo),
			usecase.WithAgentRepository(agentRepo),
			usecase.WithStorageRepository(storageRepo),
			usecase.WithLLMClient(mockLLMClient),
		)

		// Create test message for general mode
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> !help how are you?",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000",
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute - should work in general mode
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err)

		// Verify LLM mock call counts
		gt.Equal(t, len(mockLLMClient.NewSessionCalls()), 1)

		// Verify Slack mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, "1234567890.100000")
		gt.S(t, postMessageCall.Text).Contains("Welcome to Tamamo general mode!")

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)

		// Verify thread was created with general mode UUID
		threads := repo.GetAllThreadsForTest()
		gt.A(t, threads).Length(1)

		createdThread := threads[0]
		gt.V(t, createdThread.AgentUUID).NotEqual(nil)
		gt.V(t, *createdThread.AgentUUID).Equal(types.UUID("00000000-0000-0000-0000-000000000000"))
		gt.V(t, createdThread.AgentVersion).Equal("general-v1")
	})

	t.Run("agent not found error", func(t *testing.T) {
		repo := memory.New()

		// Mock agent repository that returns "not found"
		mockAgentRepo := &mock.AgentRepositoryMock{
			GetAgentByAgentIDActiveFunc: func(ctx context.Context, agentID string) (*agent.Agent, error) {
				return nil, slack.ErrAgentNotFound
			},
			ListActiveAgentsFunc: func(ctx context.Context, offset, limit int) ([]*agent.Agent, int, error) {
				// Return some sample agents for error message
				return []*agent.Agent{
					{
						ID:          types.NewUUID(ctx),
						AgentID:     "code-helper",
						Name:        "Code Helper",
						Description: "Helps with coding tasks",
						Status:      agent.StatusActive,
					},
					{
						ID:          types.NewUUID(ctx),
						AgentID:     "data-analyzer",
						Name:        "Data Analyzer",
						Description: "Analyzes data and generates insights",
						Status:      agent.StatusActive,
					},
				}, 2, nil
			},
		}

		var capturedErrorMsg string
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				capturedErrorMsg = text
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Mock LLM to enable agent functionality
		mockSession := &MockSession{
			generateContentFunc: func(ctx context.Context, input ...gollem.Input) (*gollem.Response, error) {
				return &gollem.Response{
					Texts: []string{"This should not be reached due to agent error"},
				}, nil
			},
		}

		mockLLMClient := &llm_mock.LLMClientMock{
			NewSessionFunc: func(ctx context.Context, options ...gollem.SessionOption) (gollem.Session, error) {
				return mockSession, nil
			},
		}

		// Create repositories
		storageAdapter := newMockStorageAdapter()
		storageRepo := storage.New(storageAdapter)

		// Create usecase with agent repository and LLM
		uc := usecase.New(
			usecase.WithSlackClient(mockClient),
			usecase.WithRepository(repo),
			usecase.WithAgentRepository(mockAgentRepo),
			usecase.WithStorageRepository(storageRepo),
			usecase.WithLLMClient(mockLLMClient),
		)

		// Create test message with invalid agent
		ev := &slackevents.EventsAPIEvent{
			TeamID: teamID,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{
					User:            userID,
					Text:            "<@U12345BOT> invalid-agent do something",
					TimeStamp:       "1234567890.123456",
					Channel:         channelID,
					ThreadTimeStamp: "1234567890.100000",
				},
			},
		}
		msg := slack.NewMessage(context.Background(), ev)

		// Execute - should handle error gracefully
		err := uc.HandleSlackAppMention(context.Background(), *msg)
		gt.NoError(t, err) // Should not return error, but send error message to Slack

		// Verify mock call counts
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)
		gt.Equal(t, len(mockAgentRepo.GetAgentByAgentIDActiveCalls()), 1)
		gt.Equal(t, len(mockAgentRepo.ListActiveAgentsCalls()), 1)

		// Verify PostMessage call details
		postMessageCall := mockClient.PostMessageCalls()[0]
		gt.Equal(t, postMessageCall.ChannelID, channelID)
		gt.Equal(t, postMessageCall.ThreadTS, "1234567890.100000")
		gt.V(t, strings.Contains(capturedErrorMsg, "Agent ID 'invalid-agent' not found")).Equal(true)
		gt.V(t, strings.Contains(capturedErrorMsg, "Usage: @tamamo <agent_id> [message]")).Equal(true)

		// Verify IsBotUser call details
		isBotUserCall := mockClient.IsBotUserCalls()[0]
		gt.Equal(t, isBotUserCall.UserID, botUserID)

		// Verify GetAgentByAgentIDActive call details
		getAgentCall := mockAgentRepo.GetAgentByAgentIDActiveCalls()[0]
		gt.Equal(t, getAgentCall.AgentID, "invalid-agent")

		// Verify ListActiveAgents call details
		listAgentsCall := mockAgentRepo.ListActiveAgentsCalls()[0]
		gt.Equal(t, listAgentsCall.Offset, 0)
		gt.Equal(t, listAgentsCall.Limit, 10)
	})
}
