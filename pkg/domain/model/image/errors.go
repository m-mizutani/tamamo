package image

import "errors"

var (
	// ErrInvalidImageFormat indicates an unsupported image format
	ErrInvalidImageFormat = errors.New("invalid image format")

	// ErrImageTooLarge indicates the image file is too large
	ErrImageTooLarge = errors.New("image file too large")

	// ErrImageTooSmall indicates the image dimensions are too small
	ErrImageTooSmall = errors.New("image dimensions too small")

	// ErrImageDimensionsTooLarge indicates the image dimensions are too large
	ErrImageDimensionsTooLarge = errors.New("image dimensions too large")

	// ErrImageNotFound indicates the image was not found
	ErrImageNotFound = errors.New("image not found")

	// ErrStorageUnavailable indicates storage is unavailable
	ErrStorageUnavailable = errors.New("storage unavailable")

	// ErrSecurityViolation indicates a security violation was detected
	ErrSecurityViolation = errors.New("security violation detected")

	// ErrInvalidMimeType indicates an invalid MIME type
	ErrInvalidMimeType = errors.New("invalid MIME type")

	// ErrCorruptedImage indicates the image file is corrupted
	ErrCorruptedImage = errors.New("corrupted image file")
)

// ImageValidationRules defines the validation rules for images
type ImageValidationRules struct {
	MaxFileSize      int64    // Maximum file size in bytes
	MinWidth         int      // Minimum width in pixels
	MinHeight        int      // Minimum height in pixels
	MaxWidth         int      // Maximum width in pixels
	MaxHeight        int      // Maximum height in pixels
	AllowedMimeTypes []string // Allowed MIME types
}

// DefaultValidationRules returns the default validation rules
func DefaultValidationRules() ImageValidationRules {
	return ImageValidationRules{
		MaxFileSize:      10 * 1024 * 1024, // 10MB
		MinWidth:         64,
		MinHeight:        64,
		MaxWidth:         1024,
		MaxHeight:        1024,
		AllowedMimeTypes: []string{"image/jpeg", "image/png"},
	}
}
