package config

import (
	"fmt"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/service/notion"
	"github.com/urfave/cli/v3"
)

type Notion struct {
	ClientID     string
	ClientSecret string
}

// Flags returns CLI flags for Notion configuration
func (n *Notion) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "notion-client-id",
			Sources:     cli.EnvVars("TAMAMO_NOTION_CLIENT_ID"),
			Usage:       "Notion OAuth Client ID",
			Destination: &n.ClientID,
		},
		&cli.StringFlag{
			Name:        "notion-client-secret",
			Sources:     cli.EnvVars("TAMAMO_NOTION_CLIENT_SECRET"),
			Usage:       "Notion OAuth Client Secret",
			Destination: &n.ClientSecret,
		},
	}
}

// Validate validates the Notion configuration
func (n *Notion) Validate() error {
	if n.ClientID == "" {
		return goerr.New("Notion Client ID is required")
	}
	if n.ClientSecret == "" {
		return goerr.New("Notion Client Secret is required")
	}

	return nil
}

// IsEnabled returns true if Notion integration is configured
func (n *Notion) IsEnabled() bool {
	return n.ClientID != "" && n.ClientSecret != ""
}

// BuildOAuthConfig creates a notion.OAuthConfig from the configuration
func (n *Notion) BuildOAuthConfig(frontendURL string) notion.OAuthConfig {
	redirectURI := fmt.Sprintf("%s/api/oauth/notion/callback", frontendURL)

	return notion.OAuthConfig{
		ClientID:     n.ClientID,
		ClientSecret: n.ClientSecret,
		RedirectURI:  redirectURI,
	}
}
