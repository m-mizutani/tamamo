package usecase_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
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
	
	uc := usecase.NewNotionIntegrationUseCases(userRepo, oauthService)

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

		// Save integration
		err := uc.SaveIntegration(ctx, userID, workspaceID, workspaceName, workspaceIcon, botID, accessToken)
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

		// Save first integration
		err := uc.SaveIntegration(ctx, userID, "ws-1", "Workspace 1", "icon1.png", "bot-1", "token1")
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
}
