package storage

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// Client provides storage operations with compression
type Client struct {
	adapter interfaces.StorageAdapter
}

// New creates a new storage client
func New(adapter interfaces.StorageAdapter) *Client {
	return &Client{
		adapter: adapter,
	}
}

// SaveHistory saves history data to storage with gzip compression
func (c *Client) SaveHistory(ctx context.Context, threadID types.ThreadID, historyID types.HistoryID, data []byte) error {
	key := c.buildHistoryKey(threadID, historyID)

	// Compress data with gzip
	compressedData, err := c.compressData(data)
	if err != nil {
		return goerr.Wrap(err, "failed to compress history data",
			goerr.V("thread_id", threadID),
			goerr.V("history_id", historyID),
		)
	}

	// Store compressed data
	if err := c.adapter.Put(ctx, key, compressedData); err != nil {
		return goerr.Wrap(err, "failed to save history to storage",
			goerr.V("thread_id", threadID),
			goerr.V("history_id", historyID),
			goerr.V("key", key),
		)
	}

	return nil
}

// LoadHistory loads and decompresses history data from storage
func (c *Client) LoadHistory(ctx context.Context, threadID types.ThreadID, historyID types.HistoryID) ([]byte, error) {
	key := c.buildHistoryKey(threadID, historyID)

	// Get compressed data
	compressedData, err := c.adapter.Get(ctx, key)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to load history from storage",
			goerr.V("thread_id", threadID),
			goerr.V("history_id", historyID),
			goerr.V("key", key),
		)
	}

	// Decompress data
	data, err := c.decompressData(compressedData)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to decompress history data",
			goerr.V("thread_id", threadID),
			goerr.V("history_id", historyID),
		)
	}

	return data, nil
}

// SaveHistoryJSON saves history data as JSON with gzip compression
func (c *Client) SaveHistoryJSON(ctx context.Context, threadID types.ThreadID, historyID types.HistoryID, v interface{}) error {
	// Marshal to JSON
	jsonData, err := json.Marshal(v)
	if err != nil {
		return goerr.Wrap(err, "failed to marshal history to JSON",
			goerr.V("thread_id", threadID),
			goerr.V("history_id", historyID),
		)
	}

	return c.SaveHistory(ctx, threadID, historyID, jsonData)
}

// LoadHistoryJSON loads and unmarshals history data from storage
func (c *Client) LoadHistoryJSON(ctx context.Context, threadID types.ThreadID, historyID types.HistoryID) (gollem.History, error) {
	var history gollem.History

	data, err := c.LoadHistory(ctx, threadID, historyID)
	if err != nil {
		return history, err
	}

	// Unmarshal JSON
	if err := json.Unmarshal(data, &history); err != nil {
		return history, goerr.Wrap(err, "failed to unmarshal history JSON",
			goerr.V("thread_id", threadID),
			goerr.V("history_id", historyID),
		)
	}

	return history, nil
}

// buildHistoryKey constructs storage key for history data
func (c *Client) buildHistoryKey(threadID types.ThreadID, historyID types.HistoryID) string {
	return fmt.Sprintf("%s/history/%s.json.gz", threadID, historyID)
}

// compressData compresses data using gzip
func (c *Client) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return nil, goerr.Wrap(err, "failed to write data to gzip writer")
	}

	if err := writer.Close(); err != nil {
		return nil, goerr.Wrap(err, "failed to close gzip writer")
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func (c *Client) decompressData(compressedData []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create gzip reader")
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read from gzip reader")
	}

	return data, nil
}
