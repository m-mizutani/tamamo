package http

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/frontend"
	auth_controller "github.com/m-mizutani/tamamo/pkg/controller/auth"
	graphql_controller "github.com/m-mizutani/tamamo/pkg/controller/graphql"
	"github.com/m-mizutani/tamamo/pkg/controller/http/middleware"
	slack_controller "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/utils/errors"
	"github.com/m-mizutani/tamamo/pkg/utils/safe"
)

// Static file extensions that should be served directly (not fallback to SPA)
var staticFileExtensions = []string{
	".ico",            // favicon files
	".png",            // favicon PNG files
	".svg",            // SVG files
	".css",            // CSS files
	".js",             // JavaScript files
	".woff", ".woff2", // Web fonts
	".ttf", ".otf", // Font files
}

// Server represents the HTTP server
type Server struct {
	router         *chi.Mux
	slackCtrl      *slack_controller.Controller
	graphqlCtrl    *graphql_controller.Resolver
	authCtrl       *auth_controller.Controller
	authUseCase    interfaces.AuthUseCases
	enableGraphiQL bool
	slackVerifier  slack.PayloadVerifier
	noAuth         bool
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

// WithAuthController sets the authentication controller
func WithAuthController(ctrl *auth_controller.Controller) Options {
	return func(s *Server) {
		s.authCtrl = ctrl
	}
}

// WithAuthUseCase sets the authentication use case
func WithAuthUseCase(useCase interfaces.AuthUseCases) Options {
	return func(s *Server) {
		s.authUseCase = useCase
	}
}

// WithNoAuth enables no-authentication mode
func WithNoAuth(noAuth bool) Options {
	return func(s *Server) {
		s.noAuth = noAuth
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

	// Authentication routes (before auth middleware)
	if s.authCtrl != nil && !s.noAuth {
		r.Route("/api/auth", func(r chi.Router) {
			r.Get("/login", s.authCtrl.HandleLogin)
			r.Get("/callback", s.authCtrl.HandleCallback)
			r.Post("/logout", s.authCtrl.HandleLogout)
			r.Get("/check", s.authCtrl.HandleCheck)
			r.Get("/me", s.authCtrl.HandleMe)
		})
	}

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

	// GraphQL endpoints (with authentication)
	if s.graphqlCtrl != nil {
		r.Route("/graphql", func(r chi.Router) {
			// Apply authentication middleware if enabled
			if s.authUseCase != nil && !s.noAuth {
				r.Use(middleware.AuthMiddleware(s.authUseCase, s.noAuth))
			}

			srv := handler.NewDefaultServer(
				graphql_controller.NewExecutableSchema(
					graphql_controller.Config{
						Resolvers: s.graphqlCtrl,
					},
				),
			)
			// Add GraphQL-specific logging middleware
			r.Handle("/", graphQLLoggingMiddleware(srv))
		})

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

	// Static file serving for SPA
	staticFS, err := fs.Sub(frontend.StaticFiles, "dist")
	if err == nil {
		// Check if index.html exists
		if _, err := staticFS.Open("index.html"); err == nil {
			// Dedicated favicon handlers for better reliability
			r.Get("/favicon.ico", faviconHandler(staticFS, "favicon.ico", "image/x-icon"))

			// Serve static files and handle SPA routing
			r.HandleFunc("/*", spaHandler(staticFS, s.enableGraphiQL))
		}
	} else {
		// Log error but continue without serving frontend
		wrappedErr := goerr.Wrap(err, "failed to load frontend static files")
		errors.Handle(context.Background(), wrappedErr)
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			safe.Write(r.Context(), w, []byte("Frontend not available"))
		})
	}

	return s
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// faviconHandler serves favicon files with appropriate Content-Type headers
func faviconHandler(staticFS fs.FS, filename, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := staticFS.Open(filename)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		if _, err := io.Copy(w, file); err != nil {
			http.Error(w, "Failed to serve file", http.StatusInternalServerError)
			return
		}
	}
}

// spaHandler handles SPA routing by serving static files and falling back to index.html
func spaHandler(staticFS fs.FS, enableGraphiQL bool) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFS))

	return func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.TrimPrefix(r.URL.Path, "/")

		// Check for GraphiQL endpoint when it's disabled
		if urlPath == "graphiql" && !enableGraphiQL {
			http.NotFound(w, r)
			return
		}

		// If the path is empty, serve index.html
		if urlPath == "" {
			urlPath = "index.html"
		}

		// Try to open the file to check if it exists
		if file, err := staticFS.Open(urlPath); err != nil {
			// File not found

			// For SPA routes (not assets), serve index.html for client-side routing
			// But first check if this looks like an asset request
			isStaticFile := strings.HasPrefix(urlPath, "assets/") ||
				strings.HasPrefix(urlPath, "static/")

			// Check for static file extensions
			if !isStaticFile {
				for _, ext := range staticFileExtensions {
					if strings.HasSuffix(urlPath, ext) {
						isStaticFile = true
						break
					}
				}
			}

			if isStaticFile {
				// This looks like an asset request, return 404
				http.NotFound(w, r)
				return
			}

			// For SPA routes, serve index.html
			if indexFile, err := staticFS.Open("index.html"); err == nil {
				defer indexFile.Close()
				w.Header().Set("Content-Type", "text/html")
				if _, err := io.Copy(w, indexFile); err != nil {
					http.Error(w, "Failed to serve index.html", http.StatusInternalServerError)
					return
				}
				return
			}

			// If index.html is also not found, return 404
			http.NotFound(w, r)
			return
		} else {
			// File exists, close it and let fileServer handle it
			_ = file.Close() // Ignore error as file descriptor will be cleaned up by GC
		}

		// Set appropriate Content-Type for favicon files
		if strings.HasSuffix(urlPath, ".ico") {
			w.Header().Set("Content-Type", "image/x-icon")
		} else if strings.HasSuffix(urlPath, ".png") && strings.Contains(urlPath, "favicon") {
			w.Header().Set("Content-Type", "image/png")
		}

		// Serve the requested file using the file server
		fileServer.ServeHTTP(w, r)
	}
}
