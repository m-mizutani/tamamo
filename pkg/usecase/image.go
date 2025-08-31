package usecase

import (
	"context"
	"strings"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/image"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/domain/types/apperr"
	imageService "github.com/m-mizutani/tamamo/pkg/service/image"
)

// ImageUseCaseImpl implements ImageUseCases interface
type ImageUseCaseImpl struct {
	imageProcessor *imageService.Processor
	agentImageRepo interfaces.AgentImageRepository
	agentUseCase   interfaces.AgentUseCases
}

// NewImageUseCases creates a new image use case
func NewImageUseCases(
	imageProcessor *imageService.Processor,
	agentImageRepo interfaces.AgentImageRepository,
	agentUseCase interfaces.AgentUseCases,
) interfaces.ImageUseCases {
	return &ImageUseCaseImpl{
		imageProcessor: imageProcessor,
		agentImageRepo: agentImageRepo,
		agentUseCase:   agentUseCase,
	}
}

// UploadAgentImage uploads and processes an agent image
func (uc *ImageUseCaseImpl) UploadAgentImage(ctx context.Context, req *interfaces.UploadImageRequest) (*image.AgentImage, error) {
	// Verify agent exists
	_, err := uc.agentUseCase.GetAgent(ctx, req.AgentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to verify agent",
			goerr.V("agent_id", req.AgentID), goerr.Tag(apperr.ErrTagAgentNotFound))
	}

	// Process and store the image
	agentImage, err := uc.imageProcessor.ProcessAndStore(ctx, req.AgentID, req.FileReader, req.ContentType, req.FileSize)
	if err != nil {
		return nil, uc.wrapImageError(err)
	}

	return agentImage, nil
}

// GetAgentImageData retrieves agent image data (original or thumbnail)
func (uc *ImageUseCaseImpl) GetAgentImageData(ctx context.Context, agentID types.UUID, thumbnailSize string) (*interfaces.ImageData, error) {
	// Get agent to retrieve the current image ID
	agentWithVersion, err := uc.agentUseCase.GetAgent(ctx, agentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent",
			goerr.V("agent_id", agentID), goerr.Tag(apperr.ErrTagAgentNotFound))
	}

	agent := agentWithVersion.Agent

	// Check if agent has an image
	if agent.ImageID == nil {
		return nil, goerr.New("agent has no image",
			goerr.V("agent_id", agentID), goerr.Tag(apperr.ErrTagAgentNoImage))
	}

	// Get agent image info using the image ID from agent
	agentImage, err := uc.agentImageRepo.GetByID(ctx, *agent.ImageID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get image",
			goerr.V("agent_id", agentID),
			goerr.V("image_id", *agent.ImageID), goerr.Tag(apperr.ErrTagImageNotFound))
	}

	var imageData []byte
	var contentType string = agentImage.ContentType

	if thumbnailSize != "" {
		// Serve thumbnail
		imageData, err = uc.imageProcessor.GetThumbnailData(ctx, agentImage, thumbnailSize)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to get thumbnail",
				goerr.V("agent_id", agentID),
				goerr.V("thumbnail_size", thumbnailSize), goerr.Tag(apperr.ErrTagThumbnailNotFound))
		}
	} else {
		// Serve original image
		imageData, err = uc.imageProcessor.GetImageData(ctx, agentImage)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to retrieve image data",
				goerr.V("agent_id", agentID),
				goerr.V("image_id", agentImage.ID),
				goerr.V("storage_key", agentImage.StorageKey), goerr.Tag(apperr.ErrTagImageRetrievalFailed))
		}
	}

	return &interfaces.ImageData{
		Data:        imageData,
		ContentType: contentType,
	}, nil
}

// GetAgentImageInfo retrieves agent image metadata
func (uc *ImageUseCaseImpl) GetAgentImageInfo(ctx context.Context, agentID types.UUID) (*image.AgentImage, error) {
	// Get agent to retrieve the current image ID
	agentWithVersion, err := uc.agentUseCase.GetAgent(ctx, agentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent",
			goerr.V("agent_id", agentID), goerr.Tag(apperr.ErrTagAgentNotFound))
	}

	agent := agentWithVersion.Agent

	// Check if agent has an image
	if agent.ImageID == nil {
		return nil, goerr.New("agent has no image",
			goerr.V("agent_id", agentID), goerr.Tag(apperr.ErrTagAgentNoImage))
	}

	// Get agent image info using the image ID from agent
	agentImage, err := uc.agentImageRepo.GetByID(ctx, *agent.ImageID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get image",
			goerr.V("agent_id", agentID),
			goerr.V("image_id", *agent.ImageID), goerr.Tag(apperr.ErrTagImageNotFound))
	}

	return agentImage, nil
}

// wrapImageError wraps image processing errors with appropriate context and error tags
func (uc *ImageUseCaseImpl) wrapImageError(err error) error {
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "invalid mime type"):
		return goerr.Wrap(err, "invalid file type", goerr.Tag(apperr.ErrTagInvalidFileType))
	case strings.Contains(errMsg, "image too large"):
		return goerr.Wrap(err, "image file too large", goerr.Tag(apperr.ErrTagImageTooLarge))
	case strings.Contains(errMsg, "image too small"):
		return goerr.Wrap(err, "image dimensions too small", goerr.Tag(apperr.ErrTagImageTooSmall))
	case strings.Contains(errMsg, "corrupted image"):
		return goerr.Wrap(err, "invalid or corrupted image file", goerr.Tag(apperr.ErrTagCorruptedImage))
	default:
		return goerr.Wrap(err, "failed to process image", goerr.Tag(apperr.ErrTagImageProcessingFailed))
	}
}
