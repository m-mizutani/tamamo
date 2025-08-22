package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
)

const (
	sessionCookieName    = "tamamo_session"
	oauthStateCookieName = "oauth_state"
)

// Controller handles authentication HTTP endpoints
type Controller struct {
	authUseCase  interfaces.AuthUseCases
	userUseCase  interfaces.UserUseCases
	frontendURL  string
	isProduction bool
}

// NewController creates a new authentication controller
func NewController(authUseCase interfaces.AuthUseCases, userUseCase interfaces.UserUseCases, frontendURL string, isProduction bool) *Controller {
	return &Controller{
		authUseCase:  authUseCase,
		userUseCase:  userUseCase,
		frontendURL:  frontendURL,
		isProduction: isProduction,
	}
}

// generateState generates a random state parameter for OAuth
func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", goerr.Wrap(err, "failed to generate random state")
	}
	return hex.EncodeToString(bytes), nil
}

// HandleLogin initiates the OAuth login flow
func (c *Controller) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Generate state parameter for CSRF protection
	state, err := generateState()
	if err != nil {
		ctxlog.From(ctx).Error("Failed to generate state", "error", err)
		c.writeError(w, http.StatusInternalServerError, "Failed to initiate login")
		return
	}

	// Store state in cookie for verification
	stateCookie := &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.isProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	}
	http.SetCookie(w, stateCookie)

	// Generate OAuth login URL with state
	loginURL, err := c.authUseCase.GenerateLoginURL(ctx, state)
	if err != nil {
		ctxlog.From(ctx).Error("Failed to generate login URL", "error", err)
		c.writeError(w, http.StatusInternalServerError, "Failed to initiate login")
		return
	}

	// Redirect to Slack OAuth
	http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

// HandleCallback processes the OAuth callback
func (c *Controller) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Check for OAuth errors
	if errorParam != "" {
		ctxlog.From(ctx).Warn("OAuth error received", "error", errorParam)
		c.clearStateCookie(w)
		http.Redirect(w, r, c.frontendURL+"?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	// Validate parameters
	if code == "" || state == "" {
		ctxlog.From(ctx).Warn("Missing OAuth parameters",
			"code_present", code != "",
			"state_present", state != "")
		c.clearStateCookie(w)
		http.Redirect(w, r, c.frontendURL+"?error=invalid_request", http.StatusTemporaryRedirect)
		return
	}

	// Verify state from cookie
	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		ctxlog.From(ctx).Warn("Missing state cookie", "error", err)
		http.Redirect(w, r, c.frontendURL+"?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	if state != stateCookie.Value {
		ctxlog.From(ctx).Warn("State mismatch", "expected", stateCookie.Value, "received", state)
		c.clearStateCookie(w)
		http.Redirect(w, r, c.frontendURL+"?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	// Clear state cookie after successful verification
	c.clearStateCookie(w)

	// Handle the callback (no longer needs state validation)
	session, err := c.authUseCase.HandleCallback(ctx, code)
	if err != nil {
		ctxlog.From(ctx).Error("Failed to handle OAuth callback", "error", err)
		http.Redirect(w, r, c.frontendURL+"?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	// Set session cookie
	c.setSessionCookie(w, session.ID.String(), session.ExpiresAt)

	// Redirect to frontend
	http.Redirect(w, r, c.frontendURL, http.StatusTemporaryRedirect)
}

// HandleLogout logs out the user
func (c *Controller) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session ID from cookie
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		// No session cookie, consider it a successful logout
		c.clearSessionCookie(w)
		c.writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out"})
		return
	}

	// Logout
	if err := c.authUseCase.Logout(ctx, cookie.Value); err != nil {
		ctxlog.From(ctx).Error("Failed to logout", "error", err)
		// Still clear the cookie even if logout fails
	}

	// Clear session cookie
	c.clearSessionCookie(w)
	c.writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out"})
}

// HandleMe returns the current user information
func (c *Controller) HandleMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session ID from cookie
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		c.writeError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get session
	session, err := c.authUseCase.GetSession(ctx, cookie.Value)
	if err != nil {
		if err == auth.ErrSessionNotFound || err == auth.ErrSessionExpired {
			c.clearSessionCookie(w)
			c.writeError(w, http.StatusUnauthorized, "Session expired")
			return
		}

		ctxlog.From(ctx).Error("Failed to get session", "error", err)
		c.writeError(w, http.StatusInternalServerError, "Failed to get user information")
		return
	}

	// Return user information
	c.writeJSON(w, http.StatusOK, &UserResponse{
		ID:       session.UserID.String(),
		Name:     session.UserName,
		Email:    session.Email,
		TeamID:   session.TeamID,
		TeamName: session.TeamName,
	})
}

// HandleCheck checks if the user is authenticated
func (c *Controller) HandleCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session ID from cookie
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		c.writeJSON(w, http.StatusOK, &AuthCheckResponse{
			Authenticated: false,
		})
		return
	}

	// Validate session
	session, err := c.authUseCase.GetSession(ctx, cookie.Value)
	if err != nil {
		if err == auth.ErrSessionNotFound || err == auth.ErrSessionExpired {
			c.clearSessionCookie(w)
		}
		c.writeJSON(w, http.StatusOK, &AuthCheckResponse{
			Authenticated: false,
		})
		return
	}

	// Get full user information to include DisplayName
	user, err := c.userUseCase.GetUserByID(ctx, session.UserID)
	if err != nil {
		ctxlog.From(ctx).Error("Failed to get user information", "error", err, "user_id", session.UserID)
		c.writeJSON(w, http.StatusOK, &AuthCheckResponse{
			Authenticated: false,
		})
		return
	}

	// Return authentication status
	c.writeJSON(w, http.StatusOK, &AuthCheckResponse{
		Authenticated: true,
		User: &UserResponse{
			ID:          session.UserID.String(),
			Name:        session.UserName,
			DisplayName: user.DisplayName,
			Email:       session.Email,
			TeamID:      session.TeamID,
			TeamName:    session.TeamName,
		},
	})
}

// setSessionCookie sets the session cookie
func (c *Controller) setSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.isProduction,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	}
	http.SetCookie(w, cookie)
}

// clearSessionCookie clears the session cookie
func (c *Controller) clearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   c.isProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
	}
	http.SetCookie(w, cookie)
}

// clearStateCookie clears the OAuth state cookie
func (c *Controller) clearStateCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   c.isProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

// writeJSON writes a JSON response
func (c *Controller) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log the error but don't return it since response is already being written
		ctxlog.From(context.Background()).Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response
func (c *Controller) writeError(w http.ResponseWriter, status int, message string) {
	c.writeJSON(w, status, &ErrorResponse{
		Error: ErrorDetail{
			Code:    http.StatusText(status),
			Message: message,
		},
	})
}

// Middleware creates authentication middleware
func (c *Controller) Middleware(required bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := ctxlog.From(ctx)

			// Skip authentication for auth endpoints
			if c.isAuthEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Try to get session from cookie
			session, err := c.getSessionFromRequest(ctx, r)
			if err != nil {
				if required {
					logger.Debug("authentication required but no valid session found", "error", err)
					c.clearSessionCookie(w)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				// For optional auth, continue without session
			} else {
				// Add session to context
				ctx = ContextWithUser(ctx, session)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequiredAuth creates required authentication middleware
func (c *Controller) RequiredAuth() func(http.Handler) http.Handler {
	return c.Middleware(true)
}

// OptionalAuth creates optional authentication middleware
func (c *Controller) OptionalAuth() func(http.Handler) http.Handler {
	return c.Middleware(false)
}

// getSessionFromRequest extracts session from request cookie
func (c *Controller) getSessionFromRequest(ctx context.Context, r *http.Request) (*auth.Session, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, goerr.Wrap(err, "no session cookie found")
	}

	session, err := c.authUseCase.GetSession(ctx, cookie.Value)
	if err != nil {
		return nil, goerr.Wrap(err, "invalid session")
	}

	return session, nil
}

// isAuthEndpoint checks if the path is an authentication endpoint
func (c *Controller) isAuthEndpoint(path string) bool {
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

// Context management for user sessions

type contextKey string

const (
	userContextKey contextKey = "user"
)

// ContextWithUser adds a user session to the context
func ContextWithUser(ctx context.Context, session *auth.Session) context.Context {
	return context.WithValue(ctx, userContextKey, session)
}

// UserFromContext extracts the user session from the context
func UserFromContext(ctx context.Context) (*auth.Session, bool) {
	session, ok := ctx.Value(userContextKey).(*auth.Session)
	return session, ok
}

// RequireUserFromContext extracts the user session from the context and panics if not found
func RequireUserFromContext(ctx context.Context) *auth.Session {
	session, ok := UserFromContext(ctx)
	if !ok {
		panic("user not found in context")
	}
	return session
}

// GetSessionFromContext retrieves session from context (deprecated: use UserFromContext)
func GetSessionFromContext(ctx context.Context) *auth.Session {
	session, _ := UserFromContext(ctx)
	return session
}
