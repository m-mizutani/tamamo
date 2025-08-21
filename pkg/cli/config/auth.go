package config

import (
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	authservice "github.com/m-mizutani/tamamo/pkg/service/auth"
	slackservice "github.com/m-mizutani/tamamo/pkg/service/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/urfave/cli/v3"
)

// Auth holds OAuth authentication configuration
type Auth struct {
	SlackOAuthClientID     string `masq:"secret"`
	SlackOAuthClientSecret string `masq:"secret"`
	FrontendURL            string
	NoAuthentication       bool
}

func (x *Auth) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "slack-oauth-client-id",
			Usage:       "Slack OAuth client ID",
			Sources:     cli.EnvVars("TAMAMO_SLACK_OAUTH_CLIENT_ID"),
			Destination: &x.SlackOAuthClientID,
		},
		&cli.StringFlag{
			Name:        "slack-oauth-client-secret",
			Usage:       "Slack OAuth client secret",
			Sources:     cli.EnvVars("TAMAMO_SLACK_OAUTH_CLIENT_SECRET"),
			Destination: &x.SlackOAuthClientSecret,
		},
		&cli.StringFlag{
			Name:        "frontend-url",
			Usage:       "Frontend URL for OAuth redirect",
			Sources:     cli.EnvVars("TAMAMO_FRONTEND_URL"),
			Destination: &x.FrontendURL,
			Value:       "http://localhost:3000",
		},
		&cli.BoolFlag{
			Name:        "no-authentication",
			Usage:       "Disable authentication (anonymous mode)",
			Sources:     cli.EnvVars("TAMAMO_NO_AUTHENTICATION"),
			Destination: &x.NoAuthentication,
			Value:       false,
		},
	}
}

// Validate checks if the auth configuration is valid
func (x *Auth) Validate() error {
	// If no-authentication is enabled, no validation needed
	if x.NoAuthentication {
		return nil
	}

	// Otherwise, OAuth configuration is required
	if x.SlackOAuthClientID == "" || x.SlackOAuthClientSecret == "" {
		return goerr.New("slack OAuth client ID and secret are required when authentication is enabled. Use --no-authentication to disable")
	}

	if x.FrontendURL == "" {
		return goerr.New("frontend URL is required for OAuth")
	}

	return nil
}

// ConfigureAuthUseCase creates an authentication use case
func (x *Auth) ConfigureAuthUseCase(
	sessionRepo interfaces.SessionRepository,
	oauthStateRepo interfaces.OAuthStateRepository,
	slackService *slackservice.Service,
) (interfaces.AuthUseCases, error) {
	// If no-authentication mode, return nil (will be handled by middleware)
	if x.NoAuthentication {
		return nil, nil
	}

	// Get team ID from Slack service if available
	var teamID string
	if slackService != nil {
		info, err := slackService.GetAuthTestInfo()
		if err == nil && info != nil {
			teamID = info.TeamID
		}
	}

	// Create Slack OAuth service with team if available
	var slackOAuth *authservice.SlackOAuthService
	if teamID != "" {
		slackOAuth = authservice.NewSlackOAuthServiceWithTeam(
			x.SlackOAuthClientID,
			x.SlackOAuthClientSecret,
			x.FrontendURL+"/api/auth/callback",
			teamID,
		)
	} else {
		slackOAuth = authservice.NewSlackOAuthService(
			x.SlackOAuthClientID,
			x.SlackOAuthClientSecret,
			x.FrontendURL+"/api/auth/callback",
		)
	}

	// Create auth use case
	return usecase.NewAuthUseCaseWithSlackOAuth(
		sessionRepo,
		oauthStateRepo,
		slackOAuth,
		x.FrontendURL,
	), nil
}

// IsAuthenticationEnabled returns true if authentication is enabled
func (x *Auth) IsAuthenticationEnabled() bool {
	return !x.NoAuthentication
}
