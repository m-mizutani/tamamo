package memory_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/adapters/memory"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

func TestMemoryClient_Put_Get(t *testing.T) {
	ctx := context.Background()
	client := memory.New()

	// Test data
	key := "test-key"
	data := []byte("test data")

	// Put data
	err := client.Put(ctx, key, data)
	gt.NoError(t, err)

	// Get data
	retrieved, err := client.Get(ctx, key)
	gt.NoError(t, err)
	gt.Equal(t, retrieved, data)
}

func TestMemoryClient_GetNonExistentKey(t *testing.T) {
	ctx := context.Background()
	client := memory.New()

	// Try to get non-existent key
	_, err := client.Get(ctx, "non-existent")
	gt.Error(t, err)
	gt.Equal(t, err, interfaces.ErrStorageKeyNotFound)
}

func TestMemoryClient_PutOverwrite(t *testing.T) {
	ctx := context.Background()
	client := memory.New()

	key := "test-key"
	data1 := []byte("first data")
	data2 := []byte("second data")

	// Put first data
	err := client.Put(ctx, key, data1)
	gt.NoError(t, err)

	// Put second data (overwrite)
	err = client.Put(ctx, key, data2)
	gt.NoError(t, err)

	// Get data should return second data
	retrieved, err := client.Get(ctx, key)
	gt.NoError(t, err)
	gt.Equal(t, retrieved, data2)
}

func TestMemoryClient_DataIsolation(t *testing.T) {
	ctx := context.Background()
	client := memory.New()

	key := "test-key"
	originalData := []byte("original")

	// Put data
	err := client.Put(ctx, key, originalData)
	gt.NoError(t, err)

	// Modify original data
	originalData[0] = 'X'

	// Retrieved data should not be affected
	retrieved, err := client.Get(ctx, key)
	gt.NoError(t, err)
	gt.Equal(t, retrieved[0], byte('o')) // Should still be 'o', not 'X'

	// Modify retrieved data
	retrieved[0] = 'Y'

	// Get again to ensure stored data is not affected
	retrieved2, err := client.Get(ctx, key)
	gt.NoError(t, err)
	gt.Equal(t, retrieved2[0], byte('o')) // Should still be 'o', not 'Y'
}
