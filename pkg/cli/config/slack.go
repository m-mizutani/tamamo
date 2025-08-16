package config

import (
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	slackSvc "github.com/m-mizutani/tamamo/pkg/service/slack"
	"github.com/urfave/cli/v3"
)

type Slack struct {
	OAuthToken    string `masq:"secret"`
	SigningSecret string `masq:"secret"`
}

func (x *Slack) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "slack-oauth-token",
			Usage:       "Slack OAuth token",
			Sources:     cli.EnvVars("TAMAMO_SLACK_OAUTH_TOKEN"),
			Destination: &x.OAuthToken,
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "slack-signing-secret",
			Usage:       "Slack signing secret for request verification",
			Sources:     cli.EnvVars("TAMAMO_SLACK_SIGNING_SECRET"),
			Destination: &x.SigningSecret,
			Required:    true,
		},
	}
}

func (x *Slack) Configure() (*slackSvc.Service, error) {
	if x.OAuthToken == "" {
		return nil, goerr.New("slack oauth token is required")
	}
	if x.SigningSecret == "" {
		return nil, goerr.New("slack signing secret is required")
	}

	return slackSvc.New(x.OAuthToken)
}

func (x *Slack) Verifier() slack.PayloadVerifier {
	return slack.NewVerifier(x.SigningSecret)
}
