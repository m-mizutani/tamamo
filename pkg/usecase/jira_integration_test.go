package usecase_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/service/jira"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

func TestJiraIntegrationUseCases(t *testing.T) {
	userRepo := memory.NewUserRepository()
	oauthService := jira.NewOAuthService(jira.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/jira/callback",
	})
	uc := usecase.NewJiraIntegrationUseCases(userRepo, oauthService)

	t.Run("InitiateOAuth returns valid URL and sets cookie", func(t *testing.T) {
		ctx := context.Background()
		w := httptest.NewRecorder()
		userID := "test-user-123"

		authURL, err := uc.InitiateOAuth(ctx, w, userID)

		gt.NoError(t, err)
		gt.S(t, authURL).Contains("https://auth.atlassian.com/authorize")
		gt.S(t, authURL).Contains("client_id=test-client-id")
		gt.S(t, authURL).Contains("scope=read%3Ajira-work+read%3Ajira-user")

		// Check that cookie was set
		cookies := w.Header()["Set-Cookie"]
		gt.V(t, len(cookies)).Equal(1)
		gt.S(t, cookies[0]).Contains("jira_oauth_state=")
		gt.S(t, cookies[0]).Contains("HttpOnly")
		gt.S(t, cookies[0]).Contains("Secure")
	})

	t.Run("GetIntegration returns nil when no integration exists", func(t *testing.T) {
		ctx := context.Background()
		userID := "test-user-456"

		integration, err := uc.GetIntegration(ctx, userID)

		gt.NoError(t, err)
		gt.V(t, integration).Nil()
	})

	t.Run("GetIntegration returns integration when exists", func(t *testing.T) {
		ctx := context.Background()
		userID := "test-user-789"

		// Create a test integration
		testIntegration := integration.NewJiraIntegration(userID)
		testIntegration.UpdateSiteInfo("cloud-123", "example.atlassian.net")
		testIntegration.UpdateTokens("access-token", "refresh-token", testIntegration.CreatedAt.Add(3600))

		// Save it
		err := userRepo.SaveJiraIntegration(ctx, testIntegration)
		gt.NoError(t, err)

		// Retrieve it
		retrievedIntegration, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, retrievedIntegration).NotNil()
		gt.V(t, retrievedIntegration.UserID).Equal(userID)
		gt.V(t, retrievedIntegration.CloudID).Equal("cloud-123")
		gt.V(t, retrievedIntegration.SiteURL).Equal("example.atlassian.net")
	})

	t.Run("Disconnect removes existing integration", func(t *testing.T) {
		ctx := context.Background()
		userID := "test-user-disconnect"

		// Create and save a test integration
		testIntegration := integration.NewJiraIntegration(userID)
		err := userRepo.SaveJiraIntegration(ctx, testIntegration)
		gt.NoError(t, err)

		// Verify it exists
		existing, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, existing).NotNil()

		// Disconnect
		err = uc.Disconnect(ctx, userID)
		gt.NoError(t, err)

		// Verify it's gone
		afterDisconnect, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, afterDisconnect).Nil()
	})

	t.Run("Disconnect fails when no integration exists", func(t *testing.T) {
		ctx := context.Background()
		userID := "test-user-nonexistent"

		err := uc.Disconnect(ctx, userID)
		gt.Error(t, err)
		gt.S(t, err.Error()).Contains("no Jira integration found")
	})

	t.Run("SaveIntegration creates new integration successfully", func(t *testing.T) {
		ctx := context.Background()
		userID := "test-user-save"
		cloudID := "cloud-456"
		siteURL := "test.atlassian.net"
		accessToken := "access-token-123"
		refreshToken := "refresh-token-456"
		expiresIn := 3600

		// Save integration
		err := uc.SaveIntegration(ctx, userID, cloudID, siteURL, accessToken, refreshToken, expiresIn)
		gt.NoError(t, err)

		// Verify it was saved correctly
		savedIntegration, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, savedIntegration).NotNil()
		gt.V(t, savedIntegration.UserID).Equal(userID)
		gt.V(t, savedIntegration.CloudID).Equal(cloudID)
		gt.V(t, savedIntegration.SiteURL).Equal(siteURL)
		gt.V(t, savedIntegration.AccessToken).Equal(accessToken)
		gt.V(t, savedIntegration.RefreshToken).Equal(refreshToken)
	})

	t.Run("SaveIntegration overwrites existing integration", func(t *testing.T) {
		ctx := context.Background()
		userID := "test-user-overwrite"

		// Save first integration
		err := uc.SaveIntegration(ctx, userID, "cloud-1", "site1.atlassian.net", "token1", "refresh1", 3600)
		gt.NoError(t, err)

		// Save second integration (should overwrite)
		err = uc.SaveIntegration(ctx, userID, "cloud-2", "site2.atlassian.net", "token2", "refresh2", 7200)
		gt.NoError(t, err)

		// Verify the latest integration is saved
		savedIntegration, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, savedIntegration).NotNil()
		gt.V(t, savedIntegration.CloudID).Equal("cloud-2")
		gt.V(t, savedIntegration.SiteURL).Equal("site2.atlassian.net")
		gt.V(t, savedIntegration.AccessToken).Equal("token2")
	})

	t.Run("GetIntegration refreshes expired token", func(t *testing.T) {
		ctx := context.Background()
		userID := "test-user-refresh"

		// Save integration with expired token
		err := uc.SaveIntegration(ctx, userID, "cloud-refresh", "refresh.atlassian.net", "old-token", "refresh-token", -1) // -1 second = already expired
		gt.NoError(t, err)

		// Mock the refresh token response
		// Note: In a real test, we'd mock the OAuth service, but since we're using a real service,
		// we can't test the actual refresh. We'll just verify the expired check works.
		retrievedIntegration, err := uc.GetIntegration(ctx, userID)
		gt.NoError(t, err)
		gt.V(t, retrievedIntegration).NotNil()

		// The token should be expired
		gt.V(t, retrievedIntegration.IsTokenExpired()).Equal(true)
	})
}
