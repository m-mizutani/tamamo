package slack_test

import (
	"context"
	"errors"
	"os"
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

func (m *mockSlackClientForCache) IsWorkspaceMember(ctx context.Context, email string) (bool, error) {
	return true, nil
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

	// Immediate second call should use cache
	info2, err := cache.GetChannelInfo(ctx, channelID)
	gt.NoError(t, err)
	gt.Equal(t, info2.ID, channelID)
	gt.Equal(t, mockClient.getCallCount(), 1) // Should not increment

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Third call should fetch from client again due to expiration
	info3, err := cache.GetChannelInfo(ctx, channelID)
	gt.NoError(t, err)
	gt.Equal(t, info3.ID, channelID)
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

// TestChannelCache_RealSlackAPI tests actual Slack API integration
// This test runs only when TEST_SLACK_OAUTH_TOKEN and TEST_SLACK_CHANNEL are set
func TestChannelCache_RealSlackAPI(t *testing.T) {
	token := os.Getenv("TEST_SLACK_OAUTH_TOKEN")
	channelID := os.Getenv("TEST_SLACK_CHANNEL")

	if token == "" || channelID == "" {
		t.Skip("Skipping real Slack API test: TEST_SLACK_OAUTH_TOKEN or TEST_SLACK_CHANNEL not set")
	}

	ctx := context.Background()

	// Create real Slack client
	realClient, err := slackservice.New(token)
	gt.NoError(t, err)
	cache := slackservice.NewChannelCache(realClient, time.Hour)

	// First call should fetch from actual Slack API
	info1, err := cache.GetChannelInfo(ctx, channelID)
	gt.NoError(t, err)
	gt.NotNil(t, info1)
	gt.Equal(t, info1.ID, channelID)
	gt.True(t, len(info1.Name) > 0) // Should have a non-empty name

	// Verify channel type is valid
	validTypes := []slack.ChannelType{
		slack.ChannelTypePublic,
		slack.ChannelTypePrivate,
		slack.ChannelTypeIM,
		slack.ChannelTypeMPIM,
	}
	typeValid := false
	for _, vt := range validTypes {
		if info1.Type == vt {
			typeValid = true
			break
		}
	}
	gt.True(t, typeValid)

	// Second call should use cache (verify by timing)
	start := time.Now()
	info2, err := cache.GetChannelInfo(ctx, channelID)
	elapsed := time.Since(start)
	gt.NoError(t, err)
	gt.NotNil(t, info2)

	// Cached call should be very fast (less than 10ms)
	// Real API call would typically take 100ms+
	gt.True(t, elapsed < 10*time.Millisecond)

	// Verify cached data matches original
	gt.Equal(t, info2.ID, info1.ID)
	gt.Equal(t, info2.Name, info1.Name)
	gt.Equal(t, info2.Type, info1.Type)
	gt.Equal(t, info2.IsPrivate, info1.IsPrivate)

	// Test with invalid channel ID
	invalidInfo, err := cache.GetChannelInfo(ctx, "INVALID_CHANNEL")
	gt.Error(t, err)
	gt.Nil(t, invalidInfo)

	// Test cache expiration with short TTL
	shortCache := slackservice.NewChannelCache(realClient, 100*time.Millisecond)

	// First call
	info3, err := shortCache.GetChannelInfo(ctx, channelID)
	gt.NoError(t, err)
	gt.NotNil(t, info3)

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// This should make a new API call
	info4, err := shortCache.GetChannelInfo(ctx, channelID)
	gt.NoError(t, err)
	gt.NotNil(t, info4)
	gt.Equal(t, info4.ID, channelID)
}
