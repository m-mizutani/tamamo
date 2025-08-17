package async_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/utils/async"
)

func TestWithSyncMode(t *testing.T) {
	t.Run("enables sync mode in context", func(t *testing.T) {
		ctx := context.Background()

		// Should not have sync mode initially
		executed := false
		done := make(chan bool, 1)

		async.Dispatch(ctx, func(ctx context.Context) error {
			executed = true
			done <- true
			return nil
		})

		// Should be async (not executed immediately)
		gt.False(t, executed)

		// Wait for async execution
		<-done
		gt.True(t, executed)

		// Now enable sync mode
		ctx = async.WithSyncMode(ctx)
		executed = false

		async.Dispatch(ctx, func(ctx context.Context) error {
			executed = true
			return nil
		})

		// Should be sync (executed immediately)
		gt.True(t, executed)
	})
}
