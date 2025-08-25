package image

import (
	"context"
	"time"

	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

type AgentImage struct {
	ID            types.UUID        `json:"id"`
	AgentID       types.UUID        `json:"agent_id"`
	StorageKey    string            `json:"storage_key"`
	ContentType   string            `json:"content_type"`
	FileSize      int64             `json:"file_size"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	ThumbnailKeys map[string]string `json:"thumbnail_keys,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// NewAgentImage creates a new AgentImage instance
func NewAgentImage(ctx context.Context, agentID types.UUID, storageKey string, contentType string, fileSize int64, width, height int) *AgentImage {
	now := time.Now()
	return &AgentImage{
		ID:            types.NewUUID(ctx),
		AgentID:       agentID,
		StorageKey:    storageKey,
		ContentType:   contentType,
		FileSize:      fileSize,
		Width:         width,
		Height:        height,
		ThumbnailKeys: make(map[string]string),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// AddThumbnail adds a thumbnail storage key for the specified size
func (ai *AgentImage) AddThumbnail(size string, storageKey string) {
	if ai.ThumbnailKeys == nil {
		ai.ThumbnailKeys = make(map[string]string)
	}
	ai.ThumbnailKeys[size] = storageKey
	ai.UpdatedAt = time.Now()
}

// GetThumbnail returns the storage key for the specified thumbnail size
func (ai *AgentImage) GetThumbnail(size string) (string, bool) {
	if ai.ThumbnailKeys == nil {
		return "", false
	}
	storageKey, exists := ai.ThumbnailKeys[size]
	return storageKey, exists
}

// Update updates the updatedAt timestamp
func (ai *AgentImage) Update() {
	ai.UpdatedAt = time.Now()
}
