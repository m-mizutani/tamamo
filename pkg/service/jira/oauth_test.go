package jira_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/service/jira"
)

func TestOAuthService_GenerateOAuthURL(t *testing.T) {
	config := jira.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/jira/callback",
	}
	oauthService := jira.NewOAuthService(config)

	t.Run("generate valid OAuth URL", func(t *testing.T) {
		oauthURL, state, err := oauthService.GenerateOAuthURL()
		gt.NoError(t, err)

		// Parse the URL to verify parameters
		parsedURL, err := url.Parse(oauthURL)
		gt.NoError(t, err)

		gt.Equal(t, parsedURL.Scheme, "https")
		gt.Equal(t, parsedURL.Host, "auth.atlassian.com")
		gt.Equal(t, parsedURL.Path, "/authorize")

		// Check query parameters
		params := parsedURL.Query()
		gt.Equal(t, params.Get("audience"), "api.atlassian.com")
		gt.Equal(t, params.Get("client_id"), "test-client-id")
		scope := params.Get("scope")
		gt.S(t, scope).Contains("read:jira-user")
		gt.S(t, scope).Contains("read:jira-work")
		gt.Equal(t, params.Get("redirect_uri"), "http://localhost:8080/api/auth/jira/callback")
		gt.Equal(t, params.Get("state"), state)
		gt.Equal(t, params.Get("response_type"), "code")
		gt.Equal(t, params.Get("prompt"), "consent")

		// State should not be empty
		gt.V(t, state).NotEqual("")
	})
}

func TestOAuthService_StateCookieOperations(t *testing.T) {
	config := jira.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/jira/callback",
	}
	oauthService := jira.NewOAuthService(config)

	t.Run("set and get state cookie", func(t *testing.T) {
		// Set cookie
		rec := httptest.NewRecorder()
		state := "test-state-789"
		userID := "user-123"

		err := oauthService.SetOAuthStateCookie(rec, state, userID)
		gt.NoError(t, err)

		cookies := rec.Result().Cookies()
		gt.V(t, len(cookies)).Equal(1)
		cookie := cookies[0]

		// Get cookie
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(cookie)

		retrievedState, retrievedUserID, err := oauthService.GetOAuthStateFromCookie(req)
		gt.NoError(t, err)
		gt.Equal(t, retrievedState, state)
		gt.Equal(t, retrievedUserID, userID)
	})

	t.Run("clear state cookie", func(t *testing.T) {
		rec := httptest.NewRecorder()
		oauthService.ClearOAuthStateCookie(rec)

		cookies := rec.Result().Cookies()
		gt.V(t, len(cookies)).Equal(1)

		cookie := cookies[0]
		gt.Equal(t, cookie.Name, "jira_oauth_state")
		gt.Equal(t, cookie.MaxAge, -1)
		gt.Equal(t, cookie.Value, "")
	})

	t.Run("handle missing cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		_, _, err := oauthService.GetOAuthStateFromCookie(req)
		gt.Error(t, err)
	})

	t.Run("handle invalid cookie value", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "jira_oauth_state",
			Value: "invalid-jwt-token",
		})

		_, _, err := oauthService.GetOAuthStateFromCookie(req)
		gt.Error(t, err)
	})
}

func TestOAuthService_GenerateMultipleStates(t *testing.T) {
	config := jira.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/jira/callback",
	}
	oauthService := jira.NewOAuthService(config)

	t.Run("generate unique states", func(t *testing.T) {
		_, state1, err1 := oauthService.GenerateOAuthURL()
		gt.NoError(t, err1)

		_, state2, err2 := oauthService.GenerateOAuthURL()
		gt.NoError(t, err2)

		// States should be different
		gt.V(t, state1).NotEqual(state2)

		// States should be non-empty hex strings
		gt.V(t, len(state1) > 10).Equal(true)
		gt.V(t, len(state2) > 10).Equal(true)

		// Should be valid hex (check first few characters)
		for _, char := range state1[:10] {
			isValidHex := strings.ContainsRune("0123456789abcdef", char)
			gt.V(t, isValidHex).Equal(true)
		}
	})
}
