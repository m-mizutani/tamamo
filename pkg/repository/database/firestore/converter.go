package firestore

import (
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/image"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// agentImageToDoc converts domain AgentImage to Firestore document
func agentImageToDoc(agentImage *image.AgentImage) *agentImageDoc {
	return &agentImageDoc{
		ID:            agentImage.ID.String(),
		AgentID:       agentImage.AgentID.String(),
		StorageKey:    agentImage.StorageKey,
		ContentType:   agentImage.ContentType,
		FileSize:      agentImage.FileSize,
		Width:         agentImage.Width,
		Height:        agentImage.Height,
		ThumbnailKeys: agentImage.ThumbnailKeys,
		CreatedAt:     agentImage.CreatedAt,
		UpdatedAt:     agentImage.UpdatedAt,
	}
}

// docToAgentImage converts Firestore document to domain AgentImage
func docToAgentImage(doc *agentImageDoc) (*image.AgentImage, error) {
	agentID := types.UUID(doc.AgentID)
	if !agentID.IsValid() {
		return nil, goerr.New("invalid agent UUID", goerr.V("agent_id", doc.AgentID))
	}

	return &image.AgentImage{
		ID:            types.UUID(doc.ID),
		AgentID:       agentID,
		StorageKey:    doc.StorageKey,
		ContentType:   doc.ContentType,
		FileSize:      doc.FileSize,
		Width:         doc.Width,
		Height:        doc.Height,
		ThumbnailKeys: doc.ThumbnailKeys,
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
	}, nil
}
