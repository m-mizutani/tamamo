package image

import (
	"fmt"
	"image"
	_ "image/jpeg" // Import for JPEG support
	_ "image/png"  // Import for PNG support
	"io"
	"net/http"
	"strings"

	imageModel "github.com/m-mizutani/tamamo/pkg/domain/model/image"
)

// Validator provides image validation functionality
type Validator struct {
	rules imageModel.ImageValidationRules
}

// NewValidator creates a new image validator with default rules
func NewValidator() *Validator {
	return &Validator{
		rules: imageModel.DefaultValidationRules(),
	}
}

// NewValidatorWithRules creates a new image validator with custom rules
func NewValidatorWithRules(rules imageModel.ImageValidationRules) *Validator {
	return &Validator{
		rules: rules,
	}
}

// ValidateFile validates a complete image file
func (v *Validator) ValidateFile(file io.ReadSeeker, contentType string, fileSize int64) (*ImageMetadata, error) {
	// Validate file size
	if err := v.validateFileSize(fileSize); err != nil {
		return nil, err
	}

	// Validate MIME type
	if err := v.validateMimeType(contentType); err != nil {
		return nil, err
	}

	// Detect actual MIME type from file content
	actualMimeType, err := v.detectMimeType(file)
	if err != nil {
		return nil, err
	}

	// Verify MIME type consistency
	if !v.isMimeTypeConsistent(contentType, actualMimeType) {
		return nil, imageModel.ErrInvalidMimeType
	}

	// Reset file position for image decoding
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file position: %w", err)
	}

	// Decode image to get dimensions
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, imageModel.ErrCorruptedImage
	}

	// Validate dimensions
	if err := v.validateDimensions(config.Width, config.Height); err != nil {
		return nil, err
	}

	return &ImageMetadata{
		ContentType: actualMimeType,
		FileSize:    fileSize,
		Width:       config.Width,
		Height:      config.Height,
		Format:      format,
	}, nil
}

// validateFileSize validates the file size against rules
func (v *Validator) validateFileSize(size int64) error {
	if size > v.rules.MaxFileSize {
		return imageModel.ErrImageTooLarge
	}
	return nil
}

// validateMimeType validates the MIME type against allowed types
func (v *Validator) validateMimeType(mimeType string) error {
	for _, allowed := range v.rules.AllowedMimeTypes {
		if mimeType == allowed {
			return nil
		}
	}
	return imageModel.ErrInvalidMimeType
}

// validateDimensions validates image dimensions against rules
func (v *Validator) validateDimensions(width, height int) error {
	if width < v.rules.MinWidth || height < v.rules.MinHeight {
		return imageModel.ErrImageTooSmall
	}
	if width > v.rules.MaxWidth || height > v.rules.MaxHeight {
		return imageModel.ErrImageDimensionsTooLarge
	}
	return nil
}

// detectMimeType detects the MIME type from file content
func (v *Validator) detectMimeType(file io.ReadSeeker) (string, error) {
	// Read first 512 bytes for MIME type detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for MIME detection: %w", err)
	}

	// Detect MIME type
	mimeType := http.DetectContentType(buffer[:n])

	// Normalize MIME type
	switch {
	case strings.HasPrefix(mimeType, "image/jpeg"):
		return "image/jpeg", nil
	case strings.HasPrefix(mimeType, "image/png"):
		return "image/png", nil
	default:
		return "", imageModel.ErrInvalidImageFormat
	}
}

// isMimeTypeConsistent checks if the declared and detected MIME types are consistent
func (v *Validator) isMimeTypeConsistent(declared, detected string) bool {
	// Normalize both types
	declared = strings.ToLower(strings.TrimSpace(declared))
	detected = strings.ToLower(strings.TrimSpace(detected))

	// Direct match
	if declared == detected {
		return true
	}

	// Handle JPEG variations
	if (declared == "image/jpeg" || declared == "image/jpg") &&
		(detected == "image/jpeg" || detected == "image/jpg") {
		return true
	}

	return false
}

// ImageMetadata holds metadata about a validated image
type ImageMetadata struct {
	ContentType string
	FileSize    int64
	Width       int
	Height      int
	Format      string
}
