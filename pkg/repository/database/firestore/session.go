package firestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const sessionCollection = "sessions"

type sessionRepository struct {
	client *firestore.Client
}

// NewSessionRepository creates a new Firestore session repository
func NewSessionRepository(client *firestore.Client) interfaces.SessionRepository {
	return &sessionRepository{
		client: client,
	}
}

func (r *sessionRepository) CreateSession(ctx context.Context, session *auth.Session) error {
	if session == nil {
		return goerr.New("session is nil")
	}
	if !session.ID.IsValid() {
		return goerr.New("invalid session ID")
	}

	sessionID := session.ID.String()
	docRef := r.client.Collection(sessionCollection).Doc(sessionID)

	// Check if session already exists
	_, err := docRef.Get(ctx)
	if err == nil {
		return goerr.New("session already exists", goerr.V("sessionID", sessionID))
	}
	if status.Code(err) != codes.NotFound {
		return goerr.Wrap(err, "failed to check session existence")
	}

	// Create the session
	_, err = docRef.Create(ctx, session)
	if err != nil {
		return goerr.Wrap(err, "failed to create session")
	}

	return nil
}

func (r *sessionRepository) GetSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	docRef := r.client.Collection(sessionCollection).Doc(sessionID)
	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, auth.ErrSessionNotFound
		}
		return nil, goerr.Wrap(err, "failed to get session")
	}

	var session auth.Session
	if err := doc.DataTo(&session); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal session")
	}

	// Check if session is expired
	if session.IsExpired() {
		return nil, auth.ErrSessionExpired
	}

	return &session, nil
}

func (r *sessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	docRef := r.client.Collection(sessionCollection).Doc(sessionID)

	// Check if session exists
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return auth.ErrSessionNotFound
		}
		return goerr.Wrap(err, "failed to check session existence")
	}

	// Delete the session
	_, err = docRef.Delete(ctx)
	if err != nil {
		return goerr.Wrap(err, "failed to delete session")
	}

	return nil
}

func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	now := time.Now()

	// Query for expired sessions
	query := r.client.Collection(sessionCollection).Where("expires_at", "<", now)
	iter := query.Documents(ctx)
	defer iter.Stop()

	// Use BulkWriter for batch operations
	bulkWriter := r.client.BulkWriter(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return goerr.Wrap(err, "failed to iterate expired sessions")
		}

		_, err = bulkWriter.Delete(doc.Ref)
		if err != nil {
			return goerr.Wrap(err, "failed to add delete to bulk writer")
		}
	}

	// Flush all pending writes
	bulkWriter.End()

	return nil
}
