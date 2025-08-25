package image_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	imageModel "github.com/m-mizutani/tamamo/pkg/domain/model/image"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	imageService "github.com/m-mizutani/tamamo/pkg/service/image"
)

// mockStorageAdapter implements interfaces.StorageAdapter for testing
type mockStorageAdapter struct {
	storage map[string][]byte
}

func newMockStorageAdapter() *mockStorageAdapter {
	return &mockStorageAdapter{
		storage: make(map[string][]byte),
	}
}

func (m *mockStorageAdapter) Put(ctx context.Context, key string, data []byte) error {
	m.storage[key] = data
	return nil
}

func (m *mockStorageAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	data, exists := m.storage[key]
	if !exists {
		return nil, interfaces.ErrStorageKeyNotFound
	}
	return data, nil
}

func TestProcessor_ProcessAndStore(t *testing.T) {
	ctx := context.Background()

	// Setup dependencies
	validator := imageService.NewValidator()
	storage := newMockStorageAdapter()
	repository := memory.NewAgentImageRepository()
	config := imageService.DefaultProcessorConfig()

	// Create a nil mock agent repository for tests that don't need Agent updates
	processor := imageService.NewProcessor(validator, storage, repository, nil, config)

	// Test data
	agentID := types.NewUUID(ctx)
	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	// Process and store image
	agentImage, err := processor.ProcessAndStore(ctx, agentID, reader, "image/jpeg", int64(len(jpegData)))
	if err != nil {
		t.Fatalf("Failed to process and store image: %v", err)
	}

	// Verify agent image
	if agentImage.AgentID != agentID {
		t.Errorf("Expected agent ID %s, got %s", agentID, agentImage.AgentID)
	}
	if agentImage.ContentType != "image/jpeg" {
		t.Errorf("Expected content type image/jpeg, got %s", agentImage.ContentType)
	}
	if agentImage.FileSize != int64(len(jpegData)) {
		t.Errorf("Expected file size %d, got %d", len(jpegData), agentImage.FileSize)
	}

	// Verify storage keys
	if agentImage.StorageKey == "" {
		t.Error("Storage key should not be empty")
	}

	// Verify thumbnails were created
	expectedThumbnailSizes := []string{"small", "medium"}
	for _, size := range expectedThumbnailSizes {
		if _, exists := agentImage.ThumbnailKeys[size]; !exists {
			t.Errorf("Expected thumbnail size %s not found", size)
		}
	}

	// Verify original image is stored
	originalData, err := storage.Get(ctx, agentImage.StorageKey)
	if err != nil {
		t.Errorf("Failed to retrieve original image: %v", err)
	}
	if len(originalData) == 0 {
		t.Error("Original image data should not be empty")
	}

	// Verify thumbnails are stored
	for size, key := range agentImage.ThumbnailKeys {
		thumbnailData, err := storage.Get(ctx, key)
		if err != nil {
			t.Errorf("Failed to retrieve thumbnail %s: %v", size, err)
		}
		if len(thumbnailData) == 0 {
			t.Errorf("Thumbnail data for size %s should not be empty", size)
		}
	}

	// Verify repository storage
	storedImage, err := repository.GetByID(ctx, agentImage.ID)
	if err != nil {
		t.Errorf("Failed to retrieve image from repository: %v", err)
	}
	if storedImage.ID != agentImage.ID {
		t.Errorf("Expected stored image ID %s, got %s", agentImage.ID, storedImage.ID)
	}
}

func TestProcessor_GetImageData(t *testing.T) {
	ctx := context.Background()

	// Setup
	validator := imageService.NewValidator()
	storage := newMockStorageAdapter()
	repository := memory.NewAgentImageRepository()
	config := imageService.DefaultProcessorConfig()

	// Create a nil mock agent repository for tests that don't need Agent updates
	processor := imageService.NewProcessor(validator, storage, repository, nil, config)

	// Create and store test image
	agentID := types.NewUUID(ctx)
	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	agentImage, err := processor.ProcessAndStore(ctx, agentID, reader, "image/jpeg", int64(len(jpegData)))
	if err != nil {
		t.Fatalf("Failed to process and store image: %v", err)
	}

	// Retrieve image data
	retrievedData, err := processor.GetImageData(ctx, agentImage)
	if err != nil {
		t.Fatalf("Failed to get image data: %v", err)
	}

	// Verify data
	if len(retrievedData) == 0 {
		t.Error("Retrieved image data should not be empty")
	}
}

func TestProcessor_GetThumbnailData(t *testing.T) {
	ctx := context.Background()

	// Setup
	validator := imageService.NewValidator()
	storage := newMockStorageAdapter()
	repository := memory.NewAgentImageRepository()
	config := imageService.DefaultProcessorConfig()

	// Create a nil mock agent repository for tests that don't need Agent updates
	processor := imageService.NewProcessor(validator, storage, repository, nil, config)

	// Create and store test image
	agentID := types.NewUUID(ctx)
	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	agentImage, err := processor.ProcessAndStore(ctx, agentID, reader, "image/jpeg", int64(len(jpegData)))
	if err != nil {
		t.Fatalf("Failed to process and store image: %v", err)
	}

	// Test retrieving thumbnails
	for size := range agentImage.ThumbnailKeys {
		thumbnailData, err := processor.GetThumbnailData(ctx, agentImage, size)
		if err != nil {
			t.Errorf("Failed to get thumbnail data for size %s: %v", size, err)
		}
		if len(thumbnailData) == 0 {
			t.Errorf("Thumbnail data for size %s should not be empty", size)
		}
	}
}

func TestProcessor_GetThumbnailData_InvalidSize(t *testing.T) {
	ctx := context.Background()

	// Setup
	validator := imageService.NewValidator()
	storage := newMockStorageAdapter()
	repository := memory.NewAgentImageRepository()
	config := imageService.DefaultProcessorConfig()

	// Create a nil mock agent repository for tests that don't need Agent updates
	processor := imageService.NewProcessor(validator, storage, repository, nil, config)

	// Create test image without storing
	agentImage := &imageModel.AgentImage{
		ThumbnailKeys: map[string]string{
			"small": "thumbnail-key",
		},
	}

	// Try to get non-existent thumbnail size
	_, err := processor.GetThumbnailData(ctx, agentImage, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent thumbnail size")
	}
}

func TestProcessor_ProcessAndStore_ValidationFailure(t *testing.T) {
	ctx := context.Background()

	// Setup
	validator := imageService.NewValidator()
	storage := newMockStorageAdapter()
	repository := memory.NewAgentImageRepository()
	config := imageService.DefaultProcessorConfig()

	// Create a nil mock agent repository for tests that don't need Agent updates
	processor := imageService.NewProcessor(validator, storage, repository, nil, config)

	// Test with invalid image data
	agentID := types.NewUUID(ctx)
	invalidData := []byte("not an image")
	reader := bytes.NewReader(invalidData)

	_, err := processor.ProcessAndStore(ctx, agentID, reader, "image/jpeg", int64(len(invalidData)))
	if err == nil {
		t.Error("Expected validation error for invalid image data")
	}
}
