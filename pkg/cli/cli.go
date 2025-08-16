package cli

import (
	"context"
	"log/slog"

	"github.com/m-mizutani/tamamo/pkg/cli/config"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, args []string) error {
	var loggerCfg config.Logger
	app := &cli.Command{
		Name:  "tamamo",
		Usage: "Slack bot application",
		Flags: loggerCfg.Flags(),
		Before: func(ctx context.Context, c *cli.Command) error {
			if err := loggerCfg.Configure(); err != nil {
				return err
			}

			slog.Info("base options", "logger", loggerCfg)
			return nil
		},
		Commands: []*cli.Command{
			cmdServe(),
		},
	}

	if err := app.Run(ctx, args); err != nil {
		slog.Error("failed to run app", "error", err)
		return err
	}

	return nil
}
