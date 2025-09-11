package config

import (
	"fmt"
	"net/url"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/service/notion"
	"github.com/urfave/cli/v3"
)

type Notion struct {
	ClientID            string
	ClientSecret        string
	FrontendURL         string   // Application frontend URL for generating redirect URI
	AllowedWorkspaceIDs []string // List of allowed Notion workspace IDs
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
		&cli.StringSliceFlag{
			Name:        "notion-allowed-workspaces",
			Sources:     cli.EnvVars("TAMAMO_NOTION_ALLOWED_WORKSPACES"),
			Usage:       "Comma-separated list of allowed Notion workspace IDs",
			Destination: &n.AllowedWorkspaceIDs,
		},
		// FrontendURL is shared with Jira config, so we don't define it here
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
	if n.FrontendURL == "" {
		return goerr.New("Frontend URL is required")
	}

	// Validate frontend URL format
	_, err := url.Parse(n.FrontendURL)
	if err != nil {
		return goerr.Wrap(err, "invalid frontend URL format")
	}

	return nil
}

// IsEnabled returns true if Notion integration is configured
func (n *Notion) IsEnabled() bool {
	return n.ClientID != "" && n.ClientSecret != ""
}

// IsWorkspaceAllowed checks if a workspace ID is in the allowed list
func (n *Notion) IsWorkspaceAllowed(workspaceID string) bool {
	// If no workspace restrictions are configured, allow all workspaces
	if len(n.AllowedWorkspaceIDs) == 0 {
		return true
	}

	// Check if the workspace ID is in the allowed list
	for _, allowedID := range n.AllowedWorkspaceIDs {
		if allowedID == workspaceID {
			return true
		}
	}

	return false
}

// BuildOAuthConfig creates a notion.OAuthConfig from the configuration
func (n *Notion) BuildOAuthConfig() notion.OAuthConfig {
	redirectURI := fmt.Sprintf("%s/api/auth/notion/callback", n.FrontendURL)

	return notion.OAuthConfig{
		ClientID:            n.ClientID,
		ClientSecret:        n.ClientSecret,
		RedirectURI:         redirectURI,
		AllowedWorkspaceIDs: n.AllowedWorkspaceIDs,
	}
}
