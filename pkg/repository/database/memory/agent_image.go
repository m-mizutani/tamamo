package memory

import (
	"context"
	"sync"

	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/image"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// AgentImageRepository implements interfaces.AgentImageRepository for in-memory storage
type AgentImageRepository struct {
	images       map[string]*image.AgentImage
	agentToImage map[string]string // agentID -> imageID mapping
	mu           sync.RWMutex
}

// NewAgentImageRepository creates a new in-memory agent image repository
func NewAgentImageRepository() interfaces.AgentImageRepository {
	return &AgentImageRepository{
		images:       make(map[string]*image.AgentImage),
		agentToImage: make(map[string]string),
	}
}

// Create creates a new agent image
func (r *AgentImageRepository) Create(ctx context.Context, agentImage *image.AgentImage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a copy to avoid external modifications
	imageCopy := *agentImage

	r.images[agentImage.ID.String()] = &imageCopy
	r.agentToImage[agentImage.AgentID.String()] = agentImage.ID.String()

	return nil
}

// GetByID retrieves an agent image by its ID
func (r *AgentImageRepository) GetByID(ctx context.Context, id types.UUID) (*image.AgentImage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentImage, exists := r.images[id.String()]
	if !exists {
		return nil, image.ErrImageNotFound
	}

	// Return a copy to avoid external modifications
	imageCopy := *agentImage
	return &imageCopy, nil
}

// Update updates an existing agent image
func (r *AgentImageRepository) Update(ctx context.Context, agentImage *image.AgentImage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.images[agentImage.ID.String()]
	if !exists {
		return image.ErrImageNotFound
	}

	// Create a copy to avoid external modifications
	imageCopy := *agentImage

	r.images[agentImage.ID.String()] = &imageCopy
	r.agentToImage[agentImage.AgentID.String()] = agentImage.ID.String()

	return nil
}
