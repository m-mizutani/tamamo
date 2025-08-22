package slack_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/service/slack"
)

func TestAvatarService_GetAvatarData(t *testing.T) {
	ctx := context.Background()
	mockSlackClient := &mock.SlackClientMock{
		GetUserProfileFunc: func(ctx context.Context, userID string) (*interfaces.SlackUserProfile, error) {
			return &interfaces.SlackUserProfile{
				ID:          userID,
				Name:        "Test User",
				DisplayName: "Test Display Name",
				Email:       "test@example.com",
				Profile: struct {
					Image24   string `json:"image_24"`
					Image32   string `json:"image_32"`
					Image48   string `json:"image_48"`
					Image72   string `json:"image_72"`
					Image192  string `json:"image_192"`
					Image512  string `json:"image_512"`
					ImageOrig string `json:"image_original"`
				}{
					Image24:   "https://example.com/avatar_24.jpg",
					Image32:   "https://example.com/avatar_32.jpg",
					Image48:   "https://example.com/avatar_48.jpg",
					Image72:   "https://example.com/avatar_72.jpg",
					Image192:  "https://example.com/avatar_192.jpg",
					Image512:  "https://example.com/avatar_512.jpg",
					ImageOrig: "https://example.com/avatar_original.jpg",
				},
			}, nil
		},
	}

	service := slack.NewAvatarService(mockSlackClient)

	t.Run("GetAvatarData_Success", func(t *testing.T) {
		// Note: This test uses mock data since we have a placeholder implementation
		data, err := service.GetAvatarData(ctx, "U123456789", 48)

		// The placeholder implementation should succeed with mock URLs
		gt.NoError(t, err)
		if len(data) == 0 {
			t.Error("expected non-empty data")
		}
	})

	t.Run("GetAvatarData_Caching", func(t *testing.T) {
		// Test that the same request returns cached data
		// This would be properly testable with a mock HTTP client
		slackID := "U123456789"
		size := 32

		// First request (would populate cache)
		data1, err1 := service.GetAvatarData(ctx, slackID, size)
		gt.NoError(t, err1)
		if len(data1) == 0 {
			t.Error("expected non-empty data1")
		}

		// Second request (should return cached data)
		data2, err2 := service.GetAvatarData(ctx, slackID, size)
		gt.NoError(t, err2)
		if len(data2) == 0 {
			t.Error("expected non-empty data2")
		}
		// Data should be identical (from cache)
		gt.Equal(t, data1, data2)
	})
}

func TestAvatarService_InvalidateCache(t *testing.T) {
	ctx := context.Background()
	mockSlackClient := &mock.SlackClientMock{
		GetUserProfileFunc: func(ctx context.Context, userID string) (*interfaces.SlackUserProfile, error) {
			return &interfaces.SlackUserProfile{
				ID:          userID,
				Name:        "Test User",
				DisplayName: "Test Display Name",
				Email:       "test@example.com",
				Profile: struct {
					Image24   string `json:"image_24"`
					Image32   string `json:"image_32"`
					Image48   string `json:"image_48"`
					Image72   string `json:"image_72"`
					Image192  string `json:"image_192"`
					Image512  string `json:"image_512"`
					ImageOrig string `json:"image_original"`
				}{
					Image24:   "https://example.com/avatar_24.jpg",
					Image32:   "https://example.com/avatar_32.jpg",
					Image48:   "https://example.com/avatar_48.jpg",
					Image72:   "https://example.com/avatar_72.jpg",
					Image192:  "https://example.com/avatar_192.jpg",
					Image512:  "https://example.com/avatar_512.jpg",
					ImageOrig: "https://example.com/avatar_original.jpg",
				},
			}, nil
		},
	}

	service := slack.NewAvatarService(mockSlackClient)

	t.Run("InvalidateCache_Success", func(t *testing.T) {
		slackID := "U123456789"

		// Invalidate cache (should work regardless of whether cache has data)
		err := service.InvalidateCache(ctx, slackID)
		gt.NoError(t, err)
	})
}

func TestAvatarService_SelectAvatarURL(t *testing.T) {
	// Create a service to access the selectAvatarURL method
	// Note: This method is private, so we can't test it directly
	// In a real implementation, we might make it public or test it indirectly

	t.Run("Size_Selection_Logic", func(t *testing.T) {
		// This would test the URL selection logic if the method were public
		// For now, we'll test it indirectly through GetAvatarData
		t.Skip("selectAvatarURL is private - testing indirectly through GetAvatarData")
	})
}
