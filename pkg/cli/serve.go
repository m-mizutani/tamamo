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
	"github.com/m-mizutani/gollem/llm/gemini"
	"github.com/m-mizutani/tamamo/pkg/adapters/cs"
	mem_adapter "github.com/m-mizutani/tamamo/pkg/adapters/memory"
	"github.com/m-mizutani/tamamo/pkg/cli/config"
	server "github.com/m-mizutani/tamamo/pkg/controller/http"
	slack_controller "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/repository/database/firestore"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/repository/storage"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/urfave/cli/v3"
)

func cmdServe() *cli.Command {
	var (
		addr           string
		slackCfg       config.Slack
		firestoreCfg   config.Firestore
		geminiProject  string
		geminiModel    string
		geminiLocation string
		storageBucket  string
		storagePrefix  string
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
		&cli.StringFlag{
			Name:        "gemini-project-id",
			Sources:     cli.EnvVars("TAMAMO_GEMINI_PROJECT_ID"),
			Usage:       "Google Cloud Project ID for Gemini API (required)",
			Required:    true,
			Destination: &geminiProject,
		},
		&cli.StringFlag{
			Name:        "gemini-model",
			Sources:     cli.EnvVars("TAMAMO_GEMINI_MODEL"),
			Usage:       "Gemini model to use",
			Value:       "gemini-2.0-flash",
			Destination: &geminiModel,
		},
		&cli.StringFlag{
			Name:        "gemini-location",
			Sources:     cli.EnvVars("TAMAMO_GEMINI_LOCATION"),
			Usage:       "Google Cloud location for Gemini API",
			Value:       "us-central1",
			Destination: &geminiLocation,
		},
		&cli.StringFlag{
			Name:        "storage-bucket",
			Sources:     cli.EnvVars("TAMAMO_STORAGE_BUCKET"),
			Usage:       "Cloud Storage bucket for history storage (if not set, uses memory storage)",
			Destination: &storageBucket,
		},
		&cli.StringFlag{
			Name:        "storage-prefix",
			Sources:     cli.EnvVars("TAMAMO_STORAGE_PREFIX"),
			Usage:       "Prefix for Cloud Storage objects",
			Destination: &storagePrefix,
		},
	}
	flags = append(flags, slackCfg.Flags()...)
	flags = append(flags, firestoreCfg.Flags()...)

	return &cli.Command{
		Name:    "serve",
		Aliases: []string{"s"},
		Usage:   "Run server",
		Flags:   flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger := ctxlog.From(ctx)

			// Validate Gemini configuration
			if geminiProject == "" {
				return goerr.New("gemini-project-id is required")
			}

			// Initialize Gemini client
			logger.Info("initializing Gemini client",
				"project_id", geminiProject,
				"location", geminiLocation,
				"model", geminiModel,
			)
			geminiClient, err := gemini.New(ctx, geminiProject, geminiLocation, gemini.WithModel(geminiModel))
			if err != nil {
				return goerr.Wrap(err, "failed to create Gemini client")
			}

			// Configure storage adapter
			var storageAdapter interfaces.StorageAdapter
			if storageBucket != "" {
				// Use Cloud Storage
				logger.Info("using Cloud Storage for history",
					"bucket", storageBucket,
					"prefix", storagePrefix,
				)

				opts := []cs.Option{}
				if storagePrefix != "" {
					opts = append(opts, cs.WithPrefix(storagePrefix))
				}

				csClient, err := cs.New(ctx, storageBucket, opts...)
				if err != nil {
					return goerr.Wrap(err, "failed to create Cloud Storage client")
				}
				defer csClient.Close()
				storageAdapter = csClient
			} else {
				// Use memory storage
				logger.Warn("using in-memory storage for history (data will be lost on restart)")
				storageAdapter = mem_adapter.New()
			}

			// Create storage repository
			storageRepo := storage.New(storageAdapter)

			// Configure database repository
			var repo interfaces.ThreadRepository
			firestoreCfg.SetDefaults()

			if firestoreCfg.ProjectID != "" {
				// Use Firestore if project ID is provided
				logger.Info("using Firestore repository",
					"project_id", firestoreCfg.ProjectID,
					"database_id", firestoreCfg.DatabaseID,
				)

				client, err := firestore.New(ctx, firestoreCfg.ProjectID, firestoreCfg.DatabaseID)
				if err != nil {
					return goerr.Wrap(err, "failed to create firestore client")
				}
				defer client.Close()
				repo = client
			} else {
				// Use memory repository as fallback
				logger.Warn("using in-memory repository (data will be lost on restart)")
				repo = memory.New()
			}

			logger.Info("starting server",
				"addr", addr,
				"slack", slackCfg,
			)

			// Configure Slack service
			slackSvc, err := slackCfg.Configure()
			if err != nil {
				return goerr.Wrap(err, "failed to configure slack service")
			}

			// Create usecase with LLM integration
			uc := usecase.New(
				usecase.WithSlackClient(slackSvc),
				usecase.WithRepository(repo),
				usecase.WithStorageRepository(storageRepo),
				usecase.WithGeminiClient(geminiClient),
				usecase.WithGeminiModel(geminiModel),
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
