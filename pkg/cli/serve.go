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
	"github.com/m-mizutani/tamamo/pkg/adapters/cs"
	mem_adapter "github.com/m-mizutani/tamamo/pkg/adapters/memory"
	"github.com/m-mizutani/tamamo/pkg/cli/config"
	auth_controller "github.com/m-mizutani/tamamo/pkg/controller/auth"
	graphql_controller "github.com/m-mizutani/tamamo/pkg/controller/graphql"
	server "github.com/m-mizutani/tamamo/pkg/controller/http"
	slack_controller "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/repository/database/firestore"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/repository/storage"
	"github.com/m-mizutani/tamamo/pkg/service/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/urfave/cli/v3"
)

func cmdServe() *cli.Command {
	var (
		addr           string
		slackCfg       config.Slack
		firestoreCfg   config.Firestore
		authCfg        config.Auth
		llmCfg         config.LLMConfig
		storageBucket  string
		storagePrefix  string
		enableGraphiQL bool
		isProduction   bool
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
		&cli.BoolFlag{
			Name:        "enable-graphiql",
			Sources:     cli.EnvVars("TAMAMO_ENABLE_GRAPHIQL"),
			Usage:       "Enable GraphiQL IDE for development",
			Destination: &enableGraphiQL,
		},
		&cli.BoolFlag{
			Name:        "production",
			Sources:     cli.EnvVars("TAMAMO_PRODUCTION"),
			Usage:       "Enable production mode (sets secure cookie attributes)",
			Destination: &isProduction,
		},
	}
	flags = append(flags, slackCfg.Flags()...)
	flags = append(flags, firestoreCfg.Flags()...)
	flags = append(flags, authCfg.Flags()...)
	flags = append(flags, llmCfg.Flags()...)

	return &cli.Command{
		Name:    "serve",
		Aliases: []string{"s"},
		Usage:   "Run server",
		Flags:   flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger := ctxlog.From(ctx)

			// Validate authentication configuration
			if err := authCfg.Validate(); err != nil {
				return goerr.Wrap(err, "invalid authentication configuration")
			}

			// Load and validate LLM configuration
			logger.Info("loading LLM providers configuration")
			providersConfig, err := llmCfg.LoadAndValidate()
			if err != nil {
				return goerr.Wrap(err, "failed to load LLM configuration")
			}

			// Log detailed provider information
			logger.Info("LLM providers configuration loaded",
				"provider_count", len(providersConfig.Providers),
			)

			// Log each provider and its models
			for id, provider := range providersConfig.Providers {
				modelNames := make([]string, len(provider.Models))
				for i, model := range provider.Models {
					modelNames[i] = model.ID
				}
				logger.Info("LLM provider enabled",
					"provider_id", id,
					"display_name", provider.DisplayName,
					"model_count", len(provider.Models),
					"models", modelNames,
				)
			}

			// Log default configuration
			if providersConfig.Defaults.Provider != "" {
				logger.Info("Default LLM configuration",
					"provider", providersConfig.Defaults.Provider,
					"model", providersConfig.Defaults.Model,
				)
			}

			// Log fallback configuration
			if providersConfig.Fallback.Enabled {
				logger.Info("Fallback LLM configuration",
					"enabled", true,
					"provider", providersConfig.Fallback.Provider,
					"model", providersConfig.Fallback.Model,
				)
			} else {
				logger.Info("Fallback LLM configuration",
					"enabled", false,
				)
			}

			// Build LLM factory
			logger.Info("Building LLM factory")
			llmFactory, err := llmCfg.BuildFactory(ctx, providersConfig)
			if err != nil {
				return goerr.Wrap(err, "failed to build LLM factory")
			}
			logger.Info("LLM factory built successfully")

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
			var agentRepo interfaces.AgentRepository
			var sessionRepo interfaces.SessionRepository
			var userRepo interfaces.UserRepository
			firestoreCfg.SetDefaults()

			// Validate Firestore configuration
			if err := firestoreCfg.Validate(); err != nil {
				return goerr.Wrap(err, "invalid firestore configuration")
			}

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
				agentRepo = client // Firestore client implements both ThreadRepository and AgentRepository
				sessionRepo = firestore.NewSessionRepository(client.GetClient())
				userRepo = firestore.NewUserRepository(client.GetClient())
			} else {
				// Use memory repository as fallback
				logger.Warn("using in-memory repository (data will be lost on restart)")
				repo = memory.New()
				agentRepo = memory.NewAgentMemoryClient()
				sessionRepo = memory.NewSessionRepository()
				userRepo = memory.NewUserRepository()
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

			// Create avatar service
			avatarService := slack.NewAvatarService(slackSvc)

			// Create user use case
			userUseCase := usecase.NewUserUseCase(userRepo, avatarService, slackSvc)

			// Configure authentication use case
			var authUseCase interfaces.AuthUseCases
			var authCtrl *auth_controller.Controller
			if authCfg.IsAuthenticationEnabled() {
				authUseCase, err = authCfg.ConfigureAuthUseCase(sessionRepo, userUseCase, slackSvc)
				if err != nil {
					return goerr.Wrap(err, "failed to configure authentication")
				}
				authCtrl = auth_controller.NewController(authUseCase, userUseCase, authCfg.FrontendURL, isProduction)
				logger.Info("authentication enabled")
			} else {
				logger.Warn("authentication disabled - running in anonymous mode")
			}

			// Create usecase with LLM integration
			uc := usecase.New(
				usecase.WithSlackClient(slackSvc),
				usecase.WithRepository(repo),
				usecase.WithAgentRepository(agentRepo),
				usecase.WithStorageRepository(storageRepo),
				usecase.WithLLMFactory(llmFactory),
			)

			// Create controllers
			slackCtrl := slack_controller.New(uc)

			// Create agent use case
			agentUseCase := usecase.NewAgentUseCases(agentRepo)
			graphqlCtrl := graphql_controller.NewResolver(repo, agentUseCase, userUseCase, llmFactory)

			// Create user controller
			userCtrl := server.NewUserController(userUseCase)

			// Build HTTP server options
			serverOptions := []server.Options{
				server.WithSlackController(slackCtrl),
				server.WithGraphQLController(graphqlCtrl),
				server.WithUserController(userCtrl),
				server.WithGraphiQL(enableGraphiQL),
				server.WithSlackVerifier(slackCfg.Verifier()),
				server.WithNoAuth(authCfg.NoAuthentication),
			}

			// Add auth controller if authentication is enabled
			if authCtrl != nil {
				serverOptions = append(serverOptions,
					server.WithAuthController(authCtrl),
					server.WithAuthUseCase(authUseCase),
				)
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
