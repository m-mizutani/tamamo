package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/tamamo/pkg/controller/auth"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	authmodel "github.com/m-mizutani/tamamo/pkg/domain/model/auth"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	authUseCase      interfaces.AuthUseCases
	noAuthentication bool
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authUseCase interfaces.AuthUseCases, noAuthentication bool) *AuthMiddleware {
	return &AuthMiddleware{
		authUseCase:      authUseCase,
		noAuthentication: noAuthentication,
	}
}

// Middleware returns the authentication middleware handler
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if disabled
		if m.noAuthentication {
			// Set anonymous user context
			ctx := ContextWithUser(r.Context(), &authmodel.Session{
				UserID:   "anonymous",
				UserName: "Anonymous User",
				Email:    "anonymous@localhost",
				TeamID:   "anonymous",
				TeamName: "Anonymous Team",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Skip authentication for auth endpoints
		if isAuthEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Skip authentication for static files (if serving frontend)
		if isStaticFile(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get session from cookie
		session, err := auth.GetSessionFromRequest(r, func(sessionID string) (*authmodel.Session, error) {
			return m.authUseCase.GetSession(r.Context(), sessionID)
		})

		if err != nil {
			if err == authmodel.ErrSessionNotFound || err == authmodel.ErrSessionExpired {
				// Clear invalid cookie
				clearSessionCookie(w)
				writeUnauthorizedResponse(w, "Authentication required")
				return
			}

			ctxlog.From(r.Context()).Error("Failed to get session", "error", err)
			writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Add user to context
		ctx := ContextWithUser(r.Context(), session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth is a middleware that requires authentication
func RequireAuth(authUseCase interfaces.AuthUseCases) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session from cookie
			session, err := auth.GetSessionFromRequest(r, func(sessionID string) (*authmodel.Session, error) {
				return authUseCase.GetSession(r.Context(), sessionID)
			})

			if err != nil {
				writeUnauthorizedResponse(w, "Authentication required")
				return
			}

			// Add user to context
			ctx := ContextWithUser(r.Context(), session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isAuthEndpoint checks if the path is an authentication endpoint
func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/auth/login",
		"/api/auth/callback",
		"/api/auth/logout",
		"/api/auth/check",
		"/api/auth/me",
	}

	for _, authPath := range authPaths {
		if path == authPath {
			return true
		}
	}

	return false
}

// isStaticFile checks if the path is a static file
func isStaticFile(path string) bool {
	// Common static file extensions
	staticExtensions := []string{
		".html", ".css", ".js", ".jsx", ".ts", ".tsx",
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".woff", ".woff2", ".ttf", ".eot",
		".json", ".map",
	}

	for _, ext := range staticExtensions {
		if len(path) > len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}

	// Check for root and common SPA routes
	if path == "/" || path == "/index.html" {
		return true
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

// writeUnauthorizedResponse writes an unauthorized response
func writeUnauthorizedResponse(w http.ResponseWriter, message string) {
	writeErrorResponse(w, http.StatusUnauthorized, message)
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    http.StatusText(status),
			"message": message,
		},
	})
}
