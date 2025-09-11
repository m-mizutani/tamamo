package auth

import "errors"

var (
	// Session errors
	ErrInvalidSession  = errors.New("invalid session")
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionNotFound = errors.New("session not found")

	// OAuth errors
	ErrOAuthFailed   = errors.New("OAuth authentication failed")
	ErrInvalidState  = errors.New("invalid OAuth state parameter")
	ErrStateExpired  = errors.New("OAuth state expired")
	ErrStateNotFound = errors.New("OAuth state not found")

	// Configuration errors
	ErrMissingConfig = errors.New("missing required OAuth configuration")
	ErrInvalidConfig = errors.New("invalid OAuth configuration")

	// Authentication errors
	ErrAuthenticationRequired = errors.New("authentication required")
	ErrUnauthorized           = errors.New("unauthorized")

	// Access control errors
	ErrNotWorkspaceMember = errors.New("user is not a member of the Slack workspace")
)
