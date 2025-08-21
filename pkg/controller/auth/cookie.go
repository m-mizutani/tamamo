package auth

import (
	"net/http"

	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
)

// GetSessionFromRequest extracts the session from the request
func GetSessionFromRequest(r *http.Request, getSession func(sessionID string) (*auth.Session, error)) (*auth.Session, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, auth.ErrSessionNotFound
	}

	session, err := getSession(cookie.Value)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// IsAuthenticated checks if the request has a valid session
func IsAuthenticated(r *http.Request, getSession func(sessionID string) (*auth.Session, error)) bool {
	session, err := GetSessionFromRequest(r, getSession)
	if err != nil {
		return false
	}

	return session != nil && session.IsValid()
}
