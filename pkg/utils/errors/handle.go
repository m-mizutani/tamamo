package errors

import (
	"context"

	"github.com/m-mizutani/ctxlog"
)

// Handle logs errors with context
func Handle(ctx context.Context, err error) {
	if err == nil {
		return
	}

	logger := ctxlog.From(ctx)
	logger.Error("error occurred", "error", err)
}
