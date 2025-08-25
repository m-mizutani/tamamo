package memory

import (
	"context"
	"sync"

	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

// MemoryStorageAdapter implements a simple in-memory storage for testing
type MemoryStorageAdapter struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// New creates a new memory storage adapter for testing
func New() *MemoryStorageAdapter {
	return &MemoryStorageAdapter{
		data: make(map[string][]byte),
	}
}

// Put stores data in memory
func (m *MemoryStorageAdapter) Put(ctx context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a copy to avoid external modifications
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	m.data[key] = dataCopy

	return nil
}

// Get retrieves data from memory
func (m *MemoryStorageAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.data[key]
	if !exists {
		return nil, interfaces.ErrStorageKeyNotFound
	}

	// Return a copy to avoid external modifications
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}
