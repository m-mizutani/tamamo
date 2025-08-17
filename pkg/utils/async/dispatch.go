package async

import (
	"context"
	"runtime/debug"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/utils/errors"
)

// Dispatch executes a handler function asynchronously with proper context and panic recovery
// If sync mode is enabled in the context, the handler will be executed synchronously
func Dispatch(ctx context.Context, handler func(ctx context.Context) error) {
	// Check if sync mode is enabled (for testing)
	if isSyncMode(ctx) {
		// Execute synchronously
		if err := handler(ctx); err != nil {
			errors.Handle(ctx, err)
		}
		return
	}

	// Normal async execution
	newCtx := newBackgroundContext(ctx)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				err := goerr.New("panic in async handler",
					goerr.V("recover", r),
					goerr.V("stack", string(stack)),
				)
				errors.Handle(newCtx, err)
			}
		}()

		if err := handler(newCtx); err != nil {
			errors.Handle(newCtx, err)
		}
	}()
}

// newBackgroundContext creates a new background context preserving important values
func newBackgroundContext(ctx context.Context) context.Context {
	newCtx := context.Background()

	// Preserve logger from the original context
	logger := ctxlog.From(ctx)
	newCtx = ctxlog.With(newCtx, logger)

	return newCtx
}
