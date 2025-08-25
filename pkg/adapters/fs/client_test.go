package fs_test

import (
	"context"
	"os"
	"testing"

	"github.com/m-mizutani/tamamo/pkg/adapters/fs"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/image"
)

func TestClient_PutAndGet(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fs_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create client
	config := &fs.Config{
		BaseDirectory: tempDir,
		Permissions:   0755,
	}
	client, err := fs.New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	key := "test-file.txt"
	data := []byte("test data")

	// Test Put
	if err := client.Put(ctx, key, data); err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	// Test Get
	retrieved, err := client.Get(ctx, key)
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if string(retrieved) != string(data) {
		t.Errorf("Expected %s, got %s", string(data), string(retrieved))
	}
}

func TestClient_GetNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &fs.Config{
		BaseDirectory: tempDir,
		Permissions:   0755,
	}
	client, err := fs.New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	_, err = client.Get(ctx, "nonexistent-file.txt")
	if err != interfaces.ErrStorageKeyNotFound {
		t.Errorf("Expected ErrStorageKeyNotFound, got %v", err)
	}
}

func TestClient_SecurityValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_client_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &fs.Config{
		BaseDirectory: tempDir,
		Permissions:   0755,
	}
	client, err := fs.New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	data := []byte("test data")

	// Test path traversal attempts
	maliciousKeys := []string{
		"../etc/passwd",
		"..\\windows\\system32",
		"/etc/passwd",
		"file\x00.txt", // null byte
	}

	for _, key := range maliciousKeys {
		err := client.Put(ctx, key, data)
		if err != image.ErrSecurityViolation {
			t.Errorf("Expected security violation for key %s, got %v", key, err)
		}
	}
}
