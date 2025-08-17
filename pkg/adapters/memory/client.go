package memory

import (
	"context"
	"sync"

	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

// Client provides in-memory storage implementation
type Client struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// New creates a new memory storage client
func New() *Client {
	return &Client{
		data: make(map[string][]byte),
	}
}

// Put stores data with the given key
func (c *Client) Put(ctx context.Context, key string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a copy of the data to avoid potential race conditions
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	c.data[key] = dataCopy

	return nil
}

// Get retrieves data by the given key
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, exists := c.data[key]
	if !exists {
		return nil, interfaces.ErrStorageKeyNotFound
	}

	// Return a copy to avoid potential modifications
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return dataCopy, nil
}

// Ensure Client implements StorageAdapter interface
var _ interfaces.StorageAdapter = (*Client)(nil)
