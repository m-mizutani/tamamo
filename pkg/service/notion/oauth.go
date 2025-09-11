package notion

import (
	"bytes"
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
	ClientID            string
	ClientSecret        string
	RedirectURI         string
	AllowedWorkspaceIDs []string
}

type OAuthService struct {
	config OAuthConfig
}

// TokenResponse represents the response from Notion's OAuth token endpoint
// Note: Notion doesn't provide refresh tokens and tokens don't expire
type TokenResponse struct {
	AccessToken   string `json:"access_token"`
	TokenType     string `json:"token_type"`
	BotID         string `json:"bot_id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	WorkspaceIcon string `json:"workspace_icon"`
	Owner         struct {
		Type string `json:"type"`
		User struct {
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"user"`
	} `json:"owner"`
	DuplicatedTemplateID string `json:"duplicated_template_id,omitempty"`
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

	baseURL := "https://api.notion.com/v1/oauth/authorize"
	params := url.Values{
		"owner":         {"user"}, // Required by Notion API
		"client_id":     {s.config.ClientID},
		"redirect_uri":  {s.config.RedirectURI},
		"response_type": {"code"},
		"state":         {state},
	}

	authURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	return authURL, state, nil
}

// SetOAuthStateCookie sets the OAuth state in an HTTP-only cookie
func (s *OAuthService) SetOAuthStateCookie(w http.ResponseWriter, state, userID string) error {
	// Create JWT token with state and user ID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"state":  state,
		"userID": userID,
		"exp":    time.Now().Add(5 * time.Minute).Unix(),
	})

	// Sign with a secret (using client secret as the signing key)
	tokenString, err := token.SignedString([]byte(s.config.ClientSecret))
	if err != nil {
		return goerr.Wrap(err, "failed to sign JWT token")
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "notion_oauth_state",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes
	})

	return nil
}

// GetOAuthStateFromCookie retrieves and validates the OAuth state from cookie
func (s *OAuthService) GetOAuthStateFromCookie(r *http.Request) (string, string, error) {
	cookie, err := r.Cookie("notion_oauth_state")
	if err != nil {
		return "", "", goerr.Wrap(err, "state cookie not found")
	}

	// Parse JWT token
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, goerr.New("unexpected signing method", goerr.V("method", token.Header["alg"]))
		}
		return []byte(s.config.ClientSecret), nil
	})

	if err != nil {
		return "", "", goerr.Wrap(err, "failed to parse JWT token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", goerr.New("invalid token claims")
	}

	state, ok := claims["state"].(string)
	if !ok {
		return "", "", goerr.New("state not found in token")
	}

	userID, ok := claims["userID"].(string)
	if !ok {
		return "", "", goerr.New("userID not found in token")
	}

	return state, userID, nil
}

// ClearOAuthStateCookie clears the OAuth state cookie
func (s *OAuthService) ClearOAuthStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "notion_oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Delete cookie
	})
}

// IsWorkspaceAllowed checks if the workspace ID is in the allowed list
func (s *OAuthService) IsWorkspaceAllowed(workspaceID string) bool {
	// If no workspace restrictions are configured, allow all workspaces
	if len(s.config.AllowedWorkspaceIDs) == 0 {
		return true
	}

	// Check if the workspace ID is in the allowed list
	for _, allowedID := range s.config.AllowedWorkspaceIDs {
		if allowedID == workspaceID {
			return true
		}
	}

	return false
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func (s *OAuthService) ExchangeCodeForToken(code string) (*TokenResponse, error) {
	tokenURL := "https://api.notion.com/v1/oauth/token"

	// Prepare request body
	reqBody := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": s.config.RedirectURI,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to marshal request body")
	}

	// Create request
	req, err := http.NewRequest("POST", tokenURL, io.NopCloser(io.Reader(bytes.NewReader(jsonBody))))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set basic auth for client credentials
	req.SetBasicAuth(s.config.ClientID, s.config.ClientSecret)

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to exchange code for token")
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, goerr.New("token exchange failed",
			goerr.V("status", resp.StatusCode),
			goerr.V("response", string(body)))
	}

	// Parse response
	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, goerr.Wrap(err, "failed to decode token response")
	}

	return &tokenResponse, nil
}
