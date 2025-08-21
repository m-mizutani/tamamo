package middleware

import (
	"context"

	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
)

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
