package graphql_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/controller/graphql"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

func TestQueryResolver_Thread_Success(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testThread := &slack.Thread{
		ID:        types.NewThreadID(ctx),
		TeamID:    "T123456",
		ChannelID: "C123456",
		ThreadTS:  "1234567890.123456",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{
		GetThreadFunc: func(ctx context.Context, id types.ThreadID) (*slack.Thread, error) {
			if id == testThread.ID {
				return testThread, nil
			}
			return nil, errors.New("thread not found")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	queryResolver := resolver.Query()

	// Execute test
	result, err := queryResolver.Thread(ctx, string(testThread.ID))

	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, result.ID, testThread.ID)
	gt.Equal(t, result.TeamID, testThread.TeamID)
	gt.Equal(t, result.ChannelID, testThread.ChannelID)
	gt.Equal(t, result.ThreadTS, testThread.ThreadTS)
}

func TestQueryResolver_Thread_InvalidID(t *testing.T) {
	ctx := context.Background()

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	queryResolver := resolver.Query()

	// Execute test with invalid ID
	result, err := queryResolver.Thread(ctx, "invalid-id")

	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
	gt.V(t, err.Error()).Equal("invalid thread ID")
}

func TestQueryResolver_Thread_RepositoryError(t *testing.T) {
	ctx := context.Background()
	testID := types.NewThreadID(ctx)

	// Setup mock to return error
	mockRepo := &mock.ThreadRepositoryMock{
		GetThreadFunc: func(ctx context.Context, id types.ThreadID) (*slack.Thread, error) {
			return nil, errors.New("repository error")
		},
	}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	queryResolver := resolver.Query()

	// Execute test
	result, err := queryResolver.Thread(ctx, string(testID))

	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
	// Error message contains wrapped error details
}


func TestThreadResolver_ID(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testThread := &slack.Thread{
		ID: types.NewThreadID(ctx),
	}

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	threadResolver := resolver.Thread()

	// Execute test
	result, err := threadResolver.ID(ctx, testThread)

	// Verify results
	gt.NoError(t, err)
	gt.Equal(t, result, string(testThread.ID))
}

func TestThreadResolver_CreatedAt(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	testThread := &slack.Thread{
		CreatedAt: testTime,
	}

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	threadResolver := resolver.Thread()

	// Execute test
	result, err := threadResolver.CreatedAt(ctx, testThread)

	// Verify results
	gt.NoError(t, err)
	gt.Equal(t, result, testTime.Format(time.RFC3339))
}

func TestThreadResolver_UpdatedAt(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	testThread := &slack.Thread{
		UpdatedAt: testTime,
	}

	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{}

	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	threadResolver := resolver.Thread()

	// Execute test
	result, err := threadResolver.UpdatedAt(ctx, testThread)

	// Verify results
	gt.NoError(t, err)
	gt.Equal(t, result, testTime.Format(time.RFC3339))
}

func TestQueryResolver_Thread_NotFound(t *testing.T) {
	ctx := context.Background()
	testID := types.NewThreadID(ctx)
	
	// Setup mock to return "not found" error
	mockRepo := &mock.ThreadRepositoryMock{
		GetThreadFunc: func(ctx context.Context, id types.ThreadID) (*slack.Thread, error) {
			return nil, errors.New("thread not found")
		},
	}
	
	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	queryResolver := resolver.Query()
	
	// Execute test
	result, err := queryResolver.Thread(ctx, string(testID))
	
	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
}

func TestQueryResolver_Threads_WithNilParameters(t *testing.T) {
	ctx := context.Background()
	
	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{
		ListThreadsFunc: func(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error) {
			// Return empty result for this test
			return []*slack.Thread{}, 0, nil
		},
	}
	
	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	queryResolver := resolver.Query()
	
	// Execute test with nil parameters
	result, err := queryResolver.Threads(ctx, nil, nil)
	
	// Verify results (should handle nil gracefully with defaults)
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, len(result.Threads), 0)
	gt.Equal(t, result.TotalCount, 0)
	
	// Verify mock was called with defaults
	calls := mockRepo.ListThreadsCalls()
	gt.Equal(t, len(calls), 1)
	gt.Equal(t, calls[0].Offset, 0)  // Default offset
	gt.Equal(t, calls[0].Limit, 50)  // Default limit
}

func TestQueryResolver_Threads_WithValidParameters(t *testing.T) {
	ctx := context.Background()
	
	// Setup test data
	testThreads := []*slack.Thread{
		{
			ID:        types.NewThreadID(ctx),
			TeamID:    "T123456",
			ChannelID: "C123456",
			ThreadTS:  "1234567890.123456",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{
		ListThreadsFunc: func(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error) {
			return testThreads, len(testThreads), nil
		},
	}
	
	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	queryResolver := resolver.Query()
	
	// Execute test with valid parameters
	offset := 5
	limit := 15
	result, err := queryResolver.Threads(ctx, &offset, &limit)
	
	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	gt.Equal(t, len(result.Threads), 1)
	gt.Equal(t, result.TotalCount, 1)
	
	// Verify mock was called with correct parameters
	calls := mockRepo.ListThreadsCalls()
	gt.Equal(t, len(calls), 1)
	gt.Equal(t, calls[0].Offset, 5)
	gt.Equal(t, calls[0].Limit, 15)
}

func TestQueryResolver_Threads_RepositoryError(t *testing.T) {
	ctx := context.Background()
	
	// Setup mock to return error
	mockRepo := &mock.ThreadRepositoryMock{
		ListThreadsFunc: func(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error) {
			return nil, 0, errors.New("repository error")
		},
	}
	
	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	queryResolver := resolver.Query()
	
	// Execute test
	offset := 0
	limit := 10
	result, err := queryResolver.Threads(ctx, &offset, &limit)
	
	// Verify results
	gt.Error(t, err)
	gt.V(t, result).Nil()
}

func TestQueryResolver_Threads_LimitCapping(t *testing.T) {
	ctx := context.Background()
	
	// Setup mock
	mockRepo := &mock.ThreadRepositoryMock{
		ListThreadsFunc: func(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error) {
			return []*slack.Thread{}, 0, nil
		},
	}
	
	// Create resolver
	resolver := graphql.NewResolver(mockRepo)
	queryResolver := resolver.Query()
	
	// Execute test with excessive limit
	offset := 0
	limit := 5000 // Should be capped to 1000
	result, err := queryResolver.Threads(ctx, &offset, &limit)
	
	// Verify results
	gt.NoError(t, err)
	gt.V(t, result).NotNil()
	
	// Verify mock was called with capped limit
	calls := mockRepo.ListThreadsCalls()
	gt.Equal(t, len(calls), 1)
	gt.Equal(t, calls[0].Offset, 0)
	gt.Equal(t, calls[0].Limit, 1000) // Should be capped
}
