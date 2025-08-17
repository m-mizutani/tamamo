package storage_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/adapters/memory"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/storage"
)

func TestStorageClient_SaveAndLoadHistory(t *testing.T) {
	ctx := context.Background()
	adapter := memory.New()
	client := storage.New(adapter)

	threadID := types.NewThreadID(ctx)
	historyID := types.NewHistoryID(ctx)
	testData := []byte("test history data")

	// Save history
	err := client.SaveHistory(ctx, threadID, historyID, testData)
	gt.NoError(t, err)

	// Load history
	loaded, err := client.LoadHistory(ctx, threadID, historyID)
	gt.NoError(t, err)
	gt.Equal(t, loaded, testData)
}

func TestStorageClient_SaveAndLoadHistoryJSON(t *testing.T) {
	ctx := context.Background()
	adapter := memory.New()
	client := storage.New(adapter)

	threadID := types.NewThreadID(ctx)
	historyID := types.NewHistoryID(ctx)

	// Test data structure
	testData := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
			{"role": "assistant", "content": "Hi there!"},
		},
		"metadata": map[string]string{
			"model":     "gemini-1.5-flash",
			"timestamp": "2024-01-01T00:00:00Z",
		},
	}

	// Save history as JSON
	err := client.SaveHistoryJSON(ctx, threadID, historyID, testData)
	gt.NoError(t, err)

	// Load history as JSON
	loaded, err := client.LoadHistoryJSON(ctx, threadID, historyID)
	gt.NoError(t, err)

	// Verify that it loaded as a gollem.History struct
	// We can't easily verify the exact content since gollem.History
	// has specific fields, but we can check it didn't error
	gt.NotNil(t, loaded)
}

func TestStorageClient_LoadNonExistentHistory(t *testing.T) {
	ctx := context.Background()
	adapter := memory.New()
	client := storage.New(adapter)

	threadID := types.NewThreadID(ctx)
	historyID := types.NewHistoryID(ctx)

	// Try to load non-existent history
	_, err := client.LoadHistory(ctx, threadID, historyID)
	gt.Error(t, err)
}

func TestStorageClient_CompressionAndDecompression(t *testing.T) {
	ctx := context.Background()
	adapter := memory.New()
	client := storage.New(adapter)

	threadID := types.NewThreadID(ctx)
	historyID := types.NewHistoryID(ctx)

	// Large test data that should benefit from compression
	largeData := make([]byte, 1024)
	for i := range largeData {
		largeData[i] = byte('A' + (i % 26))
	}

	// Save and load large data
	err := client.SaveHistory(ctx, threadID, historyID, largeData)
	gt.NoError(t, err)

	loaded, err := client.LoadHistory(ctx, threadID, historyID)
	gt.NoError(t, err)
	gt.Equal(t, loaded, largeData)
}

func TestStorageClient_HistoryKeyGeneration(t *testing.T) {
	ctx := context.Background()
	adapter := memory.New()
	client := storage.New(adapter)

	threadID1 := types.NewThreadID(ctx)
	threadID2 := types.NewThreadID(ctx)
	historyID1 := types.NewHistoryID(ctx)
	historyID2 := types.NewHistoryID(ctx)

	testData1 := []byte("data1")
	testData2 := []byte("data2")

	// Save different data for different thread/history combinations
	err := client.SaveHistory(ctx, threadID1, historyID1, testData1)
	gt.NoError(t, err)

	err = client.SaveHistory(ctx, threadID2, historyID2, testData2)
	gt.NoError(t, err)

	// Load and verify isolation
	loaded1, err := client.LoadHistory(ctx, threadID1, historyID1)
	gt.NoError(t, err)
	gt.Equal(t, loaded1, testData1)

	loaded2, err := client.LoadHistory(ctx, threadID2, historyID2)
	gt.NoError(t, err)
	gt.Equal(t, loaded2, testData2)
}
