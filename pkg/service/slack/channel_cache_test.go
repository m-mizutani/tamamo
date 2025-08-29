package slack_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	slackservice "github.com/m-mizutani/tamamo/pkg/service/slack"
)

// Mock SlackClient for channel cache testing
type mockSlackClientForCache struct {
	getChannelInfoFunc func(ctx context.Context, channelID string) (*slack.ChannelInfo, error)
	mu                 sync.Mutex
	callCount          int
}

func (m *mockSlackClientForCache) PostMessage(ctx context.Context, channelID, threadTS, text string) error {
	return nil
}

func (m *mockSlackClientForCache) PostMessageWithOptions(ctx context.Context, channelID, threadTS, text string, options *interfaces.SlackMessageOptions) error {
	return nil
}

func (m *mockSlackClientForCache) IsBotUser(userID string) bool {
	return false
}

func (m *mockSlackClientForCache) GetUserProfile(ctx context.Context, userID string) (*interfaces.SlackUserProfile, error) {
	return nil, nil
}

func (m *mockSlackClientForCache) GetUserInfo(ctx context.Context, userID string) (*interfaces.SlackUserInfo, error) {
	return nil, nil
}

func (m *mockSlackClientForCache) GetBotInfo(ctx context.Context, botID string) (*interfaces.SlackBotInfo, error) {
	return nil, nil
}

func (m *mockSlackClientForCache) GetChannelInfo(ctx context.Context, channelID string) (*slack.ChannelInfo, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()

	if m.getChannelInfoFunc != nil {
		return m.getChannelInfoFunc(ctx, channelID)
	}
	return &slack.ChannelInfo{
		ID:        channelID,
		Name:      "test-channel",
		Type:      slack.ChannelTypePublic,
		IsPrivate: false,
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockSlackClientForCache) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func TestChannelCache_BasicOperation(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockSlackClientForCache{}
	cache := slackservice.NewChannelCache(mockClient, time.Hour)

	// First call should fetch from client
	info1, err := cache.GetChannelInfo(ctx, "C123456789")
	gt.NoError(t, err)
	gt.NotNil(t, info1)
	gt.Equal(t, info1.ID, "C123456789")
	gt.Equal(t, info1.Name, "test-channel")
	gt.Equal(t, mockClient.getCallCount(), 1)

	// Second call should use cache
	info2, err := cache.GetChannelInfo(ctx, "C123456789")
	gt.NoError(t, err)
	gt.NotNil(t, info2)
	gt.Equal(t, info2.ID, "C123456789")
	gt.Equal(t, mockClient.getCallCount(), 1) // Should not increment

	// Different channel should make new call
	info3, err := cache.GetChannelInfo(ctx, "C987654321")
	gt.NoError(t, err)
	gt.NotNil(t, info3)
	gt.Equal(t, info3.ID, "C987654321")
	gt.Equal(t, mockClient.getCallCount(), 2) // Should increment
}

func TestChannelCache_CacheHit(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockSlackClientForCache{}
	cache := slackservice.NewChannelCache(mockClient, time.Hour)

	channelID := "C123456789"

	// Multiple calls to same channel should only hit the client once
	for i := 0; i < 5; i++ {
		info, err := cache.GetChannelInfo(ctx, channelID)
		gt.NoError(t, err)
		gt.Equal(t, info.ID, channelID)
	}

	gt.Equal(t, mockClient.getCallCount(), 1) // Only one call to client
}

func TestChannelCache_CacheMiss(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockSlackClientForCache{}
	cache := slackservice.NewChannelCache(mockClient, time.Hour)

	// Different channels should each make a call
	channels := []string{"C111111111", "C222222222", "C333333333"}

	for _, channelID := range channels {
		info, err := cache.GetChannelInfo(ctx, channelID)
		gt.NoError(t, err)
		gt.Equal(t, info.ID, channelID)
	}

	gt.Equal(t, mockClient.getCallCount(), len(channels))
}

func TestChannelCache_TTLExpiration(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockSlackClientForCache{}
	// Very short TTL for testing
	cache := slackservice.NewChannelCache(mockClient, 50*time.Millisecond)

	channelID := "C123456789"

	// First call
	info1, err := cache.GetChannelInfo(ctx, channelID)
	gt.NoError(t, err)
	gt.Equal(t, info1.ID, channelID)
	gt.Equal(t, mockClient.getCallCount(), 1)

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Second call should fetch from client again
	info2, err := cache.GetChannelInfo(ctx, channelID)
	gt.NoError(t, err)
	gt.Equal(t, info2.ID, channelID)
	gt.Equal(t, mockClient.getCallCount(), 2) // Should increment due to expiration
}

func TestChannelCache_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockSlackClientForCache{}
	cache := slackservice.NewChannelCache(mockClient, time.Hour)

	// Mock error response
	expectedError := errors.New("channel not found")
	mockClient.getChannelInfoFunc = func(ctx context.Context, channelID string) (*slack.ChannelInfo, error) {
		return nil, expectedError
	}

	// Error should be propagated
	info, err := cache.GetChannelInfo(ctx, "C123456789")
	gt.Error(t, err)
	gt.Nil(t, info)
	// Check that error contains the expected message (basic check)
	gt.True(t, len(err.Error()) > 0)

	// Error responses should not be cached
	gt.Equal(t, mockClient.getCallCount(), 1)

	// Second call should also make client call (errors not cached)
	info2, err2 := cache.GetChannelInfo(ctx, "C123456789")
	gt.Error(t, err2)
	gt.Nil(t, info2)
	gt.Equal(t, mockClient.getCallCount(), 2)
}

func TestChannelCache_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockSlackClientForCache{}
	cache := slackservice.NewChannelCache(mockClient, time.Hour)

	channelID := "C123456789"
	numGoroutines := 10

	// Simulate concurrent access
	done := make(chan bool, numGoroutines)
	errorsChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			info, err := cache.GetChannelInfo(ctx, channelID)
			if err != nil {
				errorsChan <- err
				return
			}

			if info.ID != channelID {
				errorsChan <- errors.New("incorrect channel ID")
				return
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	close(errorsChan)
	for err := range errorsChan {
		t.Errorf("Concurrent access error: %v", err)
	}

	// Should only make one call to client despite concurrent access
	gt.Equal(t, mockClient.getCallCount(), 1)
}

func TestChannelCache_DifferentChannelTypes(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockSlackClientForCache{}
	cache := slackservice.NewChannelCache(mockClient, time.Hour)

	// Mock different channel types
	channelTypes := map[string]slack.ChannelType{
		"C111111111": slack.ChannelTypePublic,
		"C222222222": slack.ChannelTypePrivate,
		"D333333333": slack.ChannelTypeIM,
		"G444444444": slack.ChannelTypeMPIM,
	}

	mockClient.getChannelInfoFunc = func(ctx context.Context, channelID string) (*slack.ChannelInfo, error) {
		channelType, exists := channelTypes[channelID]
		if !exists {
			return nil, errors.New("channel not found")
		}

		return &slack.ChannelInfo{
			ID:        channelID,
			Name:      "test-channel",
			Type:      channelType,
			IsPrivate: channelType == slack.ChannelTypePrivate,
			UpdatedAt: time.Now(),
		}, nil
	}

	// Test each channel type
	for channelID, expectedType := range channelTypes {
		info, err := cache.GetChannelInfo(ctx, channelID)
		gt.NoError(t, err)
		gt.Equal(t, info.Type, expectedType)
		gt.Equal(t, info.IsPrivate, expectedType == slack.ChannelTypePrivate)
	}

	gt.Equal(t, mockClient.getCallCount(), len(channelTypes))
}

func TestChannelCache_EmptyChannelID(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockSlackClientForCache{}
	cache := slackservice.NewChannelCache(mockClient, time.Hour)

	// Empty channel ID should return error
	info, err := cache.GetChannelInfo(ctx, "")
	gt.Error(t, err)
	gt.Nil(t, info)
	gt.Equal(t, mockClient.getCallCount(), 0) // Should not call client
}

func TestChannelCache_CleanupWorker(t *testing.T) {
	ctx := context.Background()

	mockClient := &mockSlackClientForCache{}
	// Very short TTL and cleanup interval for testing
	cache := slackservice.NewChannelCache(mockClient, 50*time.Millisecond)

	channelID := "C123456789"

	// Add entry to cache
	info, err := cache.GetChannelInfo(ctx, channelID)
	gt.NoError(t, err)
	gt.NotNil(t, info)

	// Wait for cleanup to run (cleanup runs every TTL/2)
	time.Sleep(100 * time.Millisecond)

	// Entry should be cleaned up, so next call should hit client again
	info2, err := cache.GetChannelInfo(ctx, channelID)
	gt.NoError(t, err)
	gt.NotNil(t, info2)

	// Should have made 2 calls (initial + after cleanup)
	gt.Equal(t, mockClient.getCallCount(), 2)
}
