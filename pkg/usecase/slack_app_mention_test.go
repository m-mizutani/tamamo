package usecase_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
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
