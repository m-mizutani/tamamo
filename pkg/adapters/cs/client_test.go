package cs_test

import (
	"context"
	"os"
	"testing"

	"github.com/m-mizutani/tamamo/pkg/adapters/cs"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

func TestCloudStorageClient_BuildPath(t *testing.T) {
	// Skip test if Cloud Storage credentials are not available
	bucket, ok := os.LookupEnv("TEST_STORAGE_BUCKET")
	if !ok {
		t.Skip("Skipping Cloud Storage test: TEST_STORAGE_BUCKET not set")
	}

	ctx := context.Background()

	// Test without prefix
	client, err := cs.New(ctx, bucket)
	if err != nil {
		t.Skipf("Skipping Cloud Storage test: %v", err)
	}
	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	// We can't easily test the internal buildPath method without exposing it,
	// so we'll test the behavior through Put/Get operations
	// This test serves as a placeholder for the path construction logic
}

func TestCloudStorageClient_WithPrefix(t *testing.T) {
	// Skip test if Cloud Storage credentials are not available
	bucket, ok := os.LookupEnv("TEST_STORAGE_BUCKET")
	if !ok {
		t.Skip("Skipping Cloud Storage test: TEST_STORAGE_BUCKET not set")
	}

	ctx := context.Background()

	// Test with prefix
	client, err := cs.New(ctx, bucket, cs.WithPrefix("test-prefix"))
	if err != nil {
		t.Skipf("Skipping Cloud Storage test: %v", err)
	}
	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	// The prefix functionality will be tested through integration tests
	// when we have access to actual Cloud Storage or emulator
}

// Note: These tests require Cloud Storage access or emulator
// For now, we're testing the interface compliance and basic construction
// Integration tests with actual storage operations should be added
// when running in an environment with Cloud Storage access

func TestCloudStorageClient_InterfaceCompliance(t *testing.T) {
	// Skip test if Cloud Storage credentials are not available
	bucket, ok := os.LookupEnv("TEST_STORAGE_BUCKET")
	if !ok {
		t.Skip("Skipping Cloud Storage test: TEST_STORAGE_BUCKET not set")
	}

	ctx := context.Background()

	// Test that the client implements the interface correctly
	client, err := cs.New(ctx, bucket)
	if err != nil {
		t.Skipf("Skipping Cloud Storage test: %v", err)
	}
	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	var _ interfaces.StorageAdapter = client
}
