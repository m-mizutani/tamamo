package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/slack-go/slack/slackevents"
)

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
				gt.Equal(t, threadTS, threadTS) // Should reply in thread
				gt.S(t, text).Contains("Hello! You mentioned me with: help me")
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

		// Verify mock was called
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)
	})

	t.Run("responds to bot mention without message", func(t *testing.T) {
		// Setup mock
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channelID, threadTS, text string) error {
				gt.S(t, text).Contains("How can I help you today?")
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

		// Verify mock was called
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		call := mockClient.PostMessageCalls()[0]
		gt.Equal(t, call.ThreadTS, "1234567890.123456") // Should use message TS as thread
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

		// Verify mock was not called for posting
		gt.Equal(t, len(mockClient.PostMessageCalls()), 0)
		gt.Equal(t, len(mockClient.IsBotUserCalls()), 1)
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

				// Verify mock was called
				calls := mockClient.PostMessageCalls()
				gt.Equal(t, len(calls), 1)
				gt.Equal(t, calls[0].ThreadTS, tc.expectedTS)
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

		// Verify slack client was still called
		gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
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

		// Verify no thread was created
		threads := repo.GetAllThreadsForTest()
		gt.A(t, threads).Length(0)
	})
}
