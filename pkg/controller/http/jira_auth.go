package http

import (
	"fmt"
	"html/template"
	"net/http"

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
		c.renderErrorPage(w, fmt.Sprintf("OAuth Error: %s - %s", errorParam, errorDescription))
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
		c.renderErrorPage(w, fmt.Sprintf("Failed to exchange authorization code: %v", err))
		return
	}

	// Get accessible Jira resources
	resources, err := c.oauthService.GetAccessibleResources(tokenResponse.AccessToken)
	if err != nil {
		c.renderErrorPage(w, fmt.Sprintf("Failed to get accessible resources: %v", err))
		return
	}

	// For now, use the first available resource
	// In a more sophisticated implementation, we might let users choose
	if len(resources) == 0 {
		c.renderErrorPage(w, "No accessible Jira sites found")
		return
	}

	resource := resources[0]

	// Save the integration through the use case
	err = c.jiraUseCases.SaveIntegration(r.Context(), userID, resource.ID, resource.URL,
		tokenResponse.AccessToken, tokenResponse.RefreshToken, tokenResponse.ExpiresIn)
	if err != nil {
		c.renderErrorPage(w, fmt.Sprintf("Failed to save integration: %v", err))
		return
	}

	// Render success page
	c.renderSuccessPage(w)
}

// renderSuccessPage renders the OAuth success page
func (c *JiraAuthController) renderSuccessPage(w http.ResponseWriter) {
	successPageHTML := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Jira Integration - Success</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            text-align: center;
            max-width: 400px;
            width: 90%;
        }
        .success-icon {
            font-size: 4rem;
            color: #22c55e;
            margin-bottom: 1rem;
        }
        h1 {
            color: #1f2937;
            margin: 0 0 1rem;
            font-size: 1.8rem;
            font-weight: 600;
        }
        p {
            color: #6b7280;
            margin: 0 0 2rem;
            line-height: 1.6;
        }
        .close-button {
            background: #3b82f6;
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        .close-button:hover {
            background: #2563eb;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>Successfully connected to Jira!</h1>
        <p>Your Jira integration has been set up successfully. You can now close this window and return to the application.</p>
        <button class="close-button" onclick="window.close()">Close Window</button>
    </div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(successPageHTML)); err != nil {
		// Log error but don't send another HTTP error as headers are already sent
		fmt.Printf("Error writing success page: %v\n", err)
	}
}

// renderErrorPage renders an OAuth error page
func (c *JiraAuthController) renderErrorPage(w http.ResponseWriter, errorMessage string) {
	errorPageTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Jira Integration - Error</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
            background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%);
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            text-align: center;
            max-width: 400px;
            width: 90%;
        }
        .error-icon {
            font-size: 4rem;
            color: #ef4444;
            margin-bottom: 1rem;
        }
        h1 {
            color: #1f2937;
            margin: 0 0 1rem;
            font-size: 1.8rem;
            font-weight: 600;
        }
        .error-message {
            background: #fef2f2;
            color: #b91c1c;
            padding: 1rem;
            border-radius: 8px;
            margin-bottom: 2rem;
            font-size: 0.9rem;
            border: 1px solid #fecaca;
        }
        .close-button {
            background: #6b7280;
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        .close-button:hover {
            background: #4b5563;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-icon">⚠</div>
        <h1>Authorization failed</h1>
        <div class="error-message">{{.ErrorMessage}}</div>
        <button class="close-button" onclick="window.close()">Close Window</button>
    </div>
</body>
</html>
`

	tmpl, err := template.New("error").Parse(errorPageTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)

	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: template.HTMLEscapeString(errorMessage),
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template execution failed", http.StatusInternalServerError)
	}
}
