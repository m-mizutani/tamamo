package firestore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/image"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const agentImagesCollection = "agent_images"

// AgentImage Firestore document structure
type agentImageDoc struct {
	ID            string            `firestore:"id"`
	AgentID       string            `firestore:"agent_id"`
	StorageKey    string            `firestore:"storage_key"`
	ContentType   string            `firestore:"content_type"`
	FileSize      int64             `firestore:"file_size"`
	Width         int               `firestore:"width"`
	Height        int               `firestore:"height"`
	ThumbnailKeys map[string]string `firestore:"thumbnail_keys"`
	CreatedAt     time.Time         `firestore:"created_at"`
	UpdatedAt     time.Time         `firestore:"updated_at"`
}

// AgentImageRepository implements interfaces.AgentImageRepository for Firestore
type AgentImageRepository struct {
	client     *firestore.Client
	databaseID string
}

// NewAgentImageRepository creates a new Firestore agent image repository
func (c *Client) NewAgentImageRepository() interfaces.AgentImageRepository {
	return &AgentImageRepository{
		client:     c.client,
		databaseID: c.databaseID,
	}
}

// Create creates a new agent image
func (r *AgentImageRepository) Create(ctx context.Context, agentImage *image.AgentImage) error {
	doc := r.client.Collection(agentImagesCollection).Doc(agentImage.ID.String())

	// Convert to Firestore document
	firestoreDoc := agentImageToDoc(agentImage)

	_, err := doc.Set(ctx, firestoreDoc)
	if err != nil {
		return fmt.Errorf("failed to create agent image: %w", err)
	}

	return nil
}

// GetByID retrieves an agent image by its ID
func (r *AgentImageRepository) GetByID(ctx context.Context, id types.UUID) (*image.AgentImage, error) {
	doc := r.client.Collection(agentImagesCollection).Doc(id.String())

	snapshot, err := doc.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, image.ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to get agent image: %w", err)
	}

	var firestoreDoc agentImageDoc
	if err := snapshot.DataTo(&firestoreDoc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent image: %w", err)
	}

	// Convert to domain model
	return docToAgentImage(&firestoreDoc)
}

// Update updates an existing agent image
func (r *AgentImageRepository) Update(ctx context.Context, agentImage *image.AgentImage) error {
	doc := r.client.Collection(agentImagesCollection).Doc(agentImage.ID.String())

	// Convert to Firestore document
	firestoreDoc := agentImageToDoc(agentImage)

	_, err := doc.Set(ctx, firestoreDoc)
	if err != nil {
		return fmt.Errorf("failed to update agent image: %w", err)
	}

	return nil
}
