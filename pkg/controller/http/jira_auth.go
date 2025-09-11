package http

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/m-mizutani/tamamo/pkg/service/jira"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

type JiraAuthController struct {
	jiraUseCases usecase.JiraIntegrationUseCases
	oauthService *jira.OAuthService
}

func NewJiraAuthController(
	jiraUseCases usecase.JiraIntegrationUseCases,
	oauthService *jira.OAuthService,
) *JiraAuthController {
	return &JiraAuthController{
		jiraUseCases: jiraUseCases,
		oauthService: oauthService,
	}
}

// HandleOAuthCallback handles the OAuth callback from Jira
func (c *JiraAuthController) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Check for OAuth errors
	if errorParam != "" {
		errorDescription := r.URL.Query().Get("error_description")
		errorMsg := fmt.Sprintf("OAuth Error: %s - %s", errorParam, errorDescription)
		redirectURL := fmt.Sprintf("/integrations/jira/error?message=%s", url.QueryEscape(errorMsg))
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
		redirectURL := fmt.Sprintf("/integrations/jira/error?message=%s", url.QueryEscape(errorMsg))
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// Get accessible Jira resources
	resources, err := c.oauthService.GetAccessibleResources(tokenResponse.AccessToken)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get accessible resources: %v", err)
		redirectURL := fmt.Sprintf("/integrations/jira/error?message=%s", url.QueryEscape(errorMsg))
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// For now, use the first available resource
	// In a more sophisticated implementation, we might let users choose
	if len(resources) == 0 {
		errorMsg := "No accessible Jira sites found"
		redirectURL := fmt.Sprintf("/integrations/jira/error?message=%s", url.QueryEscape(errorMsg))
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	resource := resources[0]

	// Save the integration through the use case
	err = c.jiraUseCases.SaveIntegration(r.Context(), userID, resource.ID, resource.URL,
		tokenResponse.AccessToken, tokenResponse.RefreshToken, tokenResponse.ExpiresIn)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to save integration: %v", err)
		redirectURL := fmt.Sprintf("/integrations/jira/error?message=%s", url.QueryEscape(errorMsg))
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// Redirect to frontend success page
	http.Redirect(w, r, "/integrations/jira/success", http.StatusSeeOther)
}
