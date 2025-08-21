package auth

import (
	"context"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
)

// TokenExchangeService handles the OAuth token exchange process
type TokenExchangeService struct {
	slackOAuth *SlackOAuthService
}

// NewTokenExchangeService creates a new token exchange service
func NewTokenExchangeService(slackOAuth *SlackOAuthService) *TokenExchangeService {
	return &TokenExchangeService{
		slackOAuth: slackOAuth,
	}
}

// ExchangeCodeForSession exchanges an OAuth code for a user session
func (s *TokenExchangeService) ExchangeCodeForSession(ctx context.Context, code string) (*auth.Session, error) {
	// Exchange code for token
	tokenResp, err := s.slackOAuth.ExchangeCodeForToken(ctx, code)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to exchange code for token")
	}

	// Get user info using the user access token
	userAccessToken := tokenResp.AuthedUser.AccessToken
	if userAccessToken == "" {
		// Fallback to main access token if user token is not available
		userAccessToken = tokenResp.AccessToken
	}

	userInfo, err := s.slackOAuth.GetUserInfo(ctx, userAccessToken)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get user info")
	}

	// Create session with user information
	session := auth.NewSession(
		ctx,
		tokenResp.AuthedUser.ID,
		userInfo.Name,
		userInfo.Email,
		tokenResp.Team.ID,
		tokenResp.Team.Name,
	)

	return session, nil
}

// ValidateClientConfig checks if the OAuth client is properly configured
func (s *TokenExchangeService) ValidateClientConfig() error {
	if s.slackOAuth == nil {
		return auth.ErrMissingConfig
	}

	if s.slackOAuth.clientID == "" || s.slackOAuth.clientSecret == "" {
		return auth.ErrMissingConfig
	}

	return nil
}
