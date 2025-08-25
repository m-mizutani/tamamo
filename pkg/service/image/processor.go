package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	imageModel "github.com/m-mizutani/tamamo/pkg/domain/model/image"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"golang.org/x/image/draw"
)

// ThumbnailSize represents thumbnail dimensions
type ThumbnailSize struct {
	Name   string
	Width  int
	Height int
}

// ProcessorConfig holds configuration for image processing
type ProcessorConfig struct {
	ThumbnailSizes []ThumbnailSize
}

// DefaultProcessorConfig returns default processor configuration
func DefaultProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		ThumbnailSizes: []ThumbnailSize{
			{Name: "small", Width: 64, Height: 64},
			{Name: "medium", Width: 256, Height: 256},
		},
	}
}

// Processor handles image processing operations
type Processor struct {
	validator       *Validator
	storage         interfaces.StorageAdapter
	repository      interfaces.AgentImageRepository
	agentRepository interfaces.AgentRepository
	config          ProcessorConfig
}

// NewProcessor creates a new image processor
func NewProcessor(
	validator *Validator,
	storage interfaces.StorageAdapter,
	repository interfaces.AgentImageRepository,
	agentRepository interfaces.AgentRepository,
	config ProcessorConfig,
) *Processor {
	return &Processor{
		validator:       validator,
		storage:         storage,
		repository:      repository,
		agentRepository: agentRepository,
		config:          config,
	}
}

// ProcessAndStore processes an image file and stores it with thumbnails
func (p *Processor) ProcessAndStore(ctx context.Context, agentID types.UUID, file io.ReadSeeker, contentType string, fileSize int64) (*imageModel.AgentImage, error) {
	slog.Debug("Starting image processing",
		slog.String("agent_id", agentID.String()),
		slog.String("content_type", contentType),
		slog.Int64("file_size", fileSize))

	// Validate the image file
	metadata, err := p.validator.ValidateFile(file, contentType, fileSize)
	if err != nil {
		return nil, goerr.Wrap(err, "image validation failed")
	}

	slog.Debug("Image validation successful",
		slog.String("agent_id", agentID.String()),
		slog.String("format", metadata.Format),
		slog.Int("width", metadata.Width),
		slog.Int("height", metadata.Height))

	// Reset file position for storage
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, goerr.Wrap(err, "failed to reset file position")
	}

	// Generate unique storage key
	imageID := types.NewUUID(ctx)
	storageKey := p.generateStorageKey(imageID, metadata.Format)

	slog.Debug("Generated storage key",
		slog.String("agent_id", agentID.String()),
		slog.String("image_id", imageID.String()),
		slog.String("storage_key", storageKey))

	// Store original image
	imageData, err := io.ReadAll(file)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read image data")
	}

	slog.Debug("Storing original image to storage",
		slog.String("agent_id", agentID.String()),
		slog.String("storage_key", storageKey),
		slog.Int("data_size", len(imageData)))

	if err := p.storage.Put(ctx, storageKey, imageData); err != nil {
		return nil, goerr.Wrap(err, "failed to store original image", goerr.V("key", storageKey))
	}

	slog.Debug("Original image stored successfully",
		slog.String("agent_id", agentID.String()),
		slog.String("storage_key", storageKey))

	// Generate and store thumbnails
	thumbnailKeys := make(map[string]string)
	if len(p.config.ThumbnailSizes) > 0 {
		thumbnails, err := p.generateThumbnails(imageData, metadata.Format)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to generate thumbnails")
		}

		for size, thumbnailData := range thumbnails {
			thumbnailKey := p.generateThumbnailKey(imageID, size, metadata.Format)
			if err := p.storage.Put(ctx, thumbnailKey, thumbnailData); err != nil {
				return nil, goerr.Wrap(err, "failed to store thumbnail", goerr.V("key", thumbnailKey), goerr.V("size", size))
			}
			thumbnailKeys[size] = thumbnailKey
		}
	}

	// Create AgentImage entity
	now := time.Now()
	agentImage := &imageModel.AgentImage{
		ID:            imageID,
		AgentID:       agentID,
		StorageKey:    storageKey,
		ContentType:   metadata.ContentType,
		FileSize:      metadata.FileSize,
		Width:         metadata.Width,
		Height:        metadata.Height,
		ThumbnailKeys: thumbnailKeys,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Store in repository
	if err := p.repository.Create(ctx, agentImage); err != nil {
		return nil, goerr.Wrap(err, "failed to store image metadata")
	}

	// Update agent with new image ID (if agent repository is available)
	if p.agentRepository != nil {
		slog.Debug("Updating agent with image ID",
			slog.String("agent_id", agentID.String()),
			slog.String("image_id", imageID.String()))

		agent, err := p.agentRepository.GetAgent(ctx, agentID)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to get agent for image ID update")
		}

		slog.Debug("Retrieved agent for update",
			slog.String("agent_id", agentID.String()),
			slog.String("agent_name", agent.Name),
			slog.Any("current_image_id", agent.ImageID))

		agent.ImageID = &imageID
		if err := p.agentRepository.UpdateAgent(ctx, agent); err != nil {
			return nil, goerr.Wrap(err, "failed to update agent with image ID")
		}

		slog.Debug("Agent updated successfully with image ID",
			slog.String("agent_id", agentID.String()),
			slog.String("image_id", imageID.String()))
	} else {
		slog.Debug("Agent repository not available, skipping agent update",
			slog.String("agent_id", agentID.String()),
			slog.String("image_id", imageID.String()))
	}

	return agentImage, nil
}

// GetImageData retrieves image data from storage
func (p *Processor) GetImageData(ctx context.Context, agentImage *imageModel.AgentImage) ([]byte, error) {
	data, err := p.storage.Get(ctx, agentImage.StorageKey)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to retrieve image data", goerr.V("key", agentImage.StorageKey))
	}
	return data, nil
}

// GetThumbnailData retrieves thumbnail data from storage
func (p *Processor) GetThumbnailData(ctx context.Context, agentImage *imageModel.AgentImage, size string) ([]byte, error) {
	thumbnailKey, exists := agentImage.ThumbnailKeys[size]
	if !exists {
		return nil, goerr.New("thumbnail size not available", goerr.V("size", size), goerr.V("available", agentImage.ThumbnailKeys))
	}

	data, err := p.storage.Get(ctx, thumbnailKey)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to retrieve thumbnail data", goerr.V("key", thumbnailKey), goerr.V("size", size))
	}
	return data, nil
}

// generateStorageKey creates a storage key for the original image
func (p *Processor) generateStorageKey(imageID types.UUID, format string) string {
	extension := p.getFileExtension(format)
	return fmt.Sprintf("images/%s%s", imageID.String(), extension)
}

// generateThumbnailKey creates a storage key for a thumbnail
func (p *Processor) generateThumbnailKey(imageID types.UUID, size, format string) string {
	extension := p.getFileExtension(format)
	return fmt.Sprintf("thumbnails/%s/%s%s", size, imageID.String(), extension)
}

// getFileExtension returns the file extension for a given format
func (p *Processor) getFileExtension(format string) string {
	switch format {
	case "jpeg":
		return ".jpg"
	case "png":
		return ".png"
	default:
		return ".bin"
	}
}

// generateThumbnails creates thumbnails for all configured sizes
func (p *Processor) generateThumbnails(imageData []byte, format string) (map[string][]byte, error) {
	// Decode original image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to decode image for thumbnail generation")
	}

	thumbnails := make(map[string][]byte)

	for _, size := range p.config.ThumbnailSizes {
		thumbnail, err := p.createThumbnail(img, size.Width, size.Height, format)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create thumbnail", goerr.V("size", size.Name))
		}
		thumbnails[size.Name] = thumbnail
	}

	return thumbnails, nil
}

// createThumbnail creates a single thumbnail with specified dimensions
func (p *Processor) createThumbnail(src image.Image, width, height int, format string) ([]byte, error) {
	// Calculate aspect-preserving dimensions
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	// Calculate scale to fit within target dimensions while preserving aspect ratio
	scaleX := float64(width) / float64(srcWidth)
	scaleY := float64(height) / float64(srcHeight)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	// Calculate actual thumbnail dimensions
	thumbWidth := int(float64(srcWidth) * scale)
	thumbHeight := int(float64(srcHeight) * scale)

	// Create thumbnail image
	thumb := image.NewRGBA(image.Rect(0, 0, thumbWidth, thumbHeight))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), src, srcBounds, draw.Over, nil)

	// Encode thumbnail
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 85}); err != nil {
			return nil, goerr.Wrap(err, "failed to encode JPEG thumbnail")
		}
	case "png":
		if err := png.Encode(&buf, thumb); err != nil {
			return nil, goerr.Wrap(err, "failed to encode PNG thumbnail")
		}
	default:
		return nil, goerr.New("unsupported format for thumbnail generation", goerr.V("format", format))
	}

	return buf.Bytes(), nil
}
