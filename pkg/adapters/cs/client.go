package cs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

// Client provides Google Cloud Storage implementation
type Client struct {
	client *storage.Client
	bucket string
	prefix string
}

// Option is a functional option for Client
type Option func(*Client)

// WithPrefix sets the prefix for all storage keys
func WithPrefix(prefix string) Option {
	return func(c *Client) {
		c.prefix = prefix
	}
}

// New creates a new Cloud Storage client
func New(ctx context.Context, bucketName string, opts ...Option) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create Cloud Storage client")
	}

	c := &Client{
		client: client,
		bucket: bucketName,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Close closes the Cloud Storage client
func (c *Client) Close() error {
	return c.client.Close()
}

// buildPath constructs the full path with prefix
func (c *Client) buildPath(key string) string {
	return fmt.Sprintf("%s%s", c.prefix, key)
}

// Put stores data with the given key
func (c *Client) Put(ctx context.Context, key string, data []byte) error {
	fullPath := c.buildPath(key)
	obj := c.client.Bucket(c.bucket).Object(fullPath)

	w := obj.NewWriter(ctx)
	defer w.Close()

	if _, err := w.Write(data); err != nil {
		return goerr.Wrap(err, "failed to write data to Cloud Storage",
			goerr.V("key", key),
			goerr.V("bucket", c.bucket),
			goerr.V("path", fullPath),
		)
	}

	if err := w.Close(); err != nil {
		return goerr.Wrap(err, "failed to close Cloud Storage writer",
			goerr.V("key", key),
			goerr.V("bucket", c.bucket),
			goerr.V("path", fullPath),
		)
	}

	return nil
}

// Get retrieves data by the given key
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	fullPath := c.buildPath(key)
	obj := c.client.Bucket(c.bucket).Object(fullPath)

	r, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, interfaces.ErrStorageKeyNotFound
		}
		return nil, goerr.Wrap(err, "failed to create Cloud Storage reader",
			goerr.V("key", key),
			goerr.V("bucket", c.bucket),
			goerr.V("path", fullPath),
		)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read data from Cloud Storage",
			goerr.V("key", key),
			goerr.V("bucket", c.bucket),
			goerr.V("path", fullPath),
		)
	}

	return data, nil
}

// Ensure Client implements StorageAdapter interface
var _ interfaces.StorageAdapter = (*Client)(nil)
