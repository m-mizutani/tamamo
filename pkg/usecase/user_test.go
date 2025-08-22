package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/service/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

func createMockSlackClient() *mock.SlackClientMock {
	return &mock.SlackClientMock{
		GetUserProfileFunc: func(ctx context.Context, userID string) (*interfaces.SlackUserProfile, error) {
			return &interfaces.SlackUserProfile{
				ID:          userID,
				Name:        "Test User",
				DisplayName: "Test Display Name",
				Email:       "test@example.com",
				Profile: struct {
					Image24   string `json:"image_24"`
					Image32   string `json:"image_32"`
					Image48   string `json:"image_48"`
					Image72   string `json:"image_72"`
					Image192  string `json:"image_192"`
					Image512  string `json:"image_512"`
					ImageOrig string `json:"image_original"`
				}{
					Image24:   "https://example.com/avatar_24.jpg",
					Image32:   "https://example.com/avatar_32.jpg",
					Image48:   "https://example.com/avatar_48.jpg",
					Image72:   "https://example.com/avatar_72.jpg",
					Image192:  "https://example.com/avatar_192.jpg",
					Image512:  "https://example.com/avatar_512.jpg",
					ImageOrig: "https://example.com/avatar_original.jpg",
				},
			}, nil
		},
	}
}

func TestUserUseCase_GetOrCreateUser(t *testing.T) {
	ctx := context.Background()

	memoryRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
	avatarService := slack.NewAvatarService(mockSlackClient)
	uc := usecase.NewUserUseCase(memoryRepo, avatarService, mockSlackClient)

	t.Run("CreateNewUser", func(t *testing.T) {
		slackID := "U123456789"
		slackName := "Test User"
		email := "test@example.com"
		teamID := "T123456789"

		u, err := uc.GetOrCreateUser(ctx, slackID, slackName, email, teamID)
		gt.NoError(t, err)
		gt.V(t, u).NotNil()
		gt.Equal(t, u.SlackID, slackID)
		gt.Equal(t, u.SlackName, slackName)
		gt.Equal(t, u.Email, email)
		gt.Equal(t, u.TeamID, teamID)
		gt.V(t, u.ID.String()).NotEqual("")
	})

	t.Run("GetExistingUser", func(t *testing.T) {
		slackID := "U987654321"
		slackName := "Existing User"
		email := "existing@example.com"
		teamID := "T987654321"

		// Create user first
		u1, err := uc.GetOrCreateUser(ctx, slackID, slackName, email, teamID)
		gt.NoError(t, err)

		// Get the same user again
		u2, err := uc.GetOrCreateUser(ctx, slackID, slackName, email, teamID)
		gt.NoError(t, err)
		gt.Equal(t, u1.ID, u2.ID)
		gt.Equal(t, u1.SlackID, u2.SlackID)
	})

	t.Run("UpdateExistingUserInfo", func(t *testing.T) {
		slackID := "U111222333"
		originalName := "Original Name"
		updatedName := "Updated Name"
		originalEmail := "original@example.com"
		updatedEmail := "updated@example.com"
		teamID := "T111222333"

		// Create user with original info
		u1, err := uc.GetOrCreateUser(ctx, slackID, originalName, originalEmail, teamID)
		gt.NoError(t, err)
		originalTime := u1.UpdatedAt

		// Sleep to ensure timestamp difference
		time.Sleep(time.Millisecond)

		// Get user with updated info
		u2, err := uc.GetOrCreateUser(ctx, slackID, updatedName, updatedEmail, teamID)
		gt.NoError(t, err)
		gt.Equal(t, u1.ID, u2.ID)
		gt.Equal(t, u2.SlackName, updatedName)
		gt.Equal(t, u2.Email, updatedEmail)
		if !u2.UpdatedAt.After(originalTime) {
			t.Error("UpdatedAt should be more recent")
		}
	})
}

func TestUserUseCase_GetUserByID(t *testing.T) {
	ctx := context.Background()

	memoryRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
	avatarService := slack.NewAvatarService(mockSlackClient)
	uc := usecase.NewUserUseCase(memoryRepo, avatarService, mockSlackClient)

	t.Run("GetExistingUser", func(t *testing.T) {
		// Create a user first
		u, err := uc.GetOrCreateUser(ctx, "U123", "Test", "test@example.com", "T123")
		gt.NoError(t, err)

		// Get the user by ID
		retrieved, err := uc.GetUserByID(ctx, u.ID)
		gt.NoError(t, err)
		gt.Equal(t, retrieved.ID, u.ID)
		gt.Equal(t, retrieved.SlackID, u.SlackID)
	})

	t.Run("GetNonexistentUser", func(t *testing.T) {
		nonexistentID := types.UserID("nonexistent-id")
		_, err := uc.GetUserByID(ctx, nonexistentID)
		gt.Error(t, err)
	})
}

func TestUserUseCase_UpdateUser(t *testing.T) {
	ctx := context.Background()

	memoryRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
	avatarService := slack.NewAvatarService(mockSlackClient)
	uc := usecase.NewUserUseCase(memoryRepo, avatarService, mockSlackClient)

	t.Run("UpdateExistingUser", func(t *testing.T) {
		// Create a user first
		u, err := uc.GetOrCreateUser(ctx, "U123", "Original", "original@example.com", "T123")
		gt.NoError(t, err)
		originalTime := u.UpdatedAt

		// Sleep to ensure timestamp difference
		time.Sleep(time.Millisecond)

		// Update the user
		u.SlackName = "Updated Name"
		err = uc.UpdateUser(ctx, u)
		gt.NoError(t, err)

		// Verify the update
		retrieved, err := uc.GetUserByID(ctx, u.ID)
		gt.NoError(t, err)
		gt.Equal(t, retrieved.SlackName, "Updated Name")
		if !retrieved.UpdatedAt.After(originalTime) {
			t.Error("UpdatedAt should be more recent")
		}
	})
}

func TestUserUseCase_GetUserAvatar(t *testing.T) {
	ctx := context.Background()

	memoryRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
	avatarService := slack.NewAvatarService(mockSlackClient)
	uc := usecase.NewUserUseCase(memoryRepo, avatarService, mockSlackClient)

	t.Run("GetAvatarForExistingUser", func(t *testing.T) {
		t.Skip("Skipping avatar test due to HTTP client dependency - requires mock HTTP client")
		// Create a user first
		u, err := uc.GetOrCreateUser(ctx, "U123", "Test", "test@example.com", "T123")
		gt.NoError(t, err)

		// Get the avatar
		avatarData, err := uc.GetUserAvatar(ctx, u.ID, 48)
		gt.NoError(t, err)
		if len(avatarData) == 0 {
			t.Error("expected non-empty avatar data")
		}
	})

	t.Run("GetAvatarForNonexistentUser", func(t *testing.T) {
		nonexistentID := types.UserID("nonexistent-id")
		_, err := uc.GetUserAvatar(ctx, nonexistentID, 48)
		gt.Error(t, err)
	})
}

func TestUserUseCase_InvalidateUserAvatarCache(t *testing.T) {
	ctx := context.Background()

	memoryRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
	avatarService := slack.NewAvatarService(mockSlackClient)
	uc := usecase.NewUserUseCase(memoryRepo, avatarService, mockSlackClient)

	t.Run("InvalidateCacheForExistingUser", func(t *testing.T) {
		// Create a user first
		u, err := uc.GetOrCreateUser(ctx, "U123", "Test", "test@example.com", "T123")
		gt.NoError(t, err)

		// Invalidate the cache
		err = uc.InvalidateUserAvatarCache(ctx, u.ID)
		gt.NoError(t, err)
	})

	t.Run("InvalidateCacheForNonexistentUser", func(t *testing.T) {
		nonexistentID := types.UserID("nonexistent-id")
		err := uc.InvalidateUserAvatarCache(ctx, nonexistentID)
		gt.Error(t, err)
	})
}
