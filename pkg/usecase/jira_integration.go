package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
	"github.com/m-mizutani/tamamo/pkg/service/jira"
)

type JiraIntegrationUseCases interface {
	InitiateOAuth(ctx context.Context, w http.ResponseWriter, userID string) (string, error)
	GetIntegration(ctx context.Context, userID string) (*integration.JiraIntegration, error)
	SaveIntegration(ctx context.Context, userID, cloudID, siteURL, accessToken, refreshToken string, expiresIn int) error
	Disconnect(ctx context.Context, userID string) error
}

type jiraIntegrationUseCases struct {
	userRepo     interfaces.UserRepository
	oauthService *jira.OAuthService
}

func NewJiraIntegrationUseCases(
	userRepo interfaces.UserRepository,
	oauthService *jira.OAuthService,
) JiraIntegrationUseCases {
	return &jiraIntegrationUseCases{
		userRepo:     userRepo,
		oauthService: oauthService,
	}
}

// InitiateOAuth starts the OAuth flow and returns the authorization URL
func (uc *jiraIntegrationUseCases) InitiateOAuth(ctx context.Context, w http.ResponseWriter, userID string) (string, error) {
	// Generate OAuth URL with state
	authURL, state, err := uc.oauthService.GenerateOAuthURL()
	if err != nil {
		return "", goerr.Wrap(err, "failed to generate OAuth URL")
	}

	// Set state cookie for CSRF protection
	if err := uc.oauthService.SetOAuthStateCookie(w, state, userID); err != nil {
		return "", goerr.Wrap(err, "failed to set state cookie")
	}

	return authURL, nil
}

// GetIntegration retrieves the current Jira integration status for a user
func (uc *jiraIntegrationUseCases) GetIntegration(ctx context.Context, userID string) (*integration.JiraIntegration, error) {
	jiraIntegration, err := uc.userRepo.GetJiraIntegration(ctx, userID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get Jira integration", goerr.V("user_id", userID))
	}

	// Return nil if no integration exists (not connected)
	if jiraIntegration == nil {
		return nil, nil
	}

	// Refresh token if expired
	if jiraIntegration.IsTokenExpired() && jiraIntegration.RefreshToken != "" {
		tokenResponse, err := uc.oauthService.RefreshAccessToken(jiraIntegration.RefreshToken)
		if err != nil {
			// If refresh fails, log the error but still return the integration
			// The GraphQL resolver will show it as disconnected
			return jiraIntegration, nil
		}

		// Update tokens
		expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
		jiraIntegration.UpdateTokens(tokenResponse.AccessToken, tokenResponse.RefreshToken, expiresAt)

		// Save updated integration
		if err := uc.userRepo.SaveJiraIntegration(ctx, jiraIntegration); err != nil {
			return nil, goerr.Wrap(err, "failed to save refreshed tokens", goerr.V("user_id", userID))
		}
	}

	return jiraIntegration, nil
}

// SaveIntegration saves the Jira integration details after successful OAuth
func (uc *jiraIntegrationUseCases) SaveIntegration(ctx context.Context, userID, cloudID, siteURL, accessToken, refreshToken string, expiresIn int) error {
	// Create new integration
	jiraIntegration := integration.NewJiraIntegration(userID)
	jiraIntegration.UpdateSiteInfo(cloudID, siteURL)

	// Calculate expiration time
	expiresAt := jiraIntegration.CreatedAt.Add(time.Duration(expiresIn) * time.Second)
	jiraIntegration.UpdateTokens(accessToken, refreshToken, expiresAt)

	// Save to repository
	if err := uc.userRepo.SaveJiraIntegration(ctx, jiraIntegration); err != nil {
		return goerr.Wrap(err, "failed to save Jira integration",
			goerr.V("user_id", userID),
			goerr.V("cloud_id", cloudID))
	}

	return nil
}

// Disconnect removes the Jira integration for a user
func (uc *jiraIntegrationUseCases) Disconnect(ctx context.Context, userID string) error {
	// Check if integration exists
	existing, err := uc.userRepo.GetJiraIntegration(ctx, userID)
	if err != nil {
		return goerr.Wrap(err, "failed to check existing integration", goerr.V("user_id", userID))
	}

	if existing == nil {
		return goerr.New("no Jira integration found", goerr.V("user_id", userID))
	}

	// Delete the integration
	if err := uc.userRepo.DeleteJiraIntegration(ctx, userID); err != nil {
		return goerr.Wrap(err, "failed to delete Jira integration", goerr.V("user_id", userID))
	}

	return nil
}
