package firestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"google.golang.org/api/iterator"
)

const (
	collectionSlackChannels = "log_slack_channels"
	subCollectionMessages   = "messages"
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

// GetSlackMessageLogs retrieves message logs from Firestore using hierarchical structure
func (c *Client) GetSlackMessageLogs(ctx context.Context, channel string, from *time.Time, to *time.Time, limit int, offset int) ([]*slack.SlackMessageLog, error) {
	// Apply default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	var results []*slack.SlackMessageLog

	// If channel is specified, query that specific channel's messages subcollection
	if channel != "" {
		query := c.client.Collection(collectionSlackChannels).
			Doc(channel).
			Collection(subCollectionMessages).
			Query

		// Apply time range filters
		if from != nil {
			query = query.Where("created_at", ">=", *from)
		}
		if to != nil {
			query = query.Where("created_at", "<=", *to)
		}

		// Order by created_at descending (newest first) only if time filters are used
		if from != nil || to != nil {
			query = query.OrderBy("created_at", firestore.Desc)
		}

		// Apply offset and limit
		if offset > 0 {
			query = query.Offset(offset)
		}
		query = query.Limit(limit)

		iter := query.Documents(ctx)
		defer iter.Stop()

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
	} else {
		// If no channel specified, use collection group query to search across all channels
		query := c.client.CollectionGroup(subCollectionMessages).Query

		// Apply time range filters
		if from != nil {
			query = query.Where("created_at", ">=", *from)
		}
		if to != nil {
			query = query.Where("created_at", "<=", *to)
		}

		// Order by created_at descending (newest first)
		query = query.OrderBy("created_at", firestore.Desc)

		// Apply offset and limit
		if offset > 0 {
			query = query.Offset(offset)
		}
		query = query.Limit(limit)

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
