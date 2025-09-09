package jira

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/m-mizutani/goerr/v2"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type OAuthService struct {
	config OAuthConfig
}

// TokenResponse represents the response from Jira's OAuth token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// AccessibleResource represents a Jira site the user has access to
type AccessibleResource struct {
	ID     string   `json:"id"`
	URL    string   `json:"url"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

func NewOAuthService(config OAuthConfig) *OAuthService {
	return &OAuthService{
		config: config,
	}
}

// generateState generates a cryptographically secure random state string
func (s *OAuthService) generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", goerr.Wrap(err, "failed to generate random state")
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateOAuthURL generates the OAuth authorization URL with state
func (s *OAuthService) GenerateOAuthURL() (string, string, error) {
	state, err := s.generateState()
	if err != nil {
		return "", "", goerr.Wrap(err, "failed to generate state")
	}

	baseURL := "https://auth.atlassian.com/authorize"
	params := url.Values{
		"audience":      {"api.atlassian.com"},
		"client_id":     {s.config.ClientID},
		"scope":         {"read:jira-work read:jira-user"},
		"redirect_uri":  {s.config.RedirectURI},
		"state":         {state},
		"response_type": {"code"},
		"prompt":        {"consent"},
	}

	authURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	return authURL, state, nil
}

// SetOAuthStateCookie sets the OAuth state in an HTTP-only cookie
func (s *OAuthService) SetOAuthStateCookie(w http.ResponseWriter, state string, userID string) error {
	// Create JWT token with state and userID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"state":   state,
		"user_id": userID,
		"exp":     time.Now().Add(5 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	})

	// Sign the token with client secret
	tokenString, err := token.SignedString([]byte(s.config.ClientSecret))
	if err != nil {
		return goerr.Wrap(err, "failed to sign JWT token")
	}

	// Set HTTP-only cookie
	cookie := &http.Cookie{
		Name:     "jira_oauth_state",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   true, // HTTPS only
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
	return nil
}

// GetOAuthStateFromCookie retrieves and validates the OAuth state from cookie
func (s *OAuthService) GetOAuthStateFromCookie(r *http.Request) (string, string, error) {
	cookie, err := r.Cookie("jira_oauth_state")
	if err != nil {
		return "", "", goerr.Wrap(err, "oauth state cookie not found")
	}

	// Parse and validate JWT token
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, goerr.New("unexpected signing method", goerr.V("method", token.Header["alg"]))
		}
		return []byte(s.config.ClientSecret), nil
	})

	if err != nil {
		return "", "", goerr.Wrap(err, "failed to parse JWT token")
	}

	if !token.Valid {
		return "", "", goerr.New("invalid JWT token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", goerr.New("failed to parse JWT claims")
	}

	state, ok := claims["state"].(string)
	if !ok {
		return "", "", goerr.New("state not found in JWT claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", goerr.New("user_id not found in JWT claims")
	}

	return state, userID, nil
}

// ClearOAuthStateCookie clears the OAuth state cookie
func (s *OAuthService) ClearOAuthStateCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "jira_oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// ValidateState validates that the provided state matches the expected state
func (s *OAuthService) ValidateState(providedState, expectedState string) bool {
	return providedState == expectedState
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func (s *OAuthService) ExchangeCodeForToken(code string) (*TokenResponse, error) {
	tokenURL := "https://auth.atlassian.com/oauth/token"

	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {s.config.ClientID},
		"client_secret": {s.config.ClientSecret},
		"code":          {code},
		"redirect_uri":  {s.config.RedirectURI},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to exchange code for token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, goerr.New("token exchange failed",
			goerr.V("status", resp.StatusCode),
			goerr.V("response", string(body)))
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, goerr.Wrap(err, "failed to decode token response")
	}

	return &tokenResponse, nil
}

// GetAccessibleResources retrieves the list of Jira sites the user has access to
func (s *OAuthService) GetAccessibleResources(accessToken string) ([]AccessibleResource, error) {
	resourcesURL := "https://api.atlassian.com/oauth/token/accessible-resources"

	req, err := http.NewRequest("GET", resourcesURL, nil)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create request")
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get accessible resources")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, goerr.New("failed to get accessible resources",
			goerr.V("status", resp.StatusCode),
			goerr.V("response", string(body)))
	}

	var resources []AccessibleResource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, goerr.Wrap(err, "failed to decode resources response")
	}

	return resources, nil
}
