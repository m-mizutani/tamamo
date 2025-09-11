package usecase_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/service/notion"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

func TestNotionIntegrationUseCases(t *testing.T) {
	// Setup
	userRepo := memory.NewUserRepository()
	oauthConfig := notion.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/notion/callback",
	}
	oauthService := notion.NewOAuthService(oauthConfig)
	
	// Create mock SlackClient
	mockSlackClient := &mock.SlackClientMock{
		IsWorkspaceMemberFunc: func(ctx context.Context, email string) (bool, error) {
			// By default, allow all members for tests
			return true, nil
		},
	}
	
	uc := usecase.NewNotionIntegrationUseCases(userRepo, oauthService, mockSlackClient)

	ctx := context.Background()

	t.Run("InitiateOAuth returns valid URL and sets cookie", func(t *testing.T) {
		userID := "test-user-oauth"
		w := httptest.NewRecorder()

		authURL, err := uc.InitiateOAuth(ctx, w, userID)
		gt.NoError(t, err)
		gt.B(t, authURL != "").True()

		// Check that URL contains expected parameters
		gt.B(t, len(authURL) > 0).True()
		gt.S(t, authURL).Contains("https://api.notion.com/v1/oauth/authorize")
		gt.S(t, authURL).Contains("client_id=test-client-id")
		gt.S(t, authURL).Contains("owner=user") // Notion-specific parameter

		// Check that cookie was set
		cookies := w.Result().Cookies()
		gt.V(t, len(cookies)).Equal(1)
		gt.V(t, cookies[0].Name).Equal("notion_oauth_state")
	})

	t.Run("GetIntegration returns nil when no integration exists", func(t *testing.T) {
		userID := "test-user-no-integration"

		integration, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, integration).Nil()
	})

	t.Run("GetIntegration returns integration when exists", func(t *testing.T) {
		userID := "test-user-with-integration"

		// First save an integration
		testIntegration := integration.NewNotionIntegration(userID)
		testIntegration.UpdateWorkspaceInfo("ws-123", "Test Workspace", "https://example.com/icon.png", "bot-456")
		testIntegration.UpdateTokens("test-token")
		err := userRepo.SaveNotionIntegration(ctx, testIntegration)
		gt.NoError(t, err)

		// Then retrieve it through use case
		retrieved, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, retrieved).NotNil()
		gt.V(t, retrieved.UserID).Equal(userID)
		gt.V(t, retrieved.WorkspaceID).Equal("ws-123")
		gt.V(t, retrieved.AccessToken).Equal("test-token")
	})

	t.Run("Disconnect removes existing integration", func(t *testing.T) {
		userID := "test-user-disconnect"

		// First save an integration
		testIntegration := integration.NewNotionIntegration(userID)
		testIntegration.UpdateTokens("test-token")
		err := userRepo.SaveNotionIntegration(ctx, testIntegration)
		gt.NoError(t, err)

		// Disconnect
		err = uc.Disconnect(ctx, userID)
		gt.NoError(t, err)

		// Verify it's deleted
		afterDisconnect, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, afterDisconnect).Nil()
	})

	t.Run("Disconnect fails when no integration exists", func(t *testing.T) {
		userID := "test-user-nonexistent"

		err := uc.Disconnect(ctx, userID)
		gt.Error(t, err)
		gt.S(t, err.Error()).Contains("no Notion integration found")
	})

	t.Run("SaveIntegration creates new integration successfully", func(t *testing.T) {
		ctx := context.Background()
		userID := "550e8400-e29b-41d4-a716-446655440000" // Valid UUID
		workspaceID := "ws-456"
		workspaceName := "Test Workspace"
		workspaceIcon := "https://example.com/icon.png"
		botID := "bot-789"
		accessToken := "access-token-123"

		// Create a user first (required for workspace validation)
		testUser := user.NewUser("slack-user-1", "slackname1", "Test User", "test@example.com", "team-1")
		testUser.ID = types.UserID(userID) // Override ID to match our test
		err := userRepo.Create(ctx, testUser)
		gt.NoError(t, err)

		// Save integration (workspace validation checks the user)
		err = uc.SaveIntegration(ctx, userID, workspaceID, workspaceName, workspaceIcon, botID, accessToken)
		gt.NoError(t, err)

		// Verify it was saved correctly
		savedIntegration, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, savedIntegration).NotNil()
		gt.V(t, savedIntegration.UserID).Equal(userID)
		gt.V(t, savedIntegration.WorkspaceID).Equal(workspaceID)
		gt.V(t, savedIntegration.WorkspaceName).Equal(workspaceName)
		gt.V(t, savedIntegration.WorkspaceIcon).Equal(workspaceIcon)
		gt.V(t, savedIntegration.BotID).Equal(botID)
		gt.V(t, savedIntegration.AccessToken).Equal(accessToken)
	})

	t.Run("SaveIntegration overwrites existing integration", func(t *testing.T) {
		ctx := context.Background()
		userID := "550e8400-e29b-41d4-a716-446655440001" // Valid UUID

		// Create a user first (required for workspace validation)
		testUser := user.NewUser("slack-user-2", "slackname2", "Test User 2", "test2@example.com", "team-1")
		testUser.ID = types.UserID(userID) // Override ID to match our test
		err := userRepo.Create(ctx, testUser)
		gt.NoError(t, err)

		// Save first integration
		err = uc.SaveIntegration(ctx, userID, "ws-1", "Workspace 1", "icon1.png", "bot-1", "token1")
		gt.NoError(t, err)

		// Save second integration (should overwrite)
		err = uc.SaveIntegration(ctx, userID, "ws-2", "Workspace 2", "icon2.png", "bot-2", "token2")
		gt.NoError(t, err)

		// Verify the latest integration is saved
		savedIntegration, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, savedIntegration).NotNil()
		gt.V(t, savedIntegration.WorkspaceID).Equal("ws-2")
		gt.V(t, savedIntegration.WorkspaceName).Equal("Workspace 2")
		gt.V(t, savedIntegration.AccessToken).Equal("token2")
	})

	t.Run("SaveIntegration denies non-Slack workspace members", func(t *testing.T) {
		ctx := context.Background()
		userID := "550e8400-e29b-41d4-a716-446655440002" // Valid UUID

		// Create a user first
		testUser := user.NewUser("slack-user-3", "slackname3", "Test User 3", "nonmember@example.com", "team-1")
		testUser.ID = types.UserID(userID)
		err := userRepo.Create(ctx, testUser)
		gt.NoError(t, err)

		// Create a new mock client that denies this user
		mockSlackClientDeny := &mock.SlackClientMock{
			IsWorkspaceMemberFunc: func(ctx context.Context, email string) (bool, error) {
				if email == "nonmember@example.com" {
					return false, nil // Not a workspace member
				}
				return true, nil
			},
		}

		// Create use case with deny client
		ucDeny := usecase.NewNotionIntegrationUseCases(userRepo, oauthService, mockSlackClientDeny)

		// Try to save integration (should fail)
		err = ucDeny.SaveIntegration(ctx, userID, "ws-3", "Workspace 3", "icon3.png", "bot-3", "token3")
		gt.Error(t, err)
		
		// Check that it's the correct typed error
		var gErr *goerr.Error
		gt.B(t, errors.As(err, &gErr)).True()
		gt.B(t, errors.Is(gErr.Unwrap(), auth.ErrNotWorkspaceMember)).True()

		// Verify integration was not saved
		savedIntegration, err := ucDeny.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.Nil(t, savedIntegration)
	})
}
