package usecase

import (
	"context"
	"time"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/slack-go/slack/slackevents"
)

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