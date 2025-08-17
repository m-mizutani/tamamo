package async_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/utils/async"
)

func TestDispatch(t *testing.T) {
	t.Run("executes synchronously when sync mode is enabled", func(t *testing.T) {
		ctx := async.WithSyncMode(context.Background())
		executed := false

		async.Dispatch(ctx, func(ctx context.Context) error {
			executed = true
			return nil
		})

		// Should execute immediately without waiting
		gt.True(t, executed)
	})

	t.Run("executes handler asynchronously", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		executed := false

		async.Dispatch(context.Background(), func(ctx context.Context) error {
			defer wg.Done()
			executed = true
			return nil
		})

		wg.Wait()
		gt.True(t, executed)
	})

	t.Run("recovers from panic", func(t *testing.T) {
		done := make(chan bool, 1)

		async.Dispatch(context.Background(), func(ctx context.Context) error {
			defer func() {
				done <- true
			}()
			panic("test panic")
		})

		select {
		case <-done:
			// Successfully recovered from panic
		case <-time.After(time.Second):
			t.Fatal("handler did not complete after panic")
		}
	})

	t.Run("handles errors without panic", func(t *testing.T) {
		done := make(chan bool, 1)

		async.Dispatch(context.Background(), func(ctx context.Context) error {
			defer func() {
				done <- true
			}()
			return goerr.New("test error")
		})

		select {
		case <-done:
			// Successfully handled error
		case <-time.After(time.Second):
			t.Fatal("handler did not complete")
		}
	})
}
