package usecase

import (
	"context"
	"fmt"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
	authservice "github.com/m-mizutani/tamamo/pkg/service/auth"
)

type authUseCaseImpl struct {
	sessionRepo   interfaces.SessionRepository
	slackOAuth    *authservice.SlackOAuthService
	tokenExchange *authservice.TokenExchangeService
	frontendURL   string
}

// NewAuthUseCase creates a new authentication use case
func NewAuthUseCase(
	sessionRepo interfaces.SessionRepository,
	clientID, clientSecret, frontendURL string,
) interfaces.AuthUseCases {
	// Construct redirect URI from frontend URL
	redirectURI := fmt.Sprintf("%s/api/auth/callback", frontendURL)

	slackOAuth := authservice.NewSlackOAuthService(clientID, clientSecret, redirectURI)
	tokenExchange := authservice.NewTokenExchangeService(slackOAuth)

	return &authUseCaseImpl{
		sessionRepo:   sessionRepo,
		slackOAuth:    slackOAuth,
		tokenExchange: tokenExchange,
		frontendURL:   frontendURL,
	}
}

// NewAuthUseCaseWithSlackOAuth creates a new authentication use case with a pre-configured SlackOAuthService
func NewAuthUseCaseWithSlackOAuth(
	sessionRepo interfaces.SessionRepository,
	slackOAuth *authservice.SlackOAuthService,
	frontendURL string,
) interfaces.AuthUseCases {
	tokenExchange := authservice.NewTokenExchangeService(slackOAuth)

	return &authUseCaseImpl{
		sessionRepo:   sessionRepo,
		slackOAuth:    slackOAuth,
		tokenExchange: tokenExchange,
		frontendURL:   frontendURL,
	}
}

// GenerateLoginURL generates a Slack OAuth login URL with the provided state
func (u *authUseCaseImpl) GenerateLoginURL(ctx context.Context, state string) (string, error) {
	// Generate authorization URL with provided state
	authURL := u.slackOAuth.GenerateAuthURL(state)
	return authURL, nil
}

// HandleCallback handles the OAuth callback and creates a session
func (u *authUseCaseImpl) HandleCallback(ctx context.Context, code string) (*auth.Session, error) {
	// Exchange code for session (state validation is now done in controller)
	session, err := u.tokenExchange.ExchangeCodeForSession(ctx, code)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to exchange code for session")
	}

	// Save session to repository
	if err := u.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, goerr.Wrap(err, "failed to create session")
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (u *authUseCaseImpl) GetSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	session, err := u.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		if err == auth.ErrSessionNotFound {
			return nil, auth.ErrSessionNotFound
		}
		if err == auth.ErrSessionExpired {
			return nil, auth.ErrSessionExpired
		}
		return nil, goerr.Wrap(err, "failed to get session")
	}

	// Double-check session validity
	if !session.IsValid() {
		return nil, auth.ErrInvalidSession
	}

	return session, nil
}

// Logout deletes a session
func (u *authUseCaseImpl) Logout(ctx context.Context, sessionID string) error {
	if err := u.sessionRepo.DeleteSession(ctx, sessionID); err != nil {
		if err == auth.ErrSessionNotFound {
			// Session doesn't exist, consider it a successful logout
			return nil
		}
		return goerr.Wrap(err, "failed to delete session")
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (u *authUseCaseImpl) CleanupExpiredSessions(ctx context.Context) error {
	if err := u.sessionRepo.CleanupExpiredSessions(ctx); err != nil {
		return goerr.Wrap(err, "failed to cleanup expired sessions")
	}

	return nil
}
