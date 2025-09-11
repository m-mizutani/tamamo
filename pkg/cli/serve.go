package cli

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/adapters/cs"
	"github.com/m-mizutani/tamamo/pkg/adapters/fs"
	"github.com/m-mizutani/tamamo/pkg/cli/config"
	auth_controller "github.com/m-mizutani/tamamo/pkg/controller/auth"
	graphql_controller "github.com/m-mizutani/tamamo/pkg/controller/graphql"
	server "github.com/m-mizutani/tamamo/pkg/controller/http"
	slack_controller "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/repository/database/firestore"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/repository/storage"
	"github.com/m-mizutani/tamamo/pkg/service/image"
	"github.com/m-mizutani/tamamo/pkg/service/jira"
	"github.com/m-mizutani/tamamo/pkg/service/notion"
	"github.com/m-mizutani/tamamo/pkg/service/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
	"github.com/urfave/cli/v3"
)

func cmdServe() *cli.Command {
	var (
		addr           string
		appCfg         config.App
		slackCfg       config.Slack
		firestoreCfg   config.Firestore
		authCfg        config.Auth
		llmCfg         config.LLMConfig
		storageCfg     config.Storage
		jiraCfg        config.Jira
		notionCfg      config.Notion
		enableGraphiQL bool
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
		&cli.BoolFlag{
			Name:        "enable-graphiql",
			Sources:     cli.EnvVars("TAMAMO_ENABLE_GRAPHIQL"),
			Usage:       "Enable GraphiQL IDE for development",
			Destination: &enableGraphiQL,
		},
	}
	flags = append(flags, appCfg.Flags()...)
	flags = append(flags, slackCfg.Flags()...)
	flags = append(flags, firestoreCfg.Flags()...)
	flags = append(flags, authCfg.Flags()...)
	flags = append(flags, llmCfg.Flags()...)
	flags = append(flags, storageCfg.Flags()...)
	flags = append(flags, jiraCfg.Flags()...)
	flags = append(flags, notionCfg.Flags()...)

	return &cli.Command{
		Name:    "serve",
		Aliases: []string{"s"},
		Usage:   "Run server",
		Flags:   flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger := ctxlog.From(ctx)

			// Validate application configuration
			if err := appCfg.Validate(); err != nil {
				return goerr.Wrap(err, "invalid application configuration")
			}

			// Validate authentication configuration
			if err := authCfg.Validate(); err != nil {
				return goerr.Wrap(err, "invalid authentication configuration")
			}

			// Validate storage configuration
			if err := storageCfg.Validate(); err != nil {
				return goerr.Wrap(err, "invalid storage configuration")
			}

			// Load and validate LLM configuration
			providersConfig, err := llmCfg.LoadAndValidate()
			if err != nil {
				return goerr.Wrap(err, "failed to load LLM configuration")
			}

			// Build LLM factory
			llmFactory, err := llmCfg.BuildFactory(ctx, providersConfig)
			if err != nil {
				return goerr.Wrap(err, "failed to build LLM factory")
			}

			// Log consolidated LLM configuration summary
			providerSummary := make([]string, 0, len(providersConfig.Providers))
			for id, provider := range providersConfig.Providers {
				providerSummary = append(providerSummary, fmt.Sprintf("%s(%d)", id, len(provider.Models)))
			}

			var fallbackInfo string
			if providersConfig.Fallback.Enabled {
				fallbackInfo = fmt.Sprintf("fallback=%s:%s", providersConfig.Fallback.Provider, providersConfig.Fallback.Model)
			} else {
				fallbackInfo = "fallback=disabled"
			}

			logger.Info("LLM configuration loaded",
				"providers", providerSummary,
				"default", fmt.Sprintf("%s:%s", providersConfig.Defaults.Provider, providersConfig.Defaults.Model),
				"fallback_config", fallbackInfo,
			)

			// Configure storage adapter
			logger.Info("configuring storage adapter")
			storageAdapter, storageCleanup, err := storageCfg.CreateAdapter(ctx)
			if err != nil {
				return goerr.Wrap(err, "failed to create storage adapter")
			}
			if storageCleanup != nil {
				defer storageCleanup()
			}

			if storageCfg.HasCloudStorage() {
				logger.Info("using Cloud Storage for history",
					"bucket", storageCfg.Bucket,
					"prefix", storageCfg.Prefix,
				)
			} else {
				logger.Info("using file system storage for history", "path", storageCfg.FSPath)
			}

			// Create storage repository
			storageRepo := storage.New(storageAdapter)

			// Configure database repository
			var repo interfaces.ThreadRepository
			var agentRepo interfaces.AgentRepository
			var sessionRepo interfaces.SessionRepository
			var userRepo interfaces.UserRepository
			var agentImageRepo interfaces.AgentImageRepository
			var slackMessageLogRepo interfaces.SlackMessageLogRepository
			var slackSearchConfigRepo interfaces.SlackSearchConfigRepository
			var jiraSearchConfigRepo interfaces.JiraSearchConfigRepository
			var notionSearchConfigRepo interfaces.NotionSearchConfigRepository
			firestoreCfg.SetDefaults()

			// Validate Firestore configuration
			if err := firestoreCfg.Validate(); err != nil {
				return goerr.Wrap(err, "invalid firestore configuration")
			}

			// DEBUG: Log Firestore configuration
			logger.Info("Firestore configuration check",
				"project_id", firestoreCfg.ProjectID,
				"database_id", firestoreCfg.DatabaseID,
				"project_id_empty", firestoreCfg.ProjectID == "",
			)

			if firestoreCfg.ProjectID != "" {
				// Use Firestore if project ID is provided
				logger.Info("attempting to create Firestore repository",
					"project_id", firestoreCfg.ProjectID,
					"database_id", firestoreCfg.DatabaseID,
				)

				client, err := firestore.New(ctx, firestoreCfg.ProjectID, firestoreCfg.DatabaseID)
				if err != nil {
					return goerr.Wrap(err, "failed to create firestore client")
				}
				logger.Info("successfully created Firestore repository",
					"project_id", firestoreCfg.ProjectID,
					"database_id", firestoreCfg.DatabaseID,
				)
				defer client.Close()
				repo = client
				agentRepo = client
				sessionRepo = firestore.NewSessionRepository(client.GetClient())
				userRepo = firestore.NewUserRepository(client.GetClient())
				agentImageRepo = client.NewAgentImageRepository()
				slackMessageLogRepo = client
				slackSearchConfigRepo = firestore.NewSlackSearchConfigRepository(client.GetClient())
				jiraSearchConfigRepo = firestore.NewJiraSearchConfigRepository(client.GetClient())
				notionSearchConfigRepo = firestore.NewNotionSearchConfigRepository(client.GetClient())
			} else {
				// Use memory repository as fallback
				logger.Warn("using in-memory repository (data will be lost on restart)")
				memoryClient := memory.New()
				repo = memoryClient
				agentRepo = memory.NewAgentMemoryClient()
				sessionRepo = memory.NewSessionRepository()
				userRepo = memory.NewUserRepository()
				agentImageRepo = memory.NewAgentImageRepository()
				slackMessageLogRepo = memoryClient
				slackSearchConfigRepo = memory.NewSlackSearchConfigRepository()
				jiraSearchConfigRepo = memory.NewJiraSearchConfigRepository()
				notionSearchConfigRepo = memory.NewNotionSearchConfigRepository()
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

			// Configure image storage adapter (use same storage config but with images subdirectory)
			logger.Info("configuring image storage adapter")
			var imageStorageAdapter interfaces.StorageAdapter
			if storageCfg.HasCloudStorage() {
				// Use Cloud Storage for images
				opts := []cs.Option{}
				if storageCfg.Prefix != "" {
					opts = append(opts, cs.WithPrefix(storageCfg.Prefix))
				}

				csClient, err := cs.New(ctx, storageCfg.Bucket, opts...)
				if err != nil {
					return goerr.Wrap(err, "failed to create Cloud Storage client for images")
				}
				defer csClient.Close()
				imageStorageAdapter = csClient
				logger.Info("using Cloud Storage for images",
					"bucket", storageCfg.Bucket,
					"prefix", storageCfg.Prefix,
				)
			} else {
				// Use file system storage for images (use same base path with images subdirectory)
				imagePath := storageCfg.FSPath + "/images"
				fsClient, err := fs.New(&fs.Config{BaseDirectory: imagePath})
				if err != nil {
					return goerr.Wrap(err, "failed to create file system storage adapter for images")
				}
				imageStorageAdapter = fsClient
				logger.Info("using file system storage for images", "path", imagePath)
			}

			// Create image processor
			validator := image.NewValidator()
			config := image.DefaultProcessorConfig()

			// DEBUG: Log repository types being used for image processor
			logger.Info("Creating image processor with repositories",
				"agentRepo_type", fmt.Sprintf("%T", agentRepo),
				"agentImageRepo_type", fmt.Sprintf("%T", agentImageRepo),
			)

			imageProcessor := image.NewProcessor(validator, imageStorageAdapter, agentImageRepo, agentRepo, config)

			// Configure authentication use case
			var authUseCase interfaces.AuthUseCases
			var authCtrl *auth_controller.Controller
			if authCfg.IsAuthenticationEnabled() {
				authUseCase, err = authCfg.ConfigureAuthUseCase(sessionRepo, userUseCase, slackSvc)
				if err != nil {
					return goerr.Wrap(err, "failed to configure authentication")
				}
				authCtrl = auth_controller.NewController(authUseCase, userUseCase, authCfg.FrontendURL)
				logger.Info("authentication enabled")
			} else {
				logger.Warn("authentication disabled - running in anonymous mode")
			}

			// Create usecase with LLM integration
			// Use FRONTEND_URL as base for agent image URLs (public access)
			serverBaseURL := os.Getenv("FRONTEND_URL")
			if serverBaseURL == "" {
				// Fallback to public URL or listen address
				serverBaseURL = os.Getenv("TAMAMO_PUBLIC_URL")
				if serverBaseURL == "" {
					serverBaseURL = "http://" + addr
				}
			}
			uc := usecase.New(
				usecase.WithSlackClient(slackSvc),
				usecase.WithRepository(repo),
				usecase.WithAgentRepository(agentRepo),
				usecase.WithAgentImageRepository(agentImageRepo),
				usecase.WithSlackMessageLogRepository(slackMessageLogRepo),
				usecase.WithStorageRepository(storageRepo),
				usecase.WithLLMFactory(llmFactory),
				usecase.WithServerBaseURL(serverBaseURL),
			)

			// Create controllers
			slackCtrl := slack_controller.New(uc, slackSvc)

			// Create agent use case
			agentUseCase := usecase.NewAgentUseCases(agentRepo)

			// Create search config use cases
			slackSearchConfigUseCases := usecase.NewSlackSearchConfig(
				usecase.WithSlackSearchConfigRepository(slackSearchConfigRepo),
				usecase.WithSlackSearchConfigAgentRepository(agentRepo),
			)
			jiraSearchConfigUseCases := usecase.NewJiraSearchConfig(
				usecase.WithJiraSearchConfigRepository(jiraSearchConfigRepo),
				usecase.WithJiraSearchConfigAgentRepository(agentRepo),
			)
			notionSearchConfigUseCases := usecase.NewNotionSearchConfig(
				usecase.WithNotionSearchConfigRepository(notionSearchConfigRepo),
				usecase.WithNotionSearchConfigAgentRepository(agentRepo),
			)

			// Create Jira integration components (if configured)
			var jiraUseCases usecase.JiraIntegrationUseCases
			var jiraAuthController *server.JiraAuthController
			if jiraCfg.IsEnabled() {
				if err := jiraCfg.Validate(); err != nil {
					return goerr.Wrap(err, "invalid Jira configuration")
				}

				jiraOAuthConfig := jiraCfg.BuildOAuthConfig(appCfg.FrontendURL)
				jiraOAuthService := jira.NewOAuthService(jiraOAuthConfig)
				jiraUseCases = usecase.NewJiraIntegrationUseCases(userRepo, jiraOAuthService)
				jiraAuthController = server.NewJiraAuthController(jiraUseCases, jiraOAuthService)

				logger.Info("Jira integration enabled",
					"client_id", jiraOAuthConfig.ClientID,
					"redirect_uri", jiraOAuthConfig.RedirectURI,
				)
			} else {
				logger.Info("Jira integration disabled (missing configuration)")
			}

			// Create Notion integration components (if configured)
			var notionUseCases usecase.NotionIntegrationUseCases
			var notionAuthController *server.NotionAuthController
			if notionCfg.IsEnabled() {
				if err := notionCfg.Validate(); err != nil {
					return goerr.Wrap(err, "invalid Notion configuration")
				}

				notionOAuthConfig := notionCfg.BuildOAuthConfig(appCfg.FrontendURL)
				notionOAuthService := notion.NewOAuthService(notionOAuthConfig)
				notionUseCases = usecase.NewNotionIntegrationUseCases(userRepo, notionOAuthService)
				notionAuthController = server.NewNotionAuthController(notionUseCases, notionOAuthService)

				logger.Info("Notion integration enabled",
					"client_id", notionOAuthConfig.ClientID,
					"redirect_uri", notionOAuthConfig.RedirectURI,
				)
			} else {
				logger.Info("Notion integration disabled (missing configuration)")
			}

			graphqlCtrl := graphql_controller.NewResolver(repo, agentUseCase, userUseCase, llmFactory, imageProcessor, agentImageRepo, jiraUseCases, notionUseCases, slackSearchConfigUseCases, jiraSearchConfigUseCases, notionSearchConfigUseCases)

			// Create user controller
			userCtrl := server.NewUserController(userUseCase)

			// Create image use case and controller
			imageUseCase := usecase.NewImageUseCases(imageProcessor, agentImageRepo, agentUseCase)
			imageCtrl := server.NewImageController(imageUseCase)

			// Build HTTP server options
			serverOptions := []server.Options{
				server.WithSlackController(slackCtrl),
				server.WithGraphQLController(graphqlCtrl),
				server.WithUserController(userCtrl),
				server.WithImageController(imageCtrl),
				server.WithGraphiQL(enableGraphiQL),
				server.WithSlackVerifier(slackCfg.Verifier()),
				server.WithNoAuth(authCfg.NoAuthentication),
			}

			// Add Jira auth controller if configured
			if jiraAuthController != nil {
				serverOptions = append(serverOptions, server.WithJiraAuthController(jiraAuthController))
			}

			// Add Notion auth controller if configured
			if notionAuthController != nil {
				serverOptions = append(serverOptions, server.WithNotionAuthController(notionAuthController))
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
