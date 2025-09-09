package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gt"
	server "github.com/m-mizutani/tamamo/pkg/controller/http"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
	"github.com/m-mizutani/tamamo/pkg/service/jira"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

func TestJiraAuthController_HandleOAuthCallback(t *testing.T) {
	// Setup test dependencies
	mockUserRepo := &mock.UserRepositoryMock{
		SaveJiraIntegrationFunc: func(ctx context.Context, integration *integration.JiraIntegration) error {
			return nil
		},
	}

	oauthConfig := jira.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/jira/callback",
	}
	oauthService := jira.NewOAuthService(oauthConfig)
	jiraUseCases := usecase.NewJiraIntegrationUseCases(mockUserRepo, oauthService)

	// Create controller
	controller := server.NewJiraAuthController(jiraUseCases, oauthService)

	t.Run("OAuth error response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/auth/jira/callback?error=access_denied&error_description=User%20denied%20access", nil)
		rec := httptest.NewRecorder()

		controller.HandleOAuthCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusBadRequest)
		gt.V(t, rec.Header().Get("Content-Type")).Equal("text/html; charset=utf-8")
	})

	t.Run("missing authorization code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/auth/jira/callback", nil)
		rec := httptest.NewRecorder()

		controller.HandleOAuthCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusBadRequest)
		gt.S(t, rec.Body.String()).Contains("Missing required parameters")
	})

	t.Run("missing state parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/auth/jira/callback?code=test-code", nil)
		rec := httptest.NewRecorder()

		controller.HandleOAuthCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusBadRequest)
		gt.S(t, rec.Body.String()).Contains("Missing required parameters")
	})

	t.Run("missing state cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/auth/jira/callback?code=test-code&state=test-state", nil)
		rec := httptest.NewRecorder()

		controller.HandleOAuthCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusBadRequest)
		gt.S(t, rec.Body.String()).Contains("Invalid state parameter")
	})
}

func TestJiraIntegrationUseCases_IntegrationFlow(t *testing.T) {
	ctx := context.Background()

	// Mock repository
	savedIntegration := (*integration.JiraIntegration)(nil)
	mockUserRepo := &mock.UserRepositoryMock{
		SaveJiraIntegrationFunc: func(ctx context.Context, integration *integration.JiraIntegration) error {
			savedIntegration = integration
			return nil
		},
		GetJiraIntegrationFunc: func(ctx context.Context, userID string) (*integration.JiraIntegration, error) {
			if savedIntegration != nil && savedIntegration.UserID == userID {
				return savedIntegration, nil
			}
			return nil, goerr.New("not found")
		},
		DeleteJiraIntegrationFunc: func(ctx context.Context, userID string) error {
			if savedIntegration != nil && savedIntegration.UserID == userID {
				savedIntegration = nil
				return nil
			}
			return goerr.New("not found")
		},
	}

	oauthConfig := jira.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/jira/callback",
	}
	oauthService := jira.NewOAuthService(oauthConfig)

	useCases := usecase.NewJiraIntegrationUseCases(mockUserRepo, oauthService)

	testUserID := "test-user-123"

	t.Run("initiate OAuth flow", func(t *testing.T) {
		rec := httptest.NewRecorder()
		oauthURL, err := useCases.InitiateOAuth(ctx, rec, testUserID)
		gt.NoError(t, err)
		gt.S(t, oauthURL).Contains("https://auth.atlassian.com/authorize")
		gt.S(t, oauthURL).Contains("client_id=test-client-id")
	})

	t.Run("simulate successful integration save", func(t *testing.T) {
		// Simulate saving an integration (normally done by OAuth callback)
		testIntegration := &integration.JiraIntegration{
			UserID:          testUserID,
			CloudID:         "test-cloud-id",
			SiteURL:         "test.atlassian.net",
			AccessToken:     "test-access-token",
			RefreshToken:    "test-refresh-token",
			TokenExpiresAt:  time.Now().Add(1 * time.Hour),
			Scopes:          []string{"read:jira-user", "read:jira-work"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err := mockUserRepo.SaveJiraIntegration(ctx, testIntegration)
		gt.NoError(t, err)
	})

	t.Run("get integration after connection", func(t *testing.T) {
		integration, err := useCases.GetIntegration(ctx, testUserID)
		gt.NoError(t, err)
		gt.V(t, integration).NotNil()
		gt.Equal(t, integration.UserID, testUserID)
		gt.Equal(t, integration.SiteURL, "test.atlassian.net")
	})

	t.Run("disconnect integration", func(t *testing.T) {
		err := useCases.Disconnect(ctx, testUserID)
		gt.NoError(t, err)

		// Verify integration was deleted
		_, err = useCases.GetIntegration(ctx, testUserID)
		gt.Error(t, err)
	})
}