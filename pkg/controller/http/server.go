package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	slack_controller "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
)

// Server represents the HTTP server
type Server struct {
	router        *chi.Mux
	slackCtrl     *slack_controller.Controller
	slackVerifier slack.PayloadVerifier
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

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return s
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
