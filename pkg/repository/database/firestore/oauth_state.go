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

const oauthStateCollection = "oauth_states"

type oauthStateRepository struct {
	client *firestore.Client
}

// NewOAuthStateRepository creates a new Firestore OAuth state repository
func NewOAuthStateRepository(client *firestore.Client) interfaces.OAuthStateRepository {
	return &oauthStateRepository{
		client: client,
	}
}

func (r *oauthStateRepository) SaveState(ctx context.Context, state *auth.OAuthState) error {
	if state == nil {
		return goerr.New("state is nil")
	}
	if state.State == "" {
		return goerr.New("state token is empty")
	}

	docRef := r.client.Collection(oauthStateCollection).Doc(state.State)

	// Check if state already exists
	_, err := docRef.Get(ctx)
	if err == nil {
		return goerr.New("state already exists", goerr.V("state", state.State))
	}
	if status.Code(err) != codes.NotFound {
		return goerr.Wrap(err, "failed to check state existence")
	}

	// Create the state
	_, err = docRef.Create(ctx, state)
	if err != nil {
		return goerr.Wrap(err, "failed to create state")
	}

	// Clean up expired states in background
	go func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = r.cleanupExpiredStates(cleanupCtx)
	}()

	return nil
}

func (r *oauthStateRepository) GetState(ctx context.Context, stateToken string) (*auth.OAuthState, error) {
	docRef := r.client.Collection(oauthStateCollection).Doc(stateToken)
	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, auth.ErrStateNotFound
		}
		return nil, goerr.Wrap(err, "failed to get state")
	}

	var state auth.OAuthState
	if err := doc.DataTo(&state); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal state")
	}

	// Check if state is expired
	if state.IsExpired() {
		return nil, auth.ErrStateExpired
	}

	return &state, nil
}

func (r *oauthStateRepository) ValidateAndDeleteState(ctx context.Context, stateToken string) error {
	docRef := r.client.Collection(oauthStateCollection).Doc(stateToken)

	// Get the state first
	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return auth.ErrStateNotFound
		}
		return goerr.Wrap(err, "failed to get state")
	}

	var state auth.OAuthState
	if err := doc.DataTo(&state); err != nil {
		return goerr.Wrap(err, "failed to unmarshal state")
	}

	// Check if state is expired
	if state.IsExpired() {
		// Delete expired state
		_, _ = docRef.Delete(ctx)
		return auth.ErrStateExpired
	}

	// Delete the state after successful validation
	_, err = docRef.Delete(ctx)
	if err != nil {
		return goerr.Wrap(err, "failed to delete state")
	}

	return nil
}

// cleanupExpiredStates removes expired OAuth states
func (r *oauthStateRepository) cleanupExpiredStates(ctx context.Context) error {
	now := time.Now()

	// Query for expired states
	query := r.client.Collection(oauthStateCollection).Where("expires_at", "<", now)
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
			return goerr.Wrap(err, "failed to iterate expired states")
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
