package async

import "context"

type contextKey string

const (
	syncModeKey contextKey = "async-sync-mode"
)

// WithSyncMode returns a new context with sync mode enabled
// When sync mode is enabled, Dispatch will execute handlers synchronously
// This is useful for testing
func WithSyncMode(ctx context.Context) context.Context {
	return context.WithValue(ctx, syncModeKey, true)
}

// isSyncMode checks if sync mode is enabled in the context
func isSyncMode(ctx context.Context) bool {
	if v, ok := ctx.Value(syncModeKey).(bool); ok {
		return v
	}
	return false
}
