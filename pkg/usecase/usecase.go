package usecase

import (
	"context"
	"time"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/storage"
	"github.com/m-mizutani/tamamo/pkg/service/llm"
	slackservice "github.com/m-mizutani/tamamo/pkg/service/slack"
	"github.com/slack-go/slack/slackevents"
)

// Slack holds all use cases
type Slack struct {
	slackClient         interfaces.SlackClient
	repository          interfaces.ThreadRepository
	agentRepository     interfaces.AgentRepository
	agentImageRepo      interfaces.AgentImageRepository
	slackMessageLogRepo interfaces.SlackMessageLogRepository
	storageRepo         *storage.Client
	llmClient           gollem.LLMClient // Deprecated: use llmFactory instead
	llmModel            string           // Deprecated: use llmFactory instead
	llmFactory          *llm.Factory
	serverBaseURL       string // Base URL for constructing image URLs
	channelCache        *slackservice.ChannelCache
}

// SlackOption is a functional option for Slack
type SlackOption func(*Slack)

// WithSlackClient sets the Slack client
func WithSlackClient(client interfaces.SlackClient) SlackOption {
	return func(uc *Slack) {
		uc.slackClient = client
	}
}

// WithRepository sets the repository
func WithRepository(repo interfaces.ThreadRepository) SlackOption {
	return func(uc *Slack) {
		uc.repository = repo
	}
}

// WithAgentRepository sets the agent repository
func WithAgentRepository(repo interfaces.AgentRepository) SlackOption {
	return func(uc *Slack) {
		uc.agentRepository = repo
	}
}

// WithAgentImageRepository sets the agent image repository
func WithAgentImageRepository(repo interfaces.AgentImageRepository) SlackOption {
	return func(uc *Slack) {
		uc.agentImageRepo = repo
	}
}

// WithStorageRepository sets the storage repository
func WithStorageRepository(repo *storage.Client) SlackOption {
	return func(uc *Slack) {
		uc.storageRepo = repo
	}
}

// WithLLMClient sets the LLM client (deprecated: use WithLLMFactory)
func WithLLMClient(client gollem.LLMClient) SlackOption {
	return func(uc *Slack) {
		uc.llmClient = client
	}
}

// WithLLMModel sets the LLM model (deprecated: use WithLLMFactory)
func WithLLMModel(model string) SlackOption {
	return func(uc *Slack) {
		uc.llmModel = model
	}
}

// WithLLMFactory sets the LLM factory for multi-provider support
func WithLLMFactory(factory *llm.Factory) SlackOption {
	return func(uc *Slack) {
		uc.llmFactory = factory
		// Also set default client for backward compatibility
		if factory != nil && factory.GetDefaultClient() != nil {
			uc.llmClient = factory.GetDefaultClient()
		}
	}
}

// WithGeminiClient is deprecated. Use WithLLMClient instead.
func WithGeminiClient(client gollem.LLMClient) SlackOption {
	return WithLLMClient(client)
}

// WithGeminiModel is deprecated. Use WithLLMModel instead.
func WithGeminiModel(model string) SlackOption {
	return WithLLMModel(model)
}

// WithSlackMessageLogRepository sets the Slack message log repository
func WithSlackMessageLogRepository(repo interfaces.SlackMessageLogRepository) SlackOption {
	return func(uc *Slack) {
		uc.slackMessageLogRepo = repo
	}
}

// WithServerBaseURL sets the server base URL for constructing image URLs
func WithServerBaseURL(baseURL string) SlackOption {
	return func(uc *Slack) {
		uc.serverBaseURL = baseURL
	}
}

// New creates a new Slack instance
func New(opts ...SlackOption) *Slack {
	uc := &Slack{}
	for _, opt := range opts {
		opt(uc)
	}

	// Initialize channel cache if slack client is available
	if uc.slackClient != nil {
		uc.channelCache = slackservice.NewChannelCache(uc.slackClient, time.Hour)
	}

	return uc
}

// LogSlackMessage logs a Slack message to the repository
func (uc *Slack) LogSlackMessage(ctx context.Context, event *slackevents.MessageEvent, teamID string) error {
	return uc.LogSlackMessageWithTeam(ctx, event, teamID)
}

// LogSlackMessageWithTeam logs a Slack message with team information
func (uc *Slack) LogSlackMessageWithTeam(ctx context.Context, event *slackevents.MessageEvent, teamID string) error {
	if uc.slackMessageLogRepo == nil || uc.channelCache == nil {
		// Message logging not configured, skip silently
		return nil
	}

	logger := ctxlog.From(ctx)

	// Get channel information (with caching)
	channelInfo, err := uc.channelCache.GetChannelInfo(ctx, event.Channel)
	if err != nil {
		// Log error but continue with partial data (best effort)
		logger.Warn("failed to get channel info, using defaults",
			"channel_id", event.Channel,
			"error", err)

		// Create default channel info
		channelInfo = &slack.ChannelInfo{
			ID:        event.Channel,
			Name:      event.Channel,           // Fallback to channel ID
			Type:      slack.ChannelTypePublic, // Default assumption
			IsPrivate: false,
			UpdatedAt: time.Now(),
		}
	}

	// Determine message type
	messageType := slack.DetermineMessageType(event.User, event.BotID)

	// Get user information if available
	var userName string
	if event.User != "" && uc.slackClient != nil {
		userInfo, err := uc.slackClient.GetUserInfo(ctx, event.User)
		if err != nil {
			// Log error but continue (best effort)
			logger.Warn("failed to get user info",
				"user_id", event.User,
				"error", err)
		} else {
			userName = userInfo.Name
		}
	}

	// Convert file attachments from the message
	var attachments []slack.Attachment
	if event.Message != nil && event.Message.Files != nil {
		for _, file := range event.Message.Files {
			attachments = append(attachments, slack.Attachment{
				ID:       file.ID,
				Name:     file.Name,
				Mimetype: file.Mimetype,
				FileType: file.Filetype,
				URL:      file.URLPrivate,
			})
		}
	}

	// Create SlackMessageLog entry
	messageLog := &slack.SlackMessageLog{
		ID:          types.NewMessageID(ctx),
		TeamID:      teamID,
		ChannelID:   event.Channel,
		ChannelName: channelInfo.Name,
		ChannelType: channelInfo.Type,
		UserID:      event.User,
		UserName:    userName,
		BotID:       event.BotID,
		MessageType: messageType,
		Text:        event.Text,
		Timestamp:   event.TimeStamp,
		ThreadTS:    event.ThreadTimeStamp,
		Attachments: attachments,
		CreatedAt:   time.Now(),
	}

	// Store in repository
	if err := uc.slackMessageLogRepo.PutSlackMessageLog(ctx, messageLog); err != nil {
		// The error will be handled and logged by the async dispatcher
		return goerr.Wrap(err, "failed to store slack message log",
			goerr.V("message_id", messageLog.ID),
			goerr.V("channel_id", event.Channel),
			goerr.V("user_id", event.User))
	}

	logger.Debug("successfully logged slack message",
		"message_id", messageLog.ID,
		"channel_id", event.Channel,
		"channel_name", channelInfo.Name,
		"channel_type", channelInfo.Type,
		"user_id", event.User,
		"message_type", messageType)

	return nil
}

// LogSlackAppMentionMessage logs a Slack app mention message
func (uc *Slack) LogSlackAppMentionMessage(ctx context.Context, event *slackevents.AppMentionEvent, teamID string) error {
	// Convert AppMentionEvent to MessageEvent for reuse
	messageEvent := &slackevents.MessageEvent{
		Type:            "message",
		User:            event.User,
		Text:            event.Text,
		TimeStamp:       event.TimeStamp,
		ThreadTimeStamp: event.ThreadTimeStamp,
		Channel:         event.Channel,
		// Note: Team and Files info not available in MessageEvent struct
	}

	return uc.LogSlackMessage(ctx, messageEvent, teamID)
}

// GetMessageLogs retrieves message logs with filtering (primarily for channel and time period)
func (uc *Slack) GetMessageLogs(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error) {
	if uc.slackMessageLogRepo == nil {
		return []*slack.SlackMessageLog{}, nil
	}
	return uc.slackMessageLogRepo.GetSlackMessageLogs(ctx, channel, from, to, limit, offset)
}

// Ensure Slack implements required interfaces
var _ interfaces.SlackEventUseCases = (*Slack)(nil)
