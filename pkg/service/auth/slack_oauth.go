package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
)

const (
	slackOAuthAuthorizeURL = "https://slack.com/oauth/v2/authorize"
	slackOAuthAccessURL    = "https://slack.com/api/oauth.v2.access"
	slackUserInfoURL       = "https://slack.com/api/users.identity"
)

// SlackOAuthService handles Slack OAuth operations
type SlackOAuthService struct {
	clientID     string
	clientSecret string
	redirectURI  string
	teamID       string // Optional: specify team for OAuth
	httpClient   *http.Client
}

// NewSlackOAuthService creates a new Slack OAuth service
func NewSlackOAuthService(clientID, clientSecret, redirectURI string) *SlackOAuthService {
	return &SlackOAuthService{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient:   &http.Client{},
	}
}

// NewSlackOAuthServiceWithTeam creates a new Slack OAuth service with team specification
func NewSlackOAuthServiceWithTeam(clientID, clientSecret, redirectURI, teamID string) *SlackOAuthService {
	return &SlackOAuthService{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		teamID:       teamID,
		httpClient:   &http.Client{},
	}
}

// GenerateAuthURL generates the Slack OAuth authorization URL
func (s *SlackOAuthService) GenerateAuthURL(state string) string {
	params := url.Values{
		"client_id":    {s.clientID},
		"redirect_uri": {s.redirectURI},
		"state":        {state},
		"user_scope":   {"identity.basic,identity.email,identity.team"},
	}

	// Add team parameter if specified
	if s.teamID != "" {
		params.Set("team", s.teamID)
	}

	return fmt.Sprintf("%s?%s", slackOAuthAuthorizeURL, params.Encode())
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func (s *SlackOAuthService) ExchangeCodeForToken(ctx context.Context, code string) (*SlackOAuthTokenResponse, error) {
	formData := url.Values{
		"client_id":     {s.clientID},
		"client_secret": {s.clientSecret},
		"code":          {code},
		"redirect_uri":  {s.redirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, slackOAuthAccessURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to exchange code for token")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, goerr.New("token exchange failed",
			goerr.V("status", resp.StatusCode),
			goerr.V("body", string(body)))
	}

	var tokenResp SlackOAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal token response")
	}

	if !tokenResp.OK {
		return nil, goerr.Wrap(auth.ErrOAuthFailed, tokenResp.Error)
	}

	return &tokenResp, nil
}

// GetUserInfo gets the user information using the access token
func (s *SlackOAuthService) GetUserInfo(ctx context.Context, accessToken string) (*SlackUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, slackUserInfoURL, nil)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get user info")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, goerr.New("failed to get user info",
			goerr.V("status", resp.StatusCode),
			goerr.V("body", string(body)))
	}

	var userResp SlackUserIdentityResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal user info response")
	}

	if !userResp.OK {
		return nil, goerr.New("failed to get user info", goerr.V("error", userResp.Error))
	}

	return &userResp.User, nil
}

// SlackOAuthTokenResponse represents the Slack OAuth token response
type SlackOAuthTokenResponse struct {
	OK          bool   `json:"ok"`
	Error       string `json:"error,omitempty"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	BotUserID   string `json:"bot_user_id,omitempty"`
	AppID       string `json:"app_id"`
	Team        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	Enterprise struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"enterprise,omitempty"`
	AuthedUser struct {
		ID          string `json:"id"`
		Scope       string `json:"scope"`
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	} `json:"authed_user"`
}

// SlackUserIdentityResponse represents the Slack user identity response
type SlackUserIdentityResponse struct {
	OK    bool          `json:"ok"`
	Error string        `json:"error,omitempty"`
	User  SlackUserInfo `json:"user"`
	Team  struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
}

// SlackUserInfo represents Slack user information
type SlackUserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
