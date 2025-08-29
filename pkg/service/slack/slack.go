package slack

import (
	"context"
	"time"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	api "github.com/slack-go/slack"
)

// Service implements Slack operations
type Service struct {
	client       *api.Client
	botUserID    string
	authTestInfo *api.AuthTestResponse
	botInfo      *api.User // Store bot user information
}

// New creates a new Slack service
func New(token string) (*Service, error) {
	client := api.New(token)

	// Get bot user ID from auth test
	resp, err := client.AuthTest()
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get bot user ID")
	}

	service := &Service{
		client:       client,
		botUserID:    resp.UserID,
		authTestInfo: resp,
	}

	// Get bot user information for proper display (optional)
	botUser, err := client.GetUserInfo(resp.UserID)
	if err != nil {
		// Failed to get bot info - will fallback to basic message posting without custom icon/username
		service.botInfo = nil
	} else {
		service.botInfo = botUser
	}

	return service, nil
}

// GetAuthTestInfo returns the auth test information including team ID
func (s *Service) GetAuthTestInfo() (*api.AuthTestResponse, error) {
	if s.authTestInfo == nil {
		return nil, goerr.New("auth test info not available")
	}
	return s.authTestInfo, nil
}

// Ensure Service implements SlackClient interface
var _ interfaces.SlackClient = (*Service)(nil)

// PostMessage posts a message to a Slack channel/thread
func (s *Service) PostMessage(ctx context.Context, channelID, threadTS, text string) error {
	options := []api.MsgOption{
		api.MsgOptionText(text, false),
		api.MsgOptionAsUser(false), // Explicitly send as bot, not as user
	}

	// Set bot username and icon for consistent display in threads
	if s.botInfo != nil {
		options = append(options, api.MsgOptionUsername(s.botInfo.Name))
		if s.botInfo.Profile.Image72 != "" {
			options = append(options, api.MsgOptionIconURL(s.botInfo.Profile.Image72))
		}
	}

	// Always reply in thread if threadTS is provided
	if threadTS != "" {
		options = append(options, api.MsgOptionTS(threadTS))
	}

	channelID, timestamp, err := s.client.PostMessageContext(
		ctx,
		channelID,
		options...,
	)
	if err != nil {
		return goerr.Wrap(err, "failed to post message to slack", goerr.V("channel", channelID), goerr.V("thread", threadTS))
	}

	logger := ctxlog.From(ctx)
	logFields := []any{
		"channel", channelID,
		"timestamp", timestamp,
		"thread", threadTS,
	}

	if s.botInfo != nil {
		logFields = append(logFields, "bot_name", s.botInfo.Name)
		logFields = append(logFields, "bot_icon", s.botInfo.Profile.Image72)
	} else {
		logFields = append(logFields, "bot_info", "nil")
	}

	logger.Debug("posted message to slack", logFields...)

	return nil
}

// PostMessageWithOptions posts a message to a Slack channel/thread with custom display options
func (s *Service) PostMessageWithOptions(ctx context.Context, channelID, threadTS, text string, options *interfaces.SlackMessageOptions) error {
	msgOptions := []api.MsgOption{
		api.MsgOptionText(text, false),
		api.MsgOptionAsUser(false), // Explicitly send as bot, not as user
	}

	// Use custom username and icon if provided
	if options != nil {
		if options.Username != "" {
			msgOptions = append(msgOptions, api.MsgOptionUsername(options.Username))
		}
		if options.IconEmoji != "" {
			msgOptions = append(msgOptions, api.MsgOptionIconEmoji(options.IconEmoji))
		} else if options.IconURL != "" {
			msgOptions = append(msgOptions, api.MsgOptionIconURL(options.IconURL))
		}
	} else {
		// Fallback to bot info if no custom options provided
		if s.botInfo != nil {
			msgOptions = append(msgOptions, api.MsgOptionUsername(s.botInfo.Name))
			if s.botInfo.Profile.Image72 != "" {
				msgOptions = append(msgOptions, api.MsgOptionIconURL(s.botInfo.Profile.Image72))
			}
		}
	}

	// Always reply in thread if threadTS is provided
	if threadTS != "" {
		msgOptions = append(msgOptions, api.MsgOptionTS(threadTS))
	}

	channelID, timestamp, err := s.client.PostMessageContext(
		ctx,
		channelID,
		msgOptions...,
	)
	if err != nil {
		return goerr.Wrap(err, "failed to post message to slack", goerr.V("channel", channelID), goerr.V("thread", threadTS))
	}

	logger := ctxlog.From(ctx)
	logFields := []any{
		"channel", channelID,
		"timestamp", timestamp,
		"thread", threadTS,
	}

	if options != nil {
		logFields = append(logFields, "custom_username", options.Username)
		logFields = append(logFields, "custom_icon", options.IconURL)
	} else if s.botInfo != nil {
		logFields = append(logFields, "bot_name", s.botInfo.Name)
		logFields = append(logFields, "bot_icon", s.botInfo.Profile.Image72)
	} else {
		logFields = append(logFields, "bot_info", "nil")
	}

	logger.Debug("posted message to slack with options", logFields...)

	return nil
}

// IsBotUser checks if the given user ID is the bot user
func (s *Service) IsBotUser(userID string) bool {
	return s.botUserID == userID
}

// GetUserProfile retrieves user profile information from Slack
func (s *Service) GetUserProfile(ctx context.Context, userID string) (*interfaces.SlackUserProfile, error) {
	user, err := s.client.GetUserInfo(userID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get user info from Slack", goerr.V("user_id", userID))
	}

	return &interfaces.SlackUserProfile{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.Profile.DisplayName,
		Email:       user.Profile.Email,
		Profile: struct {
			Image24   string `json:"image_24"`
			Image32   string `json:"image_32"`
			Image48   string `json:"image_48"`
			Image72   string `json:"image_72"`
			Image192  string `json:"image_192"`
			Image512  string `json:"image_512"`
			ImageOrig string `json:"image_original"`
		}{
			Image24:   user.Profile.Image24,
			Image32:   user.Profile.Image32,
			Image48:   user.Profile.Image48,
			Image72:   user.Profile.Image72,
			Image192:  user.Profile.Image192,
			Image512:  user.Profile.Image512,
			ImageOrig: user.Profile.ImageOriginal,
		},
	}, nil
}

// GetUserInfo retrieves user information from Slack
func (s *Service) GetUserInfo(ctx context.Context, userID string) (*interfaces.SlackUserInfo, error) {
	user, err := s.client.GetUserInfo(userID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get user info from Slack", goerr.V("user_id", userID))
	}

	return &interfaces.SlackUserInfo{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.Profile.DisplayName,
		RealName:    user.Profile.RealName,
	}, nil
}

// GetBotInfo retrieves bot information from Slack
func (s *Service) GetBotInfo(ctx context.Context, botID string) (*interfaces.SlackBotInfo, error) {
	bot, err := s.client.GetBotInfo(api.GetBotInfoParameters{Bot: botID})
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get bot info from Slack", goerr.V("bot_id", botID))
	}

	return &interfaces.SlackBotInfo{
		ID:   bot.ID,
		Name: bot.Name,
	}, nil
}

// GetChannelInfo retrieves channel information from Slack API
func (s *Service) GetChannelInfo(ctx context.Context, channelID string) (*slack.ChannelInfo, error) {
	if channelID == "" {
		return nil, goerr.New("channelID cannot be empty")
	}

	// Use conversations.info API to get channel information
	channel, err := s.client.GetConversationInfo(&api.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get channel info from Slack",
			goerr.V("channel_id", channelID))
	}

	// Determine channel type from Slack API response
	channelType := slack.DetermineChannelType(
		channel.IsChannel,
		channel.IsGroup,
		channel.IsIM,
		channel.IsMpIM,
	)

	return &slack.ChannelInfo{
		ID:        channel.ID,
		Name:      channel.Name,
		Type:      channelType,
		IsPrivate: channel.IsPrivate,
		UpdatedAt: time.Now(),
	}, nil
}

// ThreadService provides thread-specific operations
type ThreadService struct {
	service   *Service
	channelID string
	threadTS  string
}

// NewThread creates a new thread service
func (s *Service) NewThread(channelID, threadTS string) *ThreadService {
	return &ThreadService{
		service:   s,
		channelID: channelID,
		threadTS:  threadTS,
	}
}

// Reply posts a message to the thread
func (t *ThreadService) Reply(ctx context.Context, text string) error {
	return t.service.PostMessage(ctx, t.channelID, t.threadTS, text)
}
