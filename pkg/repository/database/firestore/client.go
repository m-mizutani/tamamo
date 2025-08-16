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

// Client is a Firestore implementation of ThreadRepository
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

// GetOrPutThread gets an existing thread or creates a new one atomically using Firestore transaction
func (c *Client) GetOrPutThread(ctx context.Context, teamID, channelID, threadTS string) (*slack.Thread, error) {
	var result *slack.Thread

	err := c.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Query for existing thread by channel and timestamp
		query := c.client.Collection("threads").
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

		// Thread doesn't exist, create new one
		t := slack.NewThread(ctx, teamID, channelID, threadTS)
		if err := t.Validate(); err != nil {
			return goerr.Wrap(err, "invalid thread", goerr.V("thread_id", t.ID))
		}

		// Store the new thread
		if err := tx.Set(c.client.Collection("threads").Doc(t.ID.String()), t); err != nil {
			return goerr.Wrap(err, "failed to create thread", goerr.V("thread_id", t.ID))
		}

		result = t
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
	doc, err := c.client.Collection("threads").Doc(id.String()).Get(ctx)
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
	iter := c.client.Collection("threads").
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
	_, err = c.client.Collection("threads").Doc(threadID.String()).
		Collection("messages").Doc(msg.ID.String()).Set(ctx, msg)
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
	iter := c.client.Collection("threads").Doc(threadID.String()).
		Collection("messages").OrderBy("CreatedAt", firestore.Asc).Documents(ctx)
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
