package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config defines the configuration for the filesystem storage adapter
type Config struct {
	BaseDirectory string      `yaml:"base_directory"`
	Permissions   os.FileMode `yaml:"permissions,omitempty"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.BaseDirectory == "" {
		return fmt.Errorf("base_directory is required")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(c.BaseDirectory)
	if err != nil {
		return fmt.Errorf("invalid base_directory: %w", err)
	}
	c.BaseDirectory = absPath

	// Set default permissions if not specified
	if c.Permissions == 0 {
		c.Permissions = 0755
	}

	return nil
}

// EnsureDirectory creates the base directory if it doesn't exist
func (c *Config) EnsureDirectory() error {
	if err := os.MkdirAll(c.BaseDirectory, c.Permissions); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}
	return nil
}
