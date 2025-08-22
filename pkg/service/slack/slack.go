package slack

import (
	"context"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	api "github.com/slack-go/slack"
)

// Service implements Slack operations
type Service struct {
	client       *api.Client
	botUserID    string
	authTestInfo *api.AuthTestResponse
}

// New creates a new Slack service
func New(token string) (*Service, error) {
	client := api.New(token)

	// Get bot user ID from auth test
	resp, err := client.AuthTest()
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get bot user ID")
	}

	return &Service{
		client:       client,
		botUserID:    resp.UserID,
		authTestInfo: resp,
	}, nil
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

	ctxlog.From(ctx).Debug("posted message to slack",
		"channel", channelID,
		"timestamp", timestamp,
		"thread", threadTS,
	)

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
