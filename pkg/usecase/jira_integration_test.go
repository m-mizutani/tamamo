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
}
