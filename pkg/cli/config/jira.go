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
	BaseURL      string // Application base URL for generating redirect URI
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
			Name:        "base-url",
			Sources:     cli.EnvVars("TAMAMO_BASE_URL"),
			Usage:       "Application base URL (e.g., https://app.example.com)",
			Value:       "http://localhost:8080",
			Destination: &j.BaseURL,
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
	if j.BaseURL == "" {
		return goerr.New("Base URL is required")
	}

	// Validate base URL format
	_, err := url.Parse(j.BaseURL)
	if err != nil {
		return goerr.Wrap(err, "invalid base URL format")
	}

	return nil
}

// IsEnabled returns true if Jira integration is configured
func (j *Jira) IsEnabled() bool {
	return j.ClientID != "" && j.ClientSecret != ""
}

// BuildOAuthConfig creates a jira.OAuthConfig from the configuration
func (j *Jira) BuildOAuthConfig() jira.OAuthConfig {
	redirectURI := fmt.Sprintf("%s/api/auth/jira/callback", j.BaseURL)

	return jira.OAuthConfig{
		ClientID:     j.ClientID,
		ClientSecret: j.ClientSecret,
		RedirectURI:  redirectURI,
	}
}
