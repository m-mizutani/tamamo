package database_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
)

// testUserRepository tests a UserRepository implementation
func testUserRepository(t *testing.T, repo interfaces.UserRepository) {
	ctx := context.Background()

	t.Run("CreateAndGetUser", func(t *testing.T) {
		u := user.NewUser("U123456789", "Test User", "Test Display", "test@example.com", "T123456789")

		// Create user
		err := repo.Create(ctx, u)
		gt.NoError(t, err)

		// Get user by ID
		retrieved, err := repo.GetByID(ctx, u.ID)
		gt.NoError(t, err)
		gt.Equal(t, retrieved.ID, u.ID)
		gt.Equal(t, retrieved.SlackID, u.SlackID)
		gt.Equal(t, retrieved.SlackName, u.SlackName)
		gt.Equal(t, retrieved.Email, u.Email)
		gt.Equal(t, retrieved.TeamID, u.TeamID)

		// Get user by SlackID and TeamID
		retrieved2, err := repo.GetBySlackIDAndTeamID(ctx, u.SlackID, u.TeamID)
		gt.NoError(t, err)
		gt.Equal(t, retrieved2.ID, u.ID)
		gt.Equal(t, retrieved2.SlackID, u.SlackID)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		u := user.NewUser("U987654321", "Original Name", "Original Display", "original@example.com", "T987654321")

		// Create user
		err := repo.Create(ctx, u)
		gt.NoError(t, err)

		// Update user info
		u.UpdateSlackInfo("Updated Name", "Updated Display", "updated@example.com")
		err = repo.Update(ctx, u)
		gt.NoError(t, err)

		// Retrieve updated user
		retrieved, err := repo.GetByID(ctx, u.ID)
		gt.NoError(t, err)
		gt.Equal(t, retrieved.SlackName, "Updated Name")
		gt.Equal(t, retrieved.Email, "updated@example.com")
	})

	t.Run("GetNonExistentUser", func(t *testing.T) {
		// Try to get user that doesn't exist
		nonExistentID := types.NewUserID(ctx)
		_, err := repo.GetByID(ctx, nonExistentID)
		gt.Error(t, err)

		// Try to get user by SlackID and TeamID that doesn't exist
		_, err = repo.GetBySlackIDAndTeamID(ctx, "UNONEXISTENT", "TNONEXISTENT")
		gt.Error(t, err)
	})

	t.Run("CreateDuplicateUser", func(t *testing.T) {
		u1 := user.NewUser("U111111111", "User One", "Display One", "user1@example.com", "T111111111")
		u2 := user.NewUser("U111111111", "User Two", "Display Two", "user2@example.com", "T111111111")

		// Create first user
		err := repo.Create(ctx, u1)
		gt.NoError(t, err)

		// Try to create second user with same SlackID and TeamID
		err = repo.Create(ctx, u2)
		gt.Error(t, err)
	})

	t.Run("InvalidUser", func(t *testing.T) {
		// Create user with empty SlackID
		invalidUser := &user.User{
			ID:        types.NewUserID(ctx),
			SlackID:   "", // Invalid: empty SlackID
			SlackName: "Test User",
			Email:     "test@example.com",
			TeamID:    "T123456789",
		}

		err := repo.Create(ctx, invalidUser)
		gt.Error(t, err)
	})
}

func TestMemoryUserRepository(t *testing.T) {
	repo := memory.NewUserRepository()
	testUserRepository(t, repo)
	testNotionIntegration(t, repo)
}

// testNotionIntegration tests Notion integration methods
func testNotionIntegration(t *testing.T, repo interfaces.UserRepository) {
	ctx := context.Background()
	userID := "test-user-notion"

	t.Run("SaveAndGetNotionIntegration", func(t *testing.T) {
		// Create a Notion integration
		notion := integration.NewNotionIntegration(userID)
		notion.UpdateWorkspaceInfo("ws-123", "Test Workspace", "https://example.com/icon.png", "bot-456")
		notion.UpdateTokens("test-access-token")

		// Save the integration
		err := repo.SaveNotionIntegration(ctx, notion)
		gt.NoError(t, err)

		// Retrieve the integration
		retrieved, err := repo.GetNotionIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, retrieved).NotNil()
		gt.V(t, retrieved.UserID).Equal(userID)
		gt.V(t, retrieved.WorkspaceID).Equal("ws-123")
		gt.V(t, retrieved.WorkspaceName).Equal("Test Workspace")
		gt.V(t, retrieved.WorkspaceIcon).Equal("https://example.com/icon.png")
		gt.V(t, retrieved.BotID).Equal("bot-456")
		gt.V(t, retrieved.AccessToken).Equal("test-access-token")
	})

	t.Run("GetNonExistentNotionIntegration", func(t *testing.T) {
		retrieved, err := repo.GetNotionIntegration(ctx, "non-existent-user")
		gt.NoError(t, err)
		gt.V(t, retrieved).Nil()
	})

	t.Run("UpdateNotionIntegration", func(t *testing.T) {
		// Create and save initial integration
		notion := integration.NewNotionIntegration(userID)
		notion.UpdateWorkspaceInfo("ws-original", "Original Workspace", "https://example.com/old.png", "bot-old")
		notion.UpdateTokens("old-token")
		err := repo.SaveNotionIntegration(ctx, notion)
		gt.NoError(t, err)

		// Update the integration
		notion.UpdateWorkspaceInfo("ws-updated", "Updated Workspace", "https://example.com/new.png", "bot-new")
		notion.UpdateTokens("new-token")
		err = repo.SaveNotionIntegration(ctx, notion)
		gt.NoError(t, err)

		// Verify the update
		retrieved, err := repo.GetNotionIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, retrieved).NotNil()
		gt.V(t, retrieved.WorkspaceID).Equal("ws-updated")
		gt.V(t, retrieved.WorkspaceName).Equal("Updated Workspace")
		gt.V(t, retrieved.AccessToken).Equal("new-token")
	})

	t.Run("DeleteNotionIntegration", func(t *testing.T) {
		// Create and save an integration
		notion := integration.NewNotionIntegration(userID)
		notion.UpdateTokens("test-token")
		err := repo.SaveNotionIntegration(ctx, notion)
		gt.NoError(t, err)

		// Delete the integration
		err = repo.DeleteNotionIntegration(ctx, userID)
		gt.NoError(t, err)

		// Verify it's deleted
		retrieved, err := repo.GetNotionIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, retrieved).Nil()
	})

	t.Run("DeleteNonExistentNotionIntegration", func(t *testing.T) {
		// Should not error when deleting non-existent integration
		err := repo.DeleteNotionIntegration(ctx, "non-existent-user")
		gt.NoError(t, err)
	})
}

// Note: Firestore tests would be added here if TEST_FIRESTORE_PROJECT is set
// func TestFirestoreUserRepository(t *testing.T) {
//     // Skip if no Firestore configuration
//     projectID := os.Getenv("TEST_FIRESTORE_PROJECT")
//     if projectID == "" {
//         t.Skip("TEST_FIRESTORE_PROJECT not set")
//     }
//
//     databaseID := os.Getenv("TEST_FIRESTORE_DATABASE")
//     if databaseID == "" {
//         databaseID = "(default)"
//     }
//
//     client, err := firestore.NewClient(context.Background(), projectID, firestore.DatabaseID(databaseID))
//     gt.NoError(t, err)
//     defer client.Close()
//
//     repo := &firestoreRepo.Client{Client: client}
//     testUserRepository(t, repo)
// }
