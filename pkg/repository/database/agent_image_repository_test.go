package database_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/image"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
)

func TestAgentImageRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("Memory", func(t *testing.T) {
		repo := memory.NewAgentImageRepository()
		testAgentImageRepository(t, ctx, repo)
	})
}

func testAgentImageRepository(t *testing.T, ctx context.Context, repo interfaces.AgentImageRepository) {
	t.Run("CreateAndGetByID", func(t *testing.T) {
		// Create test agent image
		agentID := types.NewUUID(ctx)
		agentImage := image.NewAgentImage(ctx, agentID, "test-storage-key.jpg", "image/jpeg", 1024, 256, 256)

		// Create
		err := repo.Create(ctx, agentImage)
		gt.NoError(t, err)

		// Get by ID
		retrieved, err := repo.GetByID(ctx, agentImage.ID)
		gt.NoError(t, err)

		// Verify fields
		gt.Equal(t, retrieved.ID, agentImage.ID)
		gt.Equal(t, retrieved.AgentID, agentImage.AgentID)
		gt.Equal(t, retrieved.StorageKey, agentImage.StorageKey)
		gt.Equal(t, retrieved.ContentType, agentImage.ContentType)
		gt.Equal(t, retrieved.FileSize, agentImage.FileSize)
		gt.Equal(t, retrieved.Width, agentImage.Width)
		gt.Equal(t, retrieved.Height, agentImage.Height)
	})

	t.Run("Update", func(t *testing.T) {
		// Create test agent image
		agentID := types.NewUUID(ctx)
		agentImage := image.NewAgentImage(ctx, agentID, "test-storage-key-3.jpg", "image/jpeg", 1024, 256, 256)

		// Create
		err := repo.Create(ctx, agentImage)
		gt.NoError(t, err)

		// Add thumbnail and update
		agentImage.AddThumbnail("64", "thumb-64-key.jpg")
		agentImage.AddThumbnail("128", "thumb-128-key.jpg")

		err = repo.Update(ctx, agentImage)
		gt.NoError(t, err)

		// Get and verify thumbnails
		retrieved, err := repo.GetByID(ctx, agentImage.ID)
		gt.NoError(t, err)

		thumb64, exists := retrieved.GetThumbnail("64")
		gt.Equal(t, exists, true)
		gt.Equal(t, thumb64, "thumb-64-key.jpg")

		thumb128, exists := retrieved.GetThumbnail("128")
		gt.Equal(t, exists, true)
		gt.Equal(t, thumb128, "thumb-128-key.jpg")
	})

	t.Run("NotFound", func(t *testing.T) {
		// Try to get non-existent image by ID
		nonExistentID := types.NewUUID(ctx)
		_, err := repo.GetByID(ctx, nonExistentID)
		gt.Equal(t, err, image.ErrImageNotFound)
	})
}
