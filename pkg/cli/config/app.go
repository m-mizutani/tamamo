package config

import (
	"net/url"

	"github.com/m-mizutani/goerr/v2"
	"github.com/urfave/cli/v3"
)

// App contains general application configuration settings that are used across multiple components
type App struct {
	FrontendURL string // Application frontend URL used for OAuth redirects and other integrations
}

// Flags returns CLI flags for general application configuration
func (a *App) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "frontend-url",
			Sources:     cli.EnvVars("TAMAMO_FRONTEND_URL"),
			Usage:       "Application frontend URL (e.g., https://app.example.com)",
			Value:       "http://localhost:8080",
			Destination: &a.FrontendURL,
		},
	}
}

// Validate validates the application configuration
func (a *App) Validate() error {
	if a.FrontendURL == "" {
		return goerr.New("Frontend URL is required")
	}

	// Validate frontend URL format
	_, err := url.Parse(a.FrontendURL)
	if err != nil {
		return goerr.Wrap(err, "invalid frontend URL format")
	}

	return nil
}