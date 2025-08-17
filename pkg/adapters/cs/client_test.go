package cs_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/adapters/cs"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

func TestCloudStorageClient_BuildPath(t *testing.T) {
	ctx := context.Background()

	// Test without prefix
	client, err := cs.New(ctx, "test-bucket")
	gt.NoError(t, err)
	defer client.Close()

	// We can't easily test the internal buildPath method without exposing it,
	// so we'll test the behavior through Put/Get operations
	// This test serves as a placeholder for the path construction logic
}

func TestCloudStorageClient_WithPrefix(t *testing.T) {
	ctx := context.Background()

	// Test with prefix
	client, err := cs.New(ctx, "test-bucket", cs.WithPrefix("test-prefix"))
	gt.NoError(t, err)
	defer client.Close()

	// The prefix functionality will be tested through integration tests
	// when we have access to actual Cloud Storage or emulator
}

// Note: These tests require Cloud Storage access or emulator
// For now, we're testing the interface compliance and basic construction
// Integration tests with actual storage operations should be added
// when running in an environment with Cloud Storage access

func TestCloudStorageClient_InterfaceCompliance(t *testing.T) {
	ctx := context.Background()

	// Test that the client implements the interface correctly
	client, err := cs.New(ctx, "test-bucket")
	gt.NoError(t, err)
	defer client.Close()

	var _ interfaces.StorageAdapter = client
}
