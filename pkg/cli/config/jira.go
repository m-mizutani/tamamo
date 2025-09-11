package config

import (
	"fmt"
	"net/url"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/service/jira"
	"github.com/urfave/cli/v3"
)

type Jira struct {
	ClientID     string
	ClientSecret string
	FrontendURL  string // Application frontend URL for generating redirect URI
}

// Flags returns CLI flags for Jira configuration
func (j *Jira) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "jira-client-id",
			Sources:     cli.EnvVars("TAMAMO_JIRA_CLIENT_ID"),
			Usage:       "Jira OAuth Client ID",
			Destination: &j.ClientID,
		},
		&cli.StringFlag{
			Name:        "jira-client-secret",
			Sources:     cli.EnvVars("TAMAMO_JIRA_CLIENT_SECRET"),
			Usage:       "Jira OAuth Client Secret",
			Destination: &j.ClientSecret,
		},
		&cli.StringFlag{
			Name:        "frontend-url",
			Sources:     cli.EnvVars("TAMAMO_FRONTEND_URL"),
			Usage:       "Application frontend URL (e.g., https://app.example.com)",
			Value:       "http://localhost:8080",
			Destination: &j.FrontendURL,
		},
	}
}

// Validate validates the Jira configuration
func (j *Jira) Validate() error {
	if j.ClientID == "" {
		return goerr.New("Jira Client ID is required")
	}
	if j.ClientSecret == "" {
		return goerr.New("Jira Client Secret is required")
	}
	if j.FrontendURL == "" {
		return goerr.New("Frontend URL is required")
	}

	// Validate frontend URL format
	_, err := url.Parse(j.FrontendURL)
	if err != nil {
		return goerr.Wrap(err, "invalid frontend URL format")
	}

	return nil
}

// IsEnabled returns true if Jira integration is configured
func (j *Jira) IsEnabled() bool {
	return j.ClientID != "" && j.ClientSecret != ""
}

// BuildOAuthConfig creates a jira.OAuthConfig from the configuration
func (j *Jira) BuildOAuthConfig() jira.OAuthConfig {
	redirectURI := fmt.Sprintf("%s/api/auth/jira/callback", j.FrontendURL)

	return jira.OAuthConfig{
		ClientID:     j.ClientID,
		ClientSecret: j.ClientSecret,
		RedirectURI:  redirectURI,
	}
}
