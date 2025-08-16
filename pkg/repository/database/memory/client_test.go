package memory_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
)

func TestMemoryClient_Thread(t *testing.T) {
	ctx := context.Background()
	client := memory.New()

	t.Run("GetOrPutThread", func(t *testing.T) {
		// Create a thread using GetOrPutThread
		th, err := client.GetOrPutThread(ctx, "team123", "channel456", "1234567890.123456")
		gt.NoError(t, err)
		gt.NotNil(t, th)

		// Retrieve the thread
		retrieved, err := client.GetThread(ctx, th.ID)
		gt.NoError(t, err)
		gt.Equal(t, th.ID, retrieved.ID)
		gt.Equal(t, th.TeamID, retrieved.TeamID)
		gt.Equal(t, th.ChannelID, retrieved.ChannelID)
		gt.Equal(t, th.ThreadTS, retrieved.ThreadTS)

		// Test idempotency - should return same thread
		th2, err := client.GetOrPutThread(ctx, "team123", "channel456", "1234567890.123456")
		gt.NoError(t, err)
		gt.Equal(t, th.ID, th2.ID)
	})

	t.Run("GetNonExistentThread", func(t *testing.T) {
		nonExistentID := types.NewThreadID(ctx)
		_, err := client.GetThread(ctx, nonExistentID)
		gt.Error(t, err)
	})

	t.Run("GetOrPutThreadIdempotent", func(t *testing.T) {
		// Test that calling GetOrPutThread multiple times with same parameters returns same thread
		th1, err := client.GetOrPutThread(ctx, "team-idem", "channel-idem", "ts-idem")
		gt.NoError(t, err)
		gt.NotNil(t, th1)

		th2, err := client.GetOrPutThread(ctx, "team-idem", "channel-idem", "ts-idem")
		gt.NoError(t, err)
		gt.Equal(t, th1.ID, th2.ID)

		// Different TS should create different thread
		th3, err := client.GetOrPutThread(ctx, "team-idem", "channel-idem", "ts-different")
		gt.NoError(t, err)
		gt.NotEqual(t, th1.ID, th3.ID)
	})
}

func TestMemoryClient_Message(t *testing.T) {
	ctx := context.Background()
	client := memory.New()

	// Create thread first
	th, err := client.GetOrPutThread(ctx, "team123", "channel456", "1234567890.123456")
	gt.NoError(t, err)

	t.Run("PutAndGetMessages", func(t *testing.T) {
		// Add messages
		msg1 := &slack.Message{
			ID:        types.NewMessageID(ctx),
			ThreadID:  th.ID,
			UserID:    "user1",
			UserName:  "Alice",
			Text:      "Hello world",
			Timestamp: "1234567890.123457",
			CreatedAt: time.Now(),
		}

		msg2 := &slack.Message{
			ID:        types.NewMessageID(ctx),
			ThreadID:  th.ID,
			UserID:    "user2",
			UserName:  "Bob",
			Text:      "Hi there",
			Timestamp: "1234567890.123458",
			CreatedAt: time.Now(),
		}

		gt.NoError(t, client.PutThreadMessage(ctx, th.ID, msg1))
		gt.NoError(t, client.PutThreadMessage(ctx, th.ID, msg2))

		// Retrieve messages
		messages, err := client.GetThreadMessages(ctx, th.ID)
		gt.NoError(t, err)
		gt.A(t, messages).Length(2)
		gt.Equal(t, msg1.ID, messages[0].ID)
		gt.Equal(t, msg2.ID, messages[1].ID)
	})

	t.Run("PutMessageToNonExistentThread", func(t *testing.T) {
		nonExistentID := types.NewThreadID(ctx)
		msg := &slack.Message{
			ID:        types.NewMessageID(ctx),
			ThreadID:  nonExistentID,
			UserID:    "user1",
			UserName:  "Alice",
			Text:      "Test",
			Timestamp: "1234567890.123459",
			CreatedAt: time.Now(),
		}
		err := client.PutThreadMessage(ctx, nonExistentID, msg)
		gt.Error(t, err)
	})

	t.Run("GetMessagesFromEmptyThread", func(t *testing.T) {
		emptyThread, err := client.GetOrPutThread(ctx, "team999", "channel999", "9999999999.999999")
		gt.NoError(t, err)

		messages, err := client.GetThreadMessages(ctx, emptyThread.ID)
		gt.NoError(t, err)
		gt.A(t, messages).Length(0)
	})
}

func TestMemoryClient_Concurrency(t *testing.T) {
	ctx := context.Background()
	client := memory.New()

	t.Run("ConcurrentMessageOps", func(t *testing.T) {
		// Create thread first
		th, err := client.GetOrPutThread(ctx, "team-concurrent", "channel-concurrent", "concurrent-ts")
		gt.NoError(t, err)

		var wg sync.WaitGroup

		// Concurrent writes
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				msg := &slack.Message{
					ID:        types.NewMessageID(ctx),
					ThreadID:  th.ID,
					UserID:    "user1",
					UserName:  "Alice",
					Text:      "Message",
					Timestamp: fmt.Sprintf("1234567890.%d", idx),
					CreatedAt: time.Now(),
				}
				err := client.PutThreadMessage(ctx, th.ID, msg)
				gt.NoError(t, err)
			}(i)
		}

		wg.Wait()

		// Verify all messages were added
		messages, err := client.GetThreadMessages(ctx, th.ID)
		gt.NoError(t, err)
		gt.A(t, messages).Length(100)
	})

	t.Run("ConcurrentThreadCreation", func(t *testing.T) {
		var wg sync.WaitGroup

		// Concurrent thread creation
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				_, err := client.GetOrPutThread(ctx, fmt.Sprintf("team-%d", idx), fmt.Sprintf("channel-%d", idx), fmt.Sprintf("ts-%d", idx))
				gt.NoError(t, err)
			}(i)
		}

		wg.Wait()
	})

	t.Run("DataIsolation", func(t *testing.T) {
		// Create thread and message with unique identifiers
		th, err := client.GetOrPutThread(ctx, "isolation-team", "isolation-channel", "isolation-ts")
		gt.NoError(t, err)

		msg := &slack.Message{
			ID:        types.NewMessageID(ctx),
			ThreadID:  th.ID,
			UserID:    "user1",
			UserName:  "Alice",
			Text:      "Original text",
			Timestamp: "1234567890.123457",
			CreatedAt: time.Now(),
		}
		gt.NoError(t, client.PutThreadMessage(ctx, th.ID, msg))

		// Get messages and modify
		messages, err := client.GetThreadMessages(ctx, th.ID)
		gt.NoError(t, err)
		gt.A(t, messages).Length(1)

		// Modify the returned message
		messages[0].Text = "Modified text"

		// Get messages again and verify they're not modified
		freshMessages, err := client.GetThreadMessages(ctx, th.ID)
		gt.NoError(t, err)
		gt.Equal(t, "Original text", freshMessages[0].Text)
	})
}
