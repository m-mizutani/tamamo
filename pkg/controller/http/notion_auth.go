package http

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/m-mizutani/tamamo/pkg/service/notion"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

type NotionAuthController struct {
	notionUseCases usecase.NotionIntegrationUseCases
	oauthService   *notion.OAuthService
}

func NewNotionAuthController(
	notionUseCases usecase.NotionIntegrationUseCases,
	oauthService *notion.OAuthService,
) *NotionAuthController {
	return &NotionAuthController{
		notionUseCases: notionUseCases,
		oauthService:   oauthService,
	}
}

// HandleOAuthCallback handles the OAuth callback from Notion
func (c *NotionAuthController) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Check for OAuth errors
	if errorParam != "" {
		errorDescription := r.URL.Query().Get("error_description")
		errorMsg := fmt.Sprintf("OAuth Error: %s - %s", errorParam, errorDescription)
		redirectURL := fmt.Sprintf("/integrations/notion/error?message=%s", url.QueryEscape(errorMsg))
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// Validate required parameters
	if code == "" || state == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Validate state parameter and get user ID from cookie
	cookieState, userID, err := c.oauthService.GetOAuthStateFromCookie(r)
	if err != nil {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	if cookieState != state {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	// Clear the state cookie
	c.oauthService.ClearOAuthStateCookie(w)

	// Exchange authorization code for tokens
	tokenResponse, err := c.oauthService.ExchangeCodeForToken(code)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to exchange authorization code: %v", err)
		redirectURL := fmt.Sprintf("/integrations/notion/error?message=%s", url.QueryEscape(errorMsg))
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// Validate workspace access
	if !c.oauthService.IsWorkspaceAllowed(tokenResponse.WorkspaceID) {
		errorMsg := fmt.Sprintf("Access denied: This Notion workspace (%s) is not allowed to connect to Tamamo.", tokenResponse.WorkspaceName)
		redirectURL := fmt.Sprintf("/integrations/notion/error?message=%s", url.QueryEscape(errorMsg))
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// Save the integration through the use case
	err = c.notionUseCases.SaveIntegration(r.Context(), userID,
		tokenResponse.WorkspaceID,
		tokenResponse.WorkspaceName,
		tokenResponse.WorkspaceIcon,
		tokenResponse.BotID,
		tokenResponse.AccessToken)
	if err != nil {
		// Check if error is related to workspace membership
		if err.Error() == "user is not a member of the Slack workspace" {
			errorMsg := "Access denied: You must be a member of the Slack workspace to connect Notion."
			redirectURL := fmt.Sprintf("/integrations/notion/error?message=%s", url.QueryEscape(errorMsg))
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		errorMsg := fmt.Sprintf("Failed to save integration: %v", err)
		redirectURL := fmt.Sprintf("/integrations/notion/error?message=%s", url.QueryEscape(errorMsg))
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// Redirect to frontend success page
	http.Redirect(w, r, "/integrations/notion/success", http.StatusSeeOther)
}
