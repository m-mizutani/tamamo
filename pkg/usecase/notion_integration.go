package usecase

import (
	"context"
	"net/http"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
	"github.com/m-mizutani/tamamo/pkg/service/notion"
)

type NotionIntegrationUseCases interface {
	InitiateOAuth(ctx context.Context, w http.ResponseWriter, userID string) (string, error)
	GetIntegration(ctx context.Context, userID string) (*integration.NotionIntegration, error)
	SaveIntegration(ctx context.Context, userID, workspaceID, workspaceName, workspaceIcon, botID, accessToken string) error
	Disconnect(ctx context.Context, userID string) error
}

type notionIntegrationUseCases struct {
	userRepo     interfaces.UserRepository
	oauthService *notion.OAuthService
}

func NewNotionIntegrationUseCases(
	userRepo interfaces.UserRepository,
	oauthService *notion.OAuthService,
) NotionIntegrationUseCases {
	return &notionIntegrationUseCases{
		userRepo:     userRepo,
		oauthService: oauthService,
	}
}

// InitiateOAuth starts the OAuth flow and returns the authorization URL
func (uc *notionIntegrationUseCases) InitiateOAuth(ctx context.Context, w http.ResponseWriter, userID string) (string, error) {
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

// GetIntegration retrieves the current Notion integration status for a user
func (uc *notionIntegrationUseCases) GetIntegration(ctx context.Context, userID string) (*integration.NotionIntegration, error) {
	notionIntegration, err := uc.userRepo.GetNotionIntegration(ctx, userID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get Notion integration", goerr.V("user_id", userID))
	}

	// Return nil if no integration exists (not connected)
	if notionIntegration == nil {
		return nil, nil
	}

	// Notion tokens don't expire, so we just return the integration
	return notionIntegration, nil
}

// SaveIntegration saves the Notion integration details after successful OAuth
func (uc *notionIntegrationUseCases) SaveIntegration(ctx context.Context, userID, workspaceID, workspaceName, workspaceIcon, botID, accessToken string) error {
	// Note: Workspace validation is now performed at the OAuth controller level
	// before this method is called, using the workspace ID from the OAuth response

	// Create new integration
	notionIntegration := integration.NewNotionIntegration(userID)
	notionIntegration.UpdateWorkspaceInfo(workspaceID, workspaceName, workspaceIcon, botID)
	notionIntegration.UpdateTokens(accessToken)

	// Save to repository
	if err := uc.userRepo.SaveNotionIntegration(ctx, notionIntegration); err != nil {
		return goerr.Wrap(err, "failed to save Notion integration",
			goerr.V("user_id", userID),
			goerr.V("workspace_id", workspaceID))
	}

	return nil
}

// Disconnect removes the Notion integration for a user
func (uc *notionIntegrationUseCases) Disconnect(ctx context.Context, userID string) error {
	// Check if integration exists
	existing, err := uc.userRepo.GetNotionIntegration(ctx, userID)
	if err != nil {
		return goerr.Wrap(err, "failed to check existing integration", goerr.V("user_id", userID))
	}

	if existing == nil {
		return goerr.New("no Notion integration found", goerr.V("user_id", userID))
	}

	// Delete the integration
	if err := uc.userRepo.DeleteNotionIntegration(ctx, userID); err != nil {
		return goerr.Wrap(err, "failed to delete Notion integration", goerr.V("user_id", userID))
	}

	return nil
}
