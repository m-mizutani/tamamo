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
	client    *api.Client
	botUserID string
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
		client:    client,
		botUserID: resp.UserID,
	}, nil
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
