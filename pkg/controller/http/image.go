package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/errors"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/utils/safe"
)

// ImageController handles image-related HTTP requests
type ImageController struct {
	imageUseCase interfaces.ImageUseCases
}

// NewImageController creates a new image controller
func NewImageController(
	imageUseCase interfaces.ImageUseCases,
) *ImageController {
	return &ImageController{
		imageUseCase: imageUseCase,
	}
}

// UploadImageResponse represents the response for image upload
type UploadImageResponse struct {
	ID            string            `json:"id"`
	AgentID       string            `json:"agent_id"`
	StorageKey    string            `json:"storage_key"`
	ContentType   string            `json:"content_type"`
	FileSize      int64             `json:"file_size"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	ThumbnailKeys map[string]string `json:"thumbnail_keys"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

// HandleUploadAgentImage handles agent image upload via multipart form
func (c *ImageController) HandleUploadAgentImage(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		http.Error(w, "Missing agent ID", http.StatusBadRequest)
		return
	}

	ctxlog.From(r.Context()).Debug("Received image upload request",
		"agent_id", agentID,
		"method", r.Method,
		"content_type", r.Header.Get("Content-Type"))

	// Validate agent ID
	uuid := types.UUID(agentID)
	if !uuid.IsValid() {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form with 10MB max memory
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get content type and file size
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		// Fallback to header filename extension detection
		contentType = "application/octet-stream"
	}
	fileSize := fileHeader.Size

	ctxlog.From(r.Context()).Debug("Processing uploaded image",
		"agent_id", agentID,
		"filename", fileHeader.Filename,
		"content_type", contentType,
		"file_size", fileSize)

	// Upload image using use case
	req := &interfaces.UploadImageRequest{
		AgentID:     uuid,
		FileReader:  file,
		ContentType: contentType,
		FileSize:    fileSize,
	}

	agentImage, err := c.imageUseCase.UploadAgentImage(r.Context(), req)
	if err != nil {
		c.handleImageError(w, err)
		return
	}

	ctxlog.From(r.Context()).Debug("Image processing completed successfully",
		"agent_id", agentID,
		"image_id", agentImage.ID.String(),
		"storage_key", agentImage.StorageKey)

	// Convert to response format
	response := &UploadImageResponse{
		ID:            agentImage.ID.String(),
		AgentID:       agentImage.AgentID.String(),
		StorageKey:    agentImage.StorageKey,
		ContentType:   agentImage.ContentType,
		FileSize:      agentImage.FileSize,
		Width:         agentImage.Width,
		Height:        agentImage.Height,
		ThumbnailKeys: agentImage.ThumbnailKeys,
		CreatedAt:     agentImage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     agentImage.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		// At this point response headers are already sent, so we can't change status code
		// Just continue - the client will see the incomplete response
		return
	}
}

// HandleGetAgentImage handles GET requests for agent image data
func (c *ImageController) HandleGetAgentImage(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		http.Error(w, "Missing agent ID", http.StatusBadRequest)
		return
	}

	ctxlog.From(r.Context()).Debug("Received image get request",
		"agent_id", agentID,
		"query_params", r.URL.RawQuery,
		"user_agent", r.Header.Get("User-Agent"))

	// Validate agent ID
	uuid := types.UUID(agentID)
	if !uuid.IsValid() {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Get thumbnail size if requested
	thumbnailSize := r.URL.Query().Get("size")

	// Get image data using use case
	imageData, err := c.imageUseCase.GetAgentImageData(r.Context(), uuid, thumbnailSize)
	if err != nil {
		c.handleImageError(w, err)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", imageData.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(imageData.Data)))
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	ctxlog.From(r.Context()).Debug("Sending image response",
		"agent_id", agentID,
		"content_type", imageData.ContentType,
		"data_size", len(imageData.Data))

	// Write image data
	w.WriteHeader(http.StatusOK)
	safe.Write(r.Context(), w, imageData.Data)
}

// HandleGetAgentImageInfo handles GET requests for agent image metadata
func (c *ImageController) HandleGetAgentImageInfo(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		http.Error(w, "Missing agent ID", http.StatusBadRequest)
		return
	}

	// Validate agent ID
	uuid := types.UUID(agentID)
	if !uuid.IsValid() {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Get agent image info using use case
	agentImage, err := c.imageUseCase.GetAgentImageInfo(r.Context(), uuid)
	if err != nil {
		c.handleImageError(w, err)
		return
	}

	// Convert to response format
	response := &UploadImageResponse{
		ID:            agentImage.ID.String(),
		AgentID:       agentImage.AgentID.String(),
		StorageKey:    agentImage.StorageKey,
		ContentType:   agentImage.ContentType,
		FileSize:      agentImage.FileSize,
		Width:         agentImage.Width,
		Height:        agentImage.Height,
		ThumbnailKeys: agentImage.ThumbnailKeys,
		CreatedAt:     agentImage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     agentImage.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleImageError handles errors from image use cases using goerr tags for stable error classification
func (c *ImageController) handleImageError(w http.ResponseWriter, err error) {
	var httpStatus int
	var message string

	// Use goerr.Unwrap to get the goerr error and check tags
	if goErr := goerr.Unwrap(err); goErr != nil {
		switch {
		// HTTP 404 errors
		case goErr.HasTag(errors.ErrTagAgentNotFound):
			httpStatus = http.StatusNotFound
			message = "Agent not found"
		case goErr.HasTag(errors.ErrTagAgentNoImage):
			httpStatus = http.StatusNotFound
			message = "Agent has no image"
		case goErr.HasTag(errors.ErrTagImageNotFound):
			httpStatus = http.StatusNotFound
			message = "Image not found"
		case goErr.HasTag(errors.ErrTagThumbnailNotFound):
			httpStatus = http.StatusNotFound
			message = "Thumbnail not found"

		// HTTP 400 errors - validation failures
		case goErr.HasTag(errors.ErrTagInvalidFileType):
			httpStatus = http.StatusBadRequest
			message = "Invalid file type. Only JPEG and PNG are allowed."
		case goErr.HasTag(errors.ErrTagImageTooLarge):
			httpStatus = http.StatusBadRequest
			message = "Image file too large."
		case goErr.HasTag(errors.ErrTagImageTooSmall):
			httpStatus = http.StatusBadRequest
			message = "Image dimensions too small."
		case goErr.HasTag(errors.ErrTagCorruptedImage):
			httpStatus = http.StatusBadRequest
			message = "Invalid or corrupted image file."

		// HTTP 500 errors - system failures
		case goErr.HasTag(errors.ErrTagImageProcessingFailed):
			httpStatus = http.StatusInternalServerError
			message = "Failed to process image"
		case goErr.HasTag(errors.ErrTagImageRetrievalFailed):
			httpStatus = http.StatusInternalServerError
			message = "Failed to retrieve image"

		default:
			httpStatus = http.StatusInternalServerError
			message = "Internal server error"
		}
	} else {
		// Fallback for non-goerr errors
		httpStatus = http.StatusInternalServerError
		message = "Internal server error"
	}

	http.Error(w, message, httpStatus)
}
