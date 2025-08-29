package interfaces

import (
	"context"

	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
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

// SlackUserInfo represents user information from Slack API
type SlackUserInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	RealName    string `json:"real_name"`
}

// SlackBotInfo represents bot information from Slack API
type SlackBotInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SlackMessageOptions holds optional display settings for Slack messages
type SlackMessageOptions struct {
	Username  string // Custom username to display
	IconURL   string // Custom icon URL to display
	IconEmoji string // Custom emoji to display (alternative to IconURL)
}

type SlackClient interface {
	PostMessage(ctx context.Context, channelID, threadTS, text string) error
	PostMessageWithOptions(ctx context.Context, channelID, threadTS, text string, options *SlackMessageOptions) error
	IsBotUser(userID string) bool
	GetUserProfile(ctx context.Context, userID string) (*SlackUserProfile, error)
	GetUserInfo(ctx context.Context, userID string) (*SlackUserInfo, error)
	GetBotInfo(ctx context.Context, botID string) (*SlackBotInfo, error)
	GetChannelInfo(ctx context.Context, channelID string) (*slack.ChannelInfo, error)
}

// UserAvatarService manages user avatar data retrieval
type UserAvatarService interface {
	GetAvatarData(ctx context.Context, slackID string, size int) ([]byte, error)
	InvalidateCache(ctx context.Context, slackID string) error
}
