package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// Collection names
	collectionThreads   = "threads"
	collectionMessages  = "messages"
	collectionHistories = "histories"
)

// Client is a Firestore implementation of ThreadRepository and HistoryRepository
type Client struct {
	client     *firestore.Client
	projectID  string
	databaseID string
}

// New creates a new Firestore client using Application Default Credentials
func New(ctx context.Context, projectID, databaseID string) (*Client, error) {
	if projectID == "" {
		return nil, goerr.New("project ID is required")
	}
	if databaseID == "" {
		databaseID = "(default)"
	}

	// Create Firestore client with ADC
	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create firestore client",
			goerr.V("project_id", projectID),
			goerr.V("database_id", databaseID))
	}

	return &Client{
		client:     client,
		projectID:  projectID,
		databaseID: databaseID,
	}, nil
}

// Close closes the Firestore client
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// GetClient returns the underlying Firestore client
func (c *Client) GetClient() *firestore.Client {
	return c.client
}

// GetOrPutThread gets an existing thread or creates a new one atomically using Firestore transaction
func (c *Client) GetOrPutThread(ctx context.Context, teamID, channelID, threadTS string) (*slack.Thread, error) {
	return c.GetOrPutThreadWithAgent(ctx, teamID, channelID, threadTS, nil, "")
}

// GetOrPutThreadWithAgent gets an existing thread or creates a new one with agent information atomically using Firestore transaction
func (c *Client) GetOrPutThreadWithAgent(ctx context.Context, teamID, channelID, threadTS string, agentUUID *types.UUID, agentVersion string) (*slack.Thread, error) {
	// Pre-generate thread outside transaction to avoid time-dependent operations inside transaction
	var newThread *slack.Thread
	if agentUUID != nil || agentVersion != "" {
		newThread = slack.NewThreadWithAgent(ctx, teamID, channelID, threadTS, agentUUID, agentVersion)
	} else {
		newThread = slack.NewThread(ctx, teamID, channelID, threadTS)
	}
	if err := newThread.Validate(); err != nil {
		return nil, goerr.Wrap(err, "invalid thread", goerr.V("thread_id", newThread.ID))
	}

	var result *slack.Thread
	err := c.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Query for existing thread by channel and timestamp
		query := c.client.Collection(collectionThreads).
			Where("ChannelID", "==", channelID).
			Where("ThreadTS", "==", threadTS).
			Limit(1)
		docs, err := tx.Documents(query).GetAll()
		if err != nil {
			return goerr.Wrap(err, "failed to query threads")
		}
		if len(docs) > 0 {
			// Thread exists, decode and return it
			t := &slack.Thread{}
			if err := docs[0].DataTo(t); err != nil {
				return goerr.Wrap(err, "failed to decode thread", goerr.V("doc_id", docs[0].Ref.ID))
			}
			result = t
			return nil
		}
		// Thread doesn't exist, use pre-generated thread
		if err := tx.Set(c.client.Collection(collectionThreads).Doc(newThread.ID.String()), newThread); err != nil {
			return goerr.Wrap(err, "failed to create thread", goerr.V("thread_id", newThread.ID))
		}
		result = newThread
		return nil
	})
	if err != nil {
		return nil, goerr.Wrap(err, "transaction failed",
			goerr.V("team_id", teamID),
			goerr.V("channel_id", channelID),
			goerr.V("thread_ts", threadTS))
	}
	return result, nil
}

// GetThread retrieves a thread from Firestore
func (c *Client) GetThread(ctx context.Context, id types.ThreadID) (*slack.Thread, error) {
	doc, err := c.client.Collection(collectionThreads).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, goerr.Wrap(slack.ErrThreadNotFound, "thread not found",
				goerr.V("thread_id", id),
				goerr.V("repository", "firestore"))
		}
		return nil, goerr.Wrap(err, "failed to get thread",
			goerr.V("thread_id", id),
			goerr.V("repository", "firestore"))
	}

	var t slack.Thread
	if err := doc.DataTo(&t); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal thread",
			goerr.V("thread_id", id),
			goerr.V("repository", "firestore"))
	}

	return &t, nil
}

// GetThreadByTS retrieves a thread by channel ID and thread timestamp from Firestore
func (c *Client) GetThreadByTS(ctx context.Context, channelID, threadTS string) (*slack.Thread, error) {
	// Query for thread by ChannelID and ThreadTS
	iter := c.client.Collection(collectionThreads).
		Where("ChannelID", "==", channelID).
		Where("ThreadTS", "==", threadTS).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, goerr.Wrap(slack.ErrThreadNotFound, "thread not found",
			goerr.V("channel_id", channelID),
			goerr.V("thread_ts", threadTS),
			goerr.V("repository", "firestore"))
	}
	if err != nil {
		return nil, goerr.Wrap(err, "failed to query thread",
			goerr.V("channel_id", channelID),
			goerr.V("thread_ts", threadTS),
			goerr.V("repository", "firestore"))
	}

	var t slack.Thread
	if err := doc.DataTo(&t); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal thread",
			goerr.V("channel_id", channelID),
			goerr.V("thread_ts", threadTS),
			goerr.V("repository", "firestore"))
	}

	return &t, nil
}

// ListThreads retrieves a paginated list of threads sorted by creation time (newest first)
func (c *Client) ListThreads(ctx context.Context, offset, limit int) ([]*slack.Thread, int, error) {
	// Validate parameters
	if offset < 0 {
		return nil, 0, goerr.New("offset must be non-negative", goerr.V("offset", offset))
	}
	if limit < 0 {
		return nil, 0, goerr.New("limit must be non-negative", goerr.V("limit", limit))
	}

	// First, get total count
	countQuery := c.client.Collection(collectionThreads)
	totalDocs, err := countQuery.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, goerr.Wrap(err, "failed to get total thread count",
			goerr.V("repository", "firestore"))
	}
	totalCount := len(totalDocs)

	// If offset is beyond total count, return empty result
	if offset >= totalCount {
		return []*slack.Thread{}, totalCount, nil
	}

	// Build query with pagination
	query := c.client.Collection(collectionThreads).
		OrderBy("CreatedAt", firestore.Desc). // Newest first
		Offset(offset)

	if limit > 0 {
		query = query.Limit(limit)
	}

	// Execute query
	iter := query.Documents(ctx)
	defer iter.Stop()

	var threads []*slack.Thread
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, goerr.Wrap(err, "failed to iterate threads",
				goerr.V("repository", "firestore"))
		}

		var thread slack.Thread
		if err := doc.DataTo(&thread); err != nil {
			return nil, 0, goerr.Wrap(err, "failed to unmarshal thread",
				goerr.V("thread_id", doc.Ref.ID),
				goerr.V("repository", "firestore"))
		}
		threads = append(threads, &thread)
	}

	return threads, totalCount, nil
}

// PutThreadMessage stores a message in a thread's subcollection
func (c *Client) PutThreadMessage(ctx context.Context, threadID types.ThreadID, msg *slack.Message) error {
	if err := msg.Validate(); err != nil {
		return goerr.Wrap(err, "invalid message",
			goerr.V("thread_id", threadID),
			goerr.V("message_id", msg.ID))
	}

	// Check if thread exists
	_, err := c.GetThread(ctx, threadID)
	if err != nil {
		return err
	}

	// Store message in subcollection
	_, err = c.client.Collection(collectionThreads).Doc(threadID.String()).
		Collection(collectionMessages).Doc(msg.ID.String()).Set(ctx, msg)
	if err != nil {
		return goerr.Wrap(err, "failed to put message",
			goerr.V("thread_id", threadID),
			goerr.V("message_id", msg.ID),
			goerr.V("repository", "firestore"))
	}

	return nil
}

// GetThreadMessages retrieves all messages in a thread
func (c *Client) GetThreadMessages(ctx context.Context, threadID types.ThreadID) ([]*slack.Message, error) {
	// Check if thread exists
	_, err := c.GetThread(ctx, threadID)
	if err != nil {
		return nil, err
	}

	// Get messages from subcollection, ordered by CreatedAt
	iter := c.client.Collection(collectionThreads).Doc(threadID.String()).
		Collection(collectionMessages).OrderBy("CreatedAt", firestore.Asc).Documents(ctx)
	defer iter.Stop()

	var messages []*slack.Message
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, goerr.Wrap(err, "failed to iterate messages",
				goerr.V("thread_id", threadID),
				goerr.V("repository", "firestore"))
		}

		var msg slack.Message
		if err := doc.DataTo(&msg); err != nil {
			return nil, goerr.Wrap(err, "failed to unmarshal message",
				goerr.V("thread_id", threadID),
				goerr.V("message_id", doc.Ref.ID),
				goerr.V("repository", "firestore"))
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// PutHistory stores a history record in Firestore
func (c *Client) PutHistory(ctx context.Context, history *slack.History) error {
	if err := history.Validate(); err != nil {
		return goerr.Wrap(err, "invalid history", goerr.V("history_id", history.ID))
	}

	// Check if thread exists
	_, err := c.GetThread(ctx, history.ThreadID)
	if err != nil {
		return err
	}

	// Store history record as subcollection of thread
	_, err = c.client.Collection(collectionThreads).Doc(history.ThreadID.String()).
		Collection(collectionHistories).Doc(history.ID.String()).Set(ctx, history)
	if err != nil {
		return goerr.Wrap(err, "failed to put history",
			goerr.V("history_id", history.ID),
			goerr.V("thread_id", history.ThreadID),
			goerr.V("repository", "firestore"))
	}

	return nil
}

// GetLatestHistory retrieves the most recent history for a thread
func (c *Client) GetLatestHistory(ctx context.Context, threadID types.ThreadID) (*slack.History, error) {
	// Check if thread exists
	_, err := c.GetThread(ctx, threadID)
	if err != nil {
		return nil, err
	}

	// Query for the latest history record in thread's subcollection
	iter := c.client.Collection(collectionThreads).Doc(threadID.String()).
		Collection(collectionHistories).
		OrderBy("CreatedAt", firestore.Desc).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		// No history found - this is expected for new threads
		return nil, goerr.Wrap(slack.ErrHistoryNotFound, "no history found for thread",
			goerr.V("thread_id", threadID),
			goerr.V("repository", "firestore"))
	}
	if err != nil {
		return nil, goerr.Wrap(err, "failed to query latest history",
			goerr.V("thread_id", threadID),
			goerr.V("repository", "firestore"))
	}

	var h slack.History
	if err := doc.DataTo(&h); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal history",
			goerr.V("thread_id", threadID),
			goerr.V("history_id", doc.Ref.ID),
			goerr.V("repository", "firestore"))
	}

	return &h, nil
}

// GetHistoryByID retrieves a specific history record by ID
// Uses collection group query to search across all histories subcollections under threads
func (c *Client) GetHistoryByID(ctx context.Context, id types.HistoryID) (*slack.History, error) {
	// Collection group query searches across all "histories" subcollections
	// that exist under any "threads" document
	iter := c.client.CollectionGroup(collectionHistories).
		Where("ID", "==", id.String()).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, goerr.Wrap(slack.ErrHistoryNotFound, "history not found",
			goerr.V("history_id", id),
			goerr.V("repository", "firestore"))
	}
	if err != nil {
		return nil, goerr.Wrap(err, "failed to query history by id",
			goerr.V("history_id", id),
			goerr.V("repository", "firestore"))
	}

	var h slack.History
	if err := doc.DataTo(&h); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal history",
			goerr.V("history_id", id),
			goerr.V("repository", "firestore"))
	}

	return &h, nil
}

// GetHistoryByIDWithThread is a more efficient version when thread ID is known
func (c *Client) GetHistoryByIDWithThread(ctx context.Context, threadID types.ThreadID, id types.HistoryID) (*slack.History, error) {
	doc, err := c.client.Collection(collectionThreads).Doc(threadID.String()).
		Collection(collectionHistories).Doc(id.String()).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, goerr.Wrap(slack.ErrHistoryNotFound, "history not found",
			goerr.V("history_id", id),
			goerr.V("repository", "firestore"))
	}
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get history",
			goerr.V("history_id", id),
			goerr.V("repository", "firestore"))
	}

	var h slack.History
	if err := doc.DataTo(&h); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal history",
			goerr.V("history_id", id),
			goerr.V("repository", "firestore"))
	}

	return &h, nil
}
