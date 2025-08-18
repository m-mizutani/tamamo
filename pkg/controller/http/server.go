package http

import (
	"context"
	"io/fs"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/frontend"
	graphql_controller "github.com/m-mizutani/tamamo/pkg/controller/graphql"
	slack_controller "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/utils/errors"
	"github.com/m-mizutani/tamamo/pkg/utils/safe"
)

// Server represents the HTTP server
type Server struct {
	router         *chi.Mux
	slackCtrl      *slack_controller.Controller
	graphqlCtrl    *graphql_controller.Resolver
	enableGraphiQL bool
	slackVerifier  slack.PayloadVerifier
}

// Options is a functional option for Server
type Options func(*Server)

// WithSlackController sets the Slack controller
func WithSlackController(ctrl *slack_controller.Controller) Options {
	return func(s *Server) {
		s.slackCtrl = ctrl
	}
}

// WithSlackVerifier sets the Slack payload verifier
func WithSlackVerifier(verifier slack.PayloadVerifier) Options {
	return func(s *Server) {
		s.slackVerifier = verifier
	}
}

// WithGraphQLController sets the GraphQL controller
func WithGraphQLController(ctrl *graphql_controller.Resolver) Options {
	return func(s *Server) {
		s.graphqlCtrl = ctrl
	}
}

// WithGraphiQL enables GraphiQL IDE
func WithGraphiQL(enable bool) Options {
	return func(s *Server) {
		s.enableGraphiQL = enable
	}
}

// New creates a new HTTP server
func New(opts ...Options) *Server {
	r := chi.NewRouter()

	s := &Server{
		router: r,
	}
	for _, opt := range opts {
		opt(s)
	}

	// Apply middleware
	r.Use(loggingMiddleware)
	r.Use(panicRecoveryMiddleware)

	// Register routes
	r.Route("/hooks", func(r chi.Router) {
		r.Route("/slack", func(r chi.Router) {
			// Apply Slack signature verification middleware
			if s.slackVerifier != nil {
				r.Use(verifySlackSignature(s.slackVerifier))
			}
			r.Post("/event", slackEventHandler(s.slackCtrl))
			// Future: r.Post("/interaction", slackInteractionHandler(s.slackCtrl))
		})
	})

	// GraphQL endpoints
	if s.graphqlCtrl != nil {
		srv := handler.NewDefaultServer(
			graphql_controller.NewExecutableSchema(
				graphql_controller.Config{
					Resolvers: s.graphqlCtrl,
				},
			),
		)
		r.Handle("/graphql", srv)

		// GraphiQL IDE (optional)
		if s.enableGraphiQL {
			r.Handle("/graphiql", playground.Handler("GraphQL playground", "/graphql"))
		}
	}

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		safe.Write(r.Context(), w, []byte("OK"))
	})

	// Serve frontend static files
	distFS, err := fs.Sub(frontend.StaticFiles, "dist")
	if err != nil {
		// Log error but continue without serving frontend
		// In production, frontend should be built and embedded
		wrappedErr := goerr.Wrap(err, "failed to load frontend static files")
		errors.Handle(context.Background(), wrappedErr)
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			safe.Write(r.Context(), w, []byte("Frontend not available"))
		})
	} else {
		r.Handle("/*", http.FileServer(http.FS(distFS)))
	}

	return s
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
