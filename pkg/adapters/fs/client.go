package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/image"
)

// Client provides filesystem storage implementation
type Client struct {
	baseDir     string
	permissions os.FileMode
	mu          sync.RWMutex
}

// New creates a new filesystem storage client
func New(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if err := config.EnsureDirectory(); err != nil {
		return nil, fmt.Errorf("failed to ensure directory: %w", err)
	}

	return &Client{
		baseDir:     config.BaseDirectory,
		permissions: config.Permissions,
	}, nil
}

// Put stores data with the given key
func (c *Client) Put(ctx context.Context, key string, data []byte) error {
	if err := c.validateKey(key); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), c.permissions); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Get retrieves data by the given key
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	if err := c.validateKey(key); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	filePath := c.getFilePath(key)

	// #nosec G304 - Path is validated by validateKey() function to prevent path traversal
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, interfaces.ErrStorageKeyNotFound
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// validateKey validates the storage key to prevent path traversal attacks
func (c *Client) validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") || strings.Contains(key, "\\") {
		return image.ErrSecurityViolation
	}

	// Ensure the key only contains safe characters
	for _, char := range key {
		if char < 32 || char == 127 { // Control characters
			return image.ErrSecurityViolation
		}
	}

	return nil
}

// getFilePath returns the full file path for the given key
func (c *Client) getFilePath(key string) string {
	return filepath.Join(c.baseDir, key)
}

// Ensure Client implements StorageAdapter interface
var _ interfaces.StorageAdapter = (*Client)(nil)
