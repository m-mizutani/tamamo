package cli

import (
	"context"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/cli/config"
	"github.com/m-mizutani/tamamo/pkg/utils/errors"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, args []string) error {
	var loggerCfg config.Logger
	app := &cli.Command{
		Name:  "tamamo",
		Usage: "Slack bot application",
		Flags: loggerCfg.Flags(),
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			logger, err := loggerCfg.Configure()
			if err != nil {
				return ctx, err
			}

			ctx = ctxlog.With(ctx, logger)
			ctxlog.From(ctx).Info("base options", "logger", loggerCfg)
			return ctx, nil
		},
		Commands: []*cli.Command{
			cmdServe(),
		},
	}

	if err := app.Run(ctx, args); err != nil {
		errors.Handle(ctx, goerr.Wrap(err, "failed to run app"))
		return err
	}

	return nil
}
