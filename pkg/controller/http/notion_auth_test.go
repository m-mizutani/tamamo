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
	"github.com/m-mizutani/tamamo/pkg/service/notion"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

func TestNotionAuthController_HandleOAuthCallback(t *testing.T) {
	// Setup test dependencies
	mockUserRepo := &mock.UserRepositoryMock{
		SaveNotionIntegrationFunc: func(ctx context.Context, integration *integration.NotionIntegration) error {
			return nil
		},
	}

	oauthConfig := notion.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/oauth/notion/callback",
	}
	oauthService := notion.NewOAuthService(oauthConfig)
	notionUseCases := usecase.NewNotionIntegrationUseCases(mockUserRepo, oauthService)

	// Create controller
	controller := server.NewNotionAuthController(notionUseCases, oauthService)

	t.Run("OAuth error response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/oauth/notion/callback?error=access_denied&error_description=User%20denied%20access", nil)
		rec := httptest.NewRecorder()

		controller.HandleOAuthCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusSeeOther)
		gt.S(t, rec.Header().Get("Location")).Contains("/integrations/notion/error")
		gt.S(t, rec.Header().Get("Location")).Contains("OAuth+Error")
	})

	t.Run("missing authorization code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/oauth/notion/callback", nil)
		rec := httptest.NewRecorder()

		controller.HandleOAuthCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusBadRequest)
		gt.S(t, rec.Body.String()).Contains("Missing required parameters")
	})

	t.Run("missing state parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/oauth/notion/callback?code=test-code", nil)
		rec := httptest.NewRecorder()

		controller.HandleOAuthCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusBadRequest)
		gt.S(t, rec.Body.String()).Contains("Missing required parameters")
	})

	t.Run("missing state cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/oauth/notion/callback?code=test-code&state=test-state", nil)
		rec := httptest.NewRecorder()

		controller.HandleOAuthCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusBadRequest)
		gt.S(t, rec.Body.String()).Contains("Invalid state parameter")
	})
}

func TestNotionIntegrationUseCases_IntegrationFlow(t *testing.T) {
	ctx := context.Background()

	// Mock repository
	savedIntegration := (*integration.NotionIntegration)(nil)
	mockUserRepo := &mock.UserRepositoryMock{
		SaveNotionIntegrationFunc: func(ctx context.Context, integration *integration.NotionIntegration) error {
			savedIntegration = integration
			return nil
		},
		GetNotionIntegrationFunc: func(ctx context.Context, userID string) (*integration.NotionIntegration, error) {
			if savedIntegration != nil && savedIntegration.UserID == userID {
				return savedIntegration, nil
			}
			return nil, goerr.New("not found")
		},
		DeleteNotionIntegrationFunc: func(ctx context.Context, userID string) error {
			if savedIntegration != nil && savedIntegration.UserID == userID {
				savedIntegration = nil
				return nil
			}
			return goerr.New("not found")
		},
	}

	oauthConfig := notion.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/oauth/notion/callback",
	}
	oauthService := notion.NewOAuthService(oauthConfig)

	useCases := usecase.NewNotionIntegrationUseCases(mockUserRepo, oauthService)

	testUserID := "test-user-123"

	t.Run("initiate OAuth flow", func(t *testing.T) {
		rec := httptest.NewRecorder()
		oauthURL, err := useCases.InitiateOAuth(ctx, rec, testUserID)
		gt.NoError(t, err)
		gt.S(t, oauthURL).Contains("https://api.notion.com/v1/oauth/authorize")
		gt.S(t, oauthURL).Contains("client_id=test-client-id")
		gt.S(t, oauthURL).Contains("owner=user") // Notion-specific parameter
	})

	t.Run("simulate successful integration save", func(t *testing.T) {
		// Simulate saving an integration (normally done by OAuth callback)
		testIntegration := &integration.NotionIntegration{
			UserID:        testUserID,
			WorkspaceID:   "test-workspace-id",
			WorkspaceName: "Test Workspace",
			WorkspaceIcon: "https://example.com/icon.png",
			BotID:         "test-bot-id",
			AccessToken:   "test-access-token",
			// No RefreshToken or TokenExpiresAt for Notion
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := mockUserRepo.SaveNotionIntegration(ctx, testIntegration)
		gt.NoError(t, err)
	})

	t.Run("get integration after connection", func(t *testing.T) {
		integration, err := useCases.GetIntegration(ctx, testUserID)
		gt.NoError(t, err)
		gt.V(t, integration).NotNil()
		gt.Equal(t, integration.UserID, testUserID)
		gt.Equal(t, integration.WorkspaceName, "Test Workspace")
		gt.Equal(t, integration.WorkspaceID, "test-workspace-id")
	})

	t.Run("disconnect integration", func(t *testing.T) {
		err := useCases.Disconnect(ctx, testUserID)
		gt.NoError(t, err)

		// Verify integration was deleted
		_, err = useCases.GetIntegration(ctx, testUserID)
		gt.Error(t, err)
	})
}
