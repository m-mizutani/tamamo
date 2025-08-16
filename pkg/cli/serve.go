package cli

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/cli/config"
	server "github.com/m-mizutani/tamamo/pkg/controller/http"
	slack_controller "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/urfave/cli/v3"
)

func cmdServe() *cli.Command {
	var (
		addr     string
		slackCfg config.Slack
	)

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "addr",
			Aliases:     []string{"a"},
			Sources:     cli.EnvVars("TAMAMO_ADDR"),
			Usage:       "Listen address (default: 127.0.0.1:8080)",
			Value:       "127.0.0.1:8080",
			Destination: &addr,
		},
	}
	flags = append(flags, slackCfg.Flags()...)

	return &cli.Command{
		Name:    "serve",
		Aliases: []string{"s"},
		Usage:   "Run server",
		Flags:   flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger := ctxlog.From(ctx)
			logger.Info("starting server",
				"addr", addr,
				"slack", slackCfg,
			)

			// Configure Slack service
			slackSvc, err := slackCfg.Configure()
			if err != nil {
				return goerr.Wrap(err, "failed to configure slack service")
			}

			// Create usecase
			uc := usecase.New(
				usecase.WithSlackClient(slackSvc),
			)

			// Create controllers
			slackCtrl := slack_controller.New(uc)

			// Build HTTP server options
			serverOptions := []server.Options{
				server.WithSlackController(slackCtrl),
				server.WithSlackVerifier(slackCfg.Verifier()),
			}

			httpServer := http.Server{
				Addr:              addr,
				Handler:           server.New(serverOptions...),
				ReadTimeout:       30 * time.Second,
				ReadHeaderTimeout: 10 * time.Second,
				BaseContext: func(l net.Listener) context.Context {
					return ctx
				},
			}

			errCh := make(chan error, 1)
			go func() {
				defer close(errCh)
				ctxlog.From(ctx).Info("server started", "addr", addr)
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errCh <- err
				}
			}()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			select {
			case err := <-errCh:
				return err
			case <-sigCh:
				ctxlog.From(ctx).Info("shutting down server...")
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				return httpServer.Shutdown(shutdownCtx)
			}
		},
	}
}
