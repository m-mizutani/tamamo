package interfaces

import (
	"context"
)

// SlackUserProfile represents a user's profile information from Slack
type SlackUserProfile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Profile     struct {
		Image24   string `json:"image_24"`
		Image32   string `json:"image_32"`
		Image48   string `json:"image_48"`
		Image72   string `json:"image_72"`
		Image192  string `json:"image_192"`
		Image512  string `json:"image_512"`
		ImageOrig string `json:"image_original"`
	} `json:"profile"`
}

type SlackClient interface {
	PostMessage(ctx context.Context, channelID, threadTS, text string) error
	IsBotUser(userID string) bool
	GetUserProfile(ctx context.Context, userID string) (*SlackUserProfile, error)
}

// UserAvatarService manages user avatar data retrieval
type UserAvatarService interface {
	GetAvatarData(ctx context.Context, slackID string, size int) ([]byte, error)
	InvalidateCache(ctx context.Context, slackID string) error
}
