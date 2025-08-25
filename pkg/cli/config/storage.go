package config

import (
	"context"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/adapters/cs"
	"github.com/m-mizutani/tamamo/pkg/adapters/fs"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/urfave/cli/v3"
)

// Storage contains configuration for storage adapters
type Storage struct {
	// Cloud Storage configuration
	Bucket string
	Prefix string

	// File System storage configuration
	FSPath string
}

// Flags returns CLI flags for Storage configuration
func (s *Storage) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "cloud-storage-bucket",
			Sources:     cli.EnvVars("TAMAMO_CLOUD_STORAGE_BUCKET"),
			Usage:       "Cloud Storage bucket for storage",
			Destination: &s.Bucket,
		},
		&cli.StringFlag{
			Name:        "cloud-storage-prefix",
			Sources:     cli.EnvVars("TAMAMO_CLOUD_STORAGE_PREFIX"),
			Usage:       "Prefix for Cloud Storage objects",
			Destination: &s.Prefix,
		},
		&cli.StringFlag{
			Name:        "file-storage-path",
			Usage:       "Path for file system storage",
			Sources:     cli.EnvVars("TAMAMO_FILE_STORAGE_PATH"),
			Destination: &s.FSPath,
		},
	}
}

// SetDefaults sets default values for Storage configuration
func (s *Storage) SetDefaults() {
	// Don't set defaults - require explicit configuration
}

// Validate validates the Storage configuration
func (s *Storage) Validate() error {
	// At least one storage backend must be configured
	hasCloudStorage := s.Bucket != ""
	hasFileStorage := s.FSPath != ""

	if !hasCloudStorage && !hasFileStorage {
		return goerr.New("at least one storage backend must be configured: use --cloud-storage-bucket for cloud storage or --file-storage-path for file system")
	}

	return nil
}

// HasCloudStorage returns true if cloud storage is configured
func (s *Storage) HasCloudStorage() bool {
	return s.Bucket != ""
}

// CreateAdapter creates appropriate storage adapter based on configuration
func (s *Storage) CreateAdapter(ctx context.Context) (interfaces.StorageAdapter, func(), error) {
	if s.HasCloudStorage() {
		// Use Cloud Storage
		opts := []cs.Option{}
		if s.Prefix != "" {
			opts = append(opts, cs.WithPrefix(s.Prefix))
		}

		csClient, err := cs.New(ctx, s.Bucket, opts...)
		if err != nil {
			return nil, nil, goerr.Wrap(err, "failed to create Cloud Storage client")
		}

		cleanup := func() {
			_ = csClient.Close() // #nosec G104 - Close error handled gracefully in cleanup
		}

		return csClient, cleanup, nil
	} else if s.FSPath != "" {
		// Use file system storage
		fsClient, err := fs.New(&fs.Config{BaseDirectory: s.FSPath})
		if err != nil {
			return nil, nil, goerr.Wrap(err, "failed to create file system storage adapter")
		}

		return fsClient, nil, nil
	} else {
		return nil, nil, goerr.New("no storage backend configured")
	}
}
