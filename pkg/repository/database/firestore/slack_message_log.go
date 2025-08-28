package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"google.golang.org/api/iterator"
)

const (
	collectionSlackChannels = "log_slack_channels"
	subCollectionMessages   = "messages"
	docChannelInfo         = "info"
)

// PutSlackMessageLog stores a Slack message log entry in Firestore using hierarchical structure
func (c *Client) PutSlackMessageLog(ctx context.Context, messageLog *slack.SlackMessageLog) error {
	if messageLog == nil {
		return goerr.New("messageLog cannot be nil")
	}
	if messageLog.ID == "" {
		return goerr.New("messageLog.ID cannot be empty")
	}
	if messageLog.ChannelID == "" {
		return goerr.New("messageLog.ChannelID cannot be empty")
	}

	// Store message in subcollection: log_slack_channels/{channelId}/messages/{messageId}
	docRef := c.client.Collection(collectionSlackChannels).
		Doc(messageLog.ChannelID).
		Collection(subCollectionMessages).
		Doc(string(messageLog.ID))

	_, err := docRef.Set(ctx, messageLog)
	if err != nil {
		return goerr.Wrap(err, "failed to store slack message log",
			goerr.V("message_id", messageLog.ID),
			goerr.V("channel_id", messageLog.ChannelID),
			goerr.V("user_id", messageLog.UserID))
	}

	// Store/update channel info if available
	if messageLog.ChannelName != "" {
		channelInfo := &slack.ChannelInfo{
			ID:        messageLog.ChannelID,
			Name:      messageLog.ChannelName,
			Type:      messageLog.ChannelType,
			IsPrivate: messageLog.ChannelType == slack.ChannelTypePrivate,
			UpdatedAt: messageLog.CreatedAt,
		}

		// Store channel info as a document in the same channel document
		infoDocRef := c.client.Collection(collectionSlackChannels).
			Doc(messageLog.ChannelID)

		_, err := infoDocRef.Set(ctx, map[string]any{
			"info": channelInfo,
		}, firestore.MergeAll)
		if err != nil {
			// Log error but don't fail the message storage
			// Channel info is supplementary data
			return goerr.Wrap(err, "failed to store channel info (non-critical)",
				goerr.V("channel_id", messageLog.ChannelID))
		}
	}

	return nil
}

// GetSlackMessageLogs retrieves message logs with filtering from Firestore using hierarchical structure
func (c *Client) GetSlackMessageLogs(ctx context.Context, filter *interfaces.SlackMessageLogFilter) ([]*slack.SlackMessageLog, error) {
	if filter == nil {
		return nil, goerr.New("filter cannot be nil")
	}

	var results []*slack.SlackMessageLog

	// If ChannelID is specified, query that specific channel's messages subcollection
	if filter.ChannelID != "" {
		channelResults, err := c.getMessagesFromChannel(ctx, filter)
		if err != nil {
			return nil, err
		}
		results = append(results, channelResults...)
	} else {
		// If no ChannelID specified, we need to use collection group query to search across all channels
		query := c.client.CollectionGroup(subCollectionMessages).Query

		// Apply non-channel filters
		if filter.UserID != "" {
			query = query.Where("user_id", "==", filter.UserID)
		}
		if filter.ChannelType != "" {
			query = query.Where("channel_type", "==", filter.ChannelType)
		}
		if filter.MessageType != "" {
			query = query.Where("message_type", "==", filter.MessageType)
		}
		if filter.FromTime != nil {
			query = query.Where("created_at", ">=", *filter.FromTime)
		}
		if filter.ToTime != nil {
			query = query.Where("created_at", "<=", *filter.ToTime)
		}

		// Order by created_at descending (newest first)
		query = query.OrderBy("created_at", firestore.Desc)

		// Apply limit
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}

		iter := query.Documents(ctx)
		defer iter.Stop()

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, goerr.Wrap(err, "failed to iterate slack message logs")
			}

			var messageLog slack.SlackMessageLog
			if err := doc.DataTo(&messageLog); err != nil {
				return nil, goerr.Wrap(err, "failed to parse slack message log",
					goerr.V("doc_id", doc.Ref.ID))
			}

			// Ensure ID is set from document ID
			if messageLog.ID == "" {
				messageLog.ID = types.MessageID(doc.Ref.ID)
			}

			results = append(results, &messageLog)
		}
	}

	return results, nil
}

// getMessagesFromChannel retrieves messages from a specific channel's subcollection
func (c *Client) getMessagesFromChannel(ctx context.Context, filter *interfaces.SlackMessageLogFilter) ([]*slack.SlackMessageLog, error) {
	query := c.client.Collection(collectionSlackChannels).
		Doc(filter.ChannelID).
		Collection(subCollectionMessages).
		Query

	// Apply filters
	if filter.UserID != "" {
		query = query.Where("user_id", "==", filter.UserID)
	}
	if filter.ChannelType != "" {
		query = query.Where("channel_type", "==", filter.ChannelType)
	}
	if filter.MessageType != "" {
		query = query.Where("message_type", "==", filter.MessageType)
	}
	if filter.FromTime != nil {
		query = query.Where("created_at", ">=", *filter.FromTime)
	}
	if filter.ToTime != nil {
		query = query.Where("created_at", "<=", *filter.ToTime)
	}

	// Order by created_at descending (newest first)
	query = query.OrderBy("created_at", firestore.Desc)

	// Apply limit
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var results []*slack.SlackMessageLog
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, goerr.Wrap(err, "failed to iterate channel messages")
		}

		var messageLog slack.SlackMessageLog
		if err := doc.DataTo(&messageLog); err != nil {
			return nil, goerr.Wrap(err, "failed to parse slack message log",
				goerr.V("doc_id", doc.Ref.ID))
		}

		// Ensure ID is set from document ID
		if messageLog.ID == "" {
			messageLog.ID = types.MessageID(doc.Ref.ID)
		}

		results = append(results, &messageLog)
	}

	return results, nil
}

// GetSlackMessageLogsByChannel retrieves message logs for a specific channel from Firestore
func (c *Client) GetSlackMessageLogsByChannel(ctx context.Context, channelID string, limit int) ([]*slack.SlackMessageLog, error) {
	if channelID == "" {
		return nil, goerr.New("channelID cannot be empty")
	}

	filter := &interfaces.SlackMessageLogFilter{
		ChannelID: channelID,
		Limit:     limit,
	}

	return c.GetSlackMessageLogs(ctx, filter)
}

// GetSlackMessageLogsByUser retrieves message logs for a specific user from Firestore
func (c *Client) GetSlackMessageLogsByUser(ctx context.Context, userID string, limit int) ([]*slack.SlackMessageLog, error) {
	if userID == "" {
		return nil, goerr.New("userID cannot be empty")
	}

	filter := &interfaces.SlackMessageLogFilter{
		UserID: userID,
		Limit:  limit,
	}

	return c.GetSlackMessageLogs(ctx, filter)
}
