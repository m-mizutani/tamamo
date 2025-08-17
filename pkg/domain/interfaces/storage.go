package interfaces

import (
	"context"
	"errors"
)

var (
	// Storage errors
	ErrStorageKeyNotFound = errors.New("storage key not found")
)

// StorageAdapter provides abstraction for storing and retrieving data
type StorageAdapter interface {
	// Put stores data with the given key
	Put(ctx context.Context, key string, data []byte) error

	// Get retrieves data by the given key
	Get(ctx context.Context, key string) ([]byte, error)
}
