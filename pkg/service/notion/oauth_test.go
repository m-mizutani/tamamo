package notion_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/service/notion"
)

func TestOAuthService_GenerateOAuthURL(t *testing.T) {
	config := notion.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/notion/callback",
	}
	oauthService := notion.NewOAuthService(config)

	t.Run("generates valid OAuth URL", func(t *testing.T) {
		authURL, state, err := oauthService.GenerateOAuthURL()
		gt.NoError(t, err)

		// Parse URL
		parsedURL, err := url.Parse(authURL)
		gt.NoError(t, err)

		// Check base URL
		gt.V(t, parsedURL.Scheme).Equal("https")
		gt.V(t, parsedURL.Host).Equal("api.notion.com")
		gt.V(t, parsedURL.Path).Equal("/v1/oauth/authorize")

		// Check query parameters
		params := parsedURL.Query()
		gt.V(t, params.Get("owner")).Equal("user") // Notion-specific parameter
		gt.V(t, params.Get("client_id")).Equal("test-client-id")
		gt.V(t, params.Get("redirect_uri")).Equal("http://localhost:8080/api/auth/notion/callback")
		gt.V(t, params.Get("response_type")).Equal("code")
		gt.V(t, params.Get("state")).Equal(state)

		// State should be a hex string
		gt.V(t, len(state)).Equal(64) // 32 bytes = 64 hex characters
	})

	t.Run("generates unique states", func(t *testing.T) {
		_, state1, err1 := oauthService.GenerateOAuthURL()
		gt.NoError(t, err1)

		_, state2, err2 := oauthService.GenerateOAuthURL()
		gt.NoError(t, err2)

		// States should be different
		gt.V(t, state1).NotEqual(state2)
	})
}

func TestOAuthService_StateCookie(t *testing.T) {
	config := notion.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/notion/callback",
	}
	oauthService := notion.NewOAuthService(config)

	t.Run("set and get state cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		state := "test-state-123"
		userID := "test-user-456"

		// Set cookie
		err := oauthService.SetOAuthStateCookie(w, state, userID)
		gt.NoError(t, err)

		// Check cookie was set
		cookies := w.Result().Cookies()
		gt.V(t, len(cookies)).Equal(1)

		cookie := cookies[0]
		gt.V(t, cookie.Name).Equal("notion_oauth_state")
		gt.B(t, cookie.HttpOnly).True()
		gt.B(t, cookie.Secure).True()
		gt.V(t, cookie.SameSite).Equal(http.SameSiteLaxMode)
		gt.V(t, cookie.MaxAge).Equal(300)

		// Create request with cookie
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(cookie)

		// Get state from cookie
		retrievedState, retrievedUserID, err := oauthService.GetOAuthStateFromCookie(req)
		gt.NoError(t, err)
		gt.V(t, retrievedState).Equal(state)
		gt.V(t, retrievedUserID).Equal(userID)
	})

	t.Run("clear state cookie", func(t *testing.T) {
		w := httptest.NewRecorder()

		// Clear cookie
		oauthService.ClearOAuthStateCookie(w)

		// Check cookie was cleared
		cookies := w.Result().Cookies()
		gt.V(t, len(cookies)).Equal(1)

		cookie := cookies[0]
		gt.V(t, cookie.Name).Equal("notion_oauth_state")
		gt.V(t, cookie.Value).Equal("")
		gt.V(t, cookie.MaxAge).Equal(-1)
	})

	t.Run("handle missing cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		_, _, err := oauthService.GetOAuthStateFromCookie(req)
		gt.Error(t, err)
	})

	t.Run("handle invalid cookie value", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "notion_oauth_state",
			Value: "invalid-jwt-token",
		})

		_, _, err := oauthService.GetOAuthStateFromCookie(req)
		gt.Error(t, err)
	})
}

func TestOAuthService_GenerateMultipleStates(t *testing.T) {
	config := notion.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURI:  "http://localhost:8080/api/auth/notion/callback",
	}
	oauthService := notion.NewOAuthService(config)

	t.Run("generate unique states", func(t *testing.T) {
		_, state1, err1 := oauthService.GenerateOAuthURL()
		gt.NoError(t, err1)

		_, state2, err2 := oauthService.GenerateOAuthURL()
		gt.NoError(t, err2)

		// States should be different
		gt.V(t, state1).NotEqual(state2)

		// States should be non-empty hex strings
		gt.B(t, len(state1) > 10).True()
		gt.B(t, len(state2) > 10).True()

		// Should be valid hex (check first few characters)
		for _, char := range state1[:10] {
			isValidHex := strings.ContainsRune("0123456789abcdef", char)
			gt.B(t, isValidHex).True()
		}
	})
}

