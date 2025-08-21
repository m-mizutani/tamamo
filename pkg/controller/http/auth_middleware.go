package http

import (
	"context"
	"net/http"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/tamamo/pkg/controller/http/middleware"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
)

// sessionContextKey is deprecated - using middleware.ContextWithUser instead
// type sessionContextKey struct{}

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authUseCase interfaces.AuthUseCases, noAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := ctxlog.From(ctx)

			// If no-auth mode is enabled, skip authentication
			if noAuth {
				// Create anonymous session
				anonymousSession := &auth.Session{
					UserID:   "anonymous",
					UserName: "Anonymous User",
					Email:    "anonymous@local",
					TeamID:   "anonymous",
					TeamName: "Anonymous Team",
				}
				ctx = middleware.ContextWithUser(ctx, anonymousSession)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Skip authentication for auth endpoints
			if isAuthEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Get session cookie
			cookie, err := r.Cookie("tamamo_session")
			if err != nil {
				// No session cookie
				logger.Debug("no session cookie found")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate session
			session, err := authUseCase.GetSession(ctx, cookie.Value)
			if err != nil {
				logger.Debug("invalid session", "error", err)
				// Clear invalid cookie
				clearSessionCookie(w)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add session to context
			ctx = middleware.ContextWithUser(ctx, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware creates optional authentication middleware (for public endpoints)
func OptionalAuthMiddleware(authUseCase interfaces.AuthUseCases, noAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// If no-auth mode is enabled, create anonymous session
			if noAuth {
				anonymousSession := &auth.Session{
					UserID:   "anonymous",
					UserName: "Anonymous User",
					Email:    "anonymous@local",
					TeamID:   "anonymous",
					TeamName: "Anonymous Team",
				}
				ctx = middleware.ContextWithUser(ctx, anonymousSession)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Try to get session from cookie
			if cookie, err := r.Cookie("tamamo_session"); err == nil {
				if session, err := authUseCase.GetSession(ctx, cookie.Value); err == nil {
					// Add session to context if valid
					ctx = middleware.ContextWithUser(ctx, session)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetSessionFromContext retrieves session from context
func GetSessionFromContext(ctx context.Context) *auth.Session {
	session, _ := middleware.UserFromContext(ctx)
	return session
}

// isAuthEndpoint checks if the path is an authentication endpoint
func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/auth/login",
		"/api/auth/callback",
		"/api/auth/logout",
		"/api/auth/check",
		"/api/auth/me",
		"/health",
	}

	for _, authPath := range authPaths {
		if path == authPath {
			return true
		}
	}

	return false
}

// clearSessionCookie clears the session cookie
func clearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "tamamo_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}
