package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
)

// testThreadRepository runs common tests for any ThreadRepository implementation
func testThreadRepository(t *testing.T, repo interfaces.ThreadRepository) {
	ctx := context.Background()

	t.Run("GetOrPutThread", func(t *testing.T) {
		// Create a thread using GetOrPutThread
		th, err := repo.GetOrPutThread(ctx, "team123", "channel456", "1234567890.123456")
		gt.NoError(t, err)

		// Retrieve the thread
		retrieved, err := repo.GetThread(ctx, th.ID)
		gt.NoError(t, err)
		gt.Equal(t, th.ID, retrieved.ID)
		gt.Equal(t, th.TeamID, retrieved.TeamID)
		gt.Equal(t, th.ChannelID, retrieved.ChannelID)
		gt.Equal(t, th.ThreadTS, retrieved.ThreadTS)
	})

	t.Run("GetNonExistentThread", func(t *testing.T) {
		nonExistentID := types.NewThreadID(ctx)
		_, err := repo.GetThread(ctx, nonExistentID)
		gt.Error(t, err)
	})

	t.Run("GetOrPutThreadIdempotent", func(t *testing.T) {
		// Create thread first time
		th1, err := repo.GetOrPutThread(ctx, "team-idem", "channel-idem", "ts-idem")
		gt.NoError(t, err)

		// Call again with same parameters
		th2, err := repo.GetOrPutThread(ctx, "team-idem", "channel-idem", "ts-idem")
		gt.NoError(t, err)

		// Should return the same thread
		gt.Equal(t, th1.ID, th2.ID)
	})

	t.Run("PutAndGetThreadMessages", func(t *testing.T) {
		// Create thread first
		th, err := repo.GetOrPutThread(ctx, "team123", "channel456", "1234567890.123456")
		gt.NoError(t, err)

		// Add messages
		msg1 := &slack.Message{
			ID:        types.NewMessageID(ctx),
			ThreadID:  th.ID,
			UserID:    "user1",
			UserName:  "Alice",
			Text:      "Hello",
			Timestamp: "1234567890.123457",
			CreatedAt: time.Now(),
		}

		msg2 := &slack.Message{
			ID:        types.NewMessageID(ctx),
			ThreadID:  th.ID,
			UserID:    "user2",
			UserName:  "Bob",
			Text:      "Hi",
			Timestamp: "1234567890.123458",
			CreatedAt: time.Now(),
		}

		gt.NoError(t, repo.PutThreadMessage(ctx, th.ID, msg1))
		gt.NoError(t, repo.PutThreadMessage(ctx, th.ID, msg2))

		// Retrieve and verify messages
		messages, err := repo.GetThreadMessages(ctx, th.ID)
		gt.NoError(t, err)
		gt.A(t, messages).Length(2)
		gt.Equal(t, msg1.ID, messages[0].ID)
		gt.Equal(t, msg2.ID, messages[1].ID)
	})

	t.Run("GetMessagesFromNonExistentThread", func(t *testing.T) {
		nonExistentID := types.NewThreadID(ctx)
		_, err := repo.GetThreadMessages(ctx, nonExistentID)
		gt.Error(t, err)
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
		err := repo.PutThreadMessage(ctx, nonExistentID, msg)
		gt.Error(t, err)
	})

	t.Run("GetThreadByTS", func(t *testing.T) {
		// Create thread
		th, err := repo.GetOrPutThread(ctx, "team-byts", "channel-byts", "ts-byts")
		gt.NoError(t, err)

		// Retrieve by channel and timestamp
		retrieved, err := repo.GetThreadByTS(ctx, "channel-byts", "ts-byts")
		gt.NoError(t, err)
		gt.Equal(t, th.ID, retrieved.ID)
		gt.Equal(t, th.TeamID, retrieved.TeamID)
		gt.Equal(t, th.ChannelID, retrieved.ChannelID)
		gt.Equal(t, th.ThreadTS, retrieved.ThreadTS)
	})

	t.Run("GetThreadByTSNonExistent", func(t *testing.T) {
		_, err := repo.GetThreadByTS(ctx, "nonexistent-channel", "nonexistent-ts")
		gt.Error(t, err)
	})
}

func TestMemoryRepository(t *testing.T) {
	repo := memory.New()
	testThreadRepository(t, repo)
}

func TestFirestoreRepository(t *testing.T) {
	t.Skip("Firestore tests require emulator setup")

	// Uncomment when running with emulator
	// client, err := firestore.New(context.Background(), "test-project", "")
	// if err != nil {
	// 	t.Fatalf("Failed to create Firestore client: %v", err)
	// }
	// defer client.Close()
	// testThreadRepository(t, client)
}
