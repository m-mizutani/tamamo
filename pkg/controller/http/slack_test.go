package http_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	server "github.com/m-mizutani/tamamo/pkg/controller/http"
	slack_ctrl "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/m-mizutani/tamamo/pkg/utils/async"
)

func TestSlackEventHandler(t *testing.T) {
	botUserID := "U12345BOT"
	signingSecret := "test-signing-secret"

	t.Run("handles URL verification challenge", func(t *testing.T) {
		// Create server without mock (URL verification doesn't need it)
		uc := usecase.New()
		slackCtrl := slack_ctrl.New(uc)
		srv := server.New(
			server.WithSlackController(slackCtrl),
		)

		// Create challenge request
		challenge := "test-challenge-string"
		body := map[string]interface{}{
			"type":      "url_verification",
			"challenge": challenge,
		}
		bodyBytes, err := json.Marshal(body)
		gt.NoError(t, err)

		req := httptest.NewRequest("POST", "/hooks/slack/event", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		// Execute
		srv.ServeHTTP(rec, req)

		// Verify response
		gt.Equal(t, rec.Code, http.StatusOK)
		gt.Equal(t, rec.Body.String(), challenge)
	})

	t.Run("handles app mention event and posts to thread", func(t *testing.T) {
		channelID := "C11111"
		threadTS := "1234567890.123456"
		userID := "U67890USER"

		// Setup mock
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channel, thread, text string) error {
				// Verify message is posted to correct thread
				gt.Equal(t, channel, channelID)
				gt.Equal(t, thread, threadTS)
				gt.S(t, text).Contains("Hello!")
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create server with mock
		uc := usecase.New(usecase.WithSlackClient(mockClient))
		slackCtrl := slack_ctrl.New(uc)
		srv := server.New(
			server.WithSlackController(slackCtrl),
		)

		// Create app mention event (format from Warren testdata)
		event := map[string]interface{}{
			"token":   "test-token",
			"team_id": "T12345",
			"type":    "event_callback",
			"event": map[string]interface{}{
				"type":      "app_mention",
				"user":      userID,
				"text":      fmt.Sprintf("<@%s> help", botUserID),
				"ts":        threadTS,
				"channel":   channelID,
				"thread_ts": threadTS,
				"event_ts":  threadTS,
			},
			"event_id":   "Ev12345",
			"event_time": 1234567890,
		}
		bodyBytes, err := json.Marshal(event)
		gt.NoError(t, err)

		req := httptest.NewRequest("POST", "/hooks/slack/event", bytes.NewReader(bodyBytes))
		// Enable sync mode for testing
		req = req.WithContext(async.WithSyncMode(req.Context()))
		rec := httptest.NewRecorder()

		// Execute
		srv.ServeHTTP(rec, req)

		// Verify response
		gt.Equal(t, rec.Code, http.StatusOK)

		// Verify mock was called (no wait needed in sync mode)
		calls := mockClient.PostMessageCalls()
		gt.Equal(t, len(calls), 1)
		if len(calls) > 0 {
			gt.Equal(t, calls[0].ChannelID, channelID)
			gt.Equal(t, calls[0].ThreadTS, threadTS)
		}
	})

	t.Run("verifies Slack signature when configured", func(t *testing.T) {
		// Setup mock
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channel, thread, text string) error {
				return nil
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create server with signature verification
		uc := usecase.New(usecase.WithSlackClient(mockClient))
		slackCtrl := slack_ctrl.New(uc)
		verifier := slack.NewVerifier(signingSecret)
		srv := server.New(
			server.WithSlackController(slackCtrl),
			server.WithSlackVerifier(verifier),
		)

		// Create app mention event (format from Warren testdata)
		event := map[string]interface{}{
			"token":   "test-token",
			"team_id": "T12345",
			"type":    "event_callback",
			"event": map[string]interface{}{
				"type":      "app_mention",
				"user":      "U67890USER",
				"text":      fmt.Sprintf("<@%s> test", botUserID),
				"ts":        "1234567890.123456",
				"channel":   "C11111",
				"thread_ts": "1234567890.123456",
				"event_ts":  "1234567890.123456",
			},
			"event_id":   "Ev12345",
			"event_time": 1234567890,
		}
		bodyBytes, err := json.Marshal(event)
		gt.NoError(t, err)

		// Test with valid signature
		t.Run("accepts valid signature", func(t *testing.T) {
			timestamp := strconv.FormatInt(time.Now().Unix(), 10)
			baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(bodyBytes))
			h := hmac.New(sha256.New, []byte(signingSecret))
			h.Write([]byte(baseString))
			signature := "v0=" + hex.EncodeToString(h.Sum(nil))

			req := httptest.NewRequest("POST", "/hooks/slack/event", bytes.NewReader(bodyBytes))
			// Enable sync mode for testing
			req = req.WithContext(async.WithSyncMode(req.Context()))
			req.Header.Set("X-Slack-Request-Timestamp", timestamp)
			req.Header.Set("X-Slack-Signature", signature)
			rec := httptest.NewRecorder()

			srv.ServeHTTP(rec, req)

			// Should accept and process the request
			gt.Equal(t, rec.Code, http.StatusOK)

			gt.Equal(t, len(mockClient.PostMessageCalls()), 1)
		})

		t.Run("rejects invalid signature", func(t *testing.T) {
			req := httptest.NewRequest("POST", "/hooks/slack/event", bytes.NewReader(bodyBytes))
			req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
			req.Header.Set("X-Slack-Signature", "v0=invalid")
			rec := httptest.NewRecorder()

			srv.ServeHTTP(rec, req)

			// Should reject the request
			gt.Equal(t, rec.Code, http.StatusUnauthorized)
		})
	})

	t.Run("always returns 200 for callback events", func(t *testing.T) {
		// Setup mock that returns error
		mockClient := &mock.SlackClientMock{
			PostMessageFunc: func(ctx context.Context, channel, thread, text string) error {
				return fmt.Errorf("mock error")
			},
			IsBotUserFunc: func(uid string) bool {
				return uid == botUserID
			},
		}

		// Create server with mock
		uc := usecase.New(usecase.WithSlackClient(mockClient))
		slackCtrl := slack_ctrl.New(uc)
		srv := server.New(
			server.WithSlackController(slackCtrl),
		)

		// Create app mention event (format from Warren testdata)
		event := map[string]interface{}{
			"token":   "test-token",
			"team_id": "T12345",
			"type":    "event_callback",
			"event": map[string]interface{}{
				"type":      "app_mention",
				"user":      "U67890USER",
				"text":      fmt.Sprintf("<@%s> test", botUserID),
				"ts":        "1234567890.123456",
				"channel":   "C11111",
				"thread_ts": "1234567890.123456",
				"event_ts":  "1234567890.123456",
			},
			"event_id":   "Ev12345",
			"event_time": 1234567890,
		}
		bodyBytes, err := json.Marshal(event)
		gt.NoError(t, err)

		req := httptest.NewRequest("POST", "/hooks/slack/event", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		// Execute
		srv.ServeHTTP(rec, req)

		// Should still return 200 even though processing failed
		gt.Equal(t, rec.Code, http.StatusOK)
	})
}
