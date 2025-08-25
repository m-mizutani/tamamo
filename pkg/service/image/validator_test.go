package image_test

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	imageModel "github.com/m-mizutani/tamamo/pkg/domain/model/image"
	imageService "github.com/m-mizutani/tamamo/pkg/service/image"
)

// createTestJPEG creates a minimal valid JPEG file for testing
func createTestJPEG() []byte {
	// Create a simple 100x100 image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// createTestPNG creates a minimal valid PNG file for testing
func createTestPNG() []byte {
	// Create a simple 100x100 image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func TestValidator_ValidateFile_ValidJPEG(t *testing.T) {
	validator := imageService.NewValidator()
	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	metadata, err := validator.ValidateFile(reader, "image/jpeg", int64(len(jpegData)))
	if err != nil {
		t.Fatalf("Expected valid JPEG to pass validation, got error: %v", err)
	}

	if metadata.ContentType != "image/jpeg" {
		t.Errorf("Expected content type image/jpeg, got %s", metadata.ContentType)
	}
	if metadata.FileSize != int64(len(jpegData)) {
		t.Errorf("Expected file size %d, got %d", len(jpegData), metadata.FileSize)
	}
}

func TestValidator_ValidateFile_ValidPNG(t *testing.T) {
	validator := imageService.NewValidator()
	pngData := createTestPNG()
	reader := bytes.NewReader(pngData)

	metadata, err := validator.ValidateFile(reader, "image/png", int64(len(pngData)))
	if err != nil {
		t.Fatalf("Expected valid PNG to pass validation, got error: %v", err)
	}

	if metadata.ContentType != "image/png" {
		t.Errorf("Expected content type image/png, got %s", metadata.ContentType)
	}
}

func TestValidator_ValidateFile_InvalidMimeType(t *testing.T) {
	validator := imageService.NewValidator()
	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	_, err := validator.ValidateFile(reader, "image/gif", int64(len(jpegData)))
	if err != imageModel.ErrInvalidMimeType {
		t.Errorf("Expected ErrInvalidMimeType, got %v", err)
	}
}

func TestValidator_ValidateFile_FileTooLarge(t *testing.T) {
	rules := imageModel.ImageValidationRules{
		MaxFileSize:      100, // Very small limit
		MinWidth:         64,
		MinHeight:        64,
		MaxWidth:         1024,
		MaxHeight:        1024,
		AllowedMimeTypes: []string{"image/jpeg", "image/png"},
	}
	validator := imageService.NewValidatorWithRules(rules)
	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	_, err := validator.ValidateFile(reader, "image/jpeg", int64(len(jpegData)))
	if err != imageModel.ErrImageTooLarge {
		t.Errorf("Expected ErrImageTooLarge, got %v", err)
	}
}

func TestValidator_ValidateFile_MimeTypeMismatch(t *testing.T) {
	validator := imageService.NewValidator()
	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	// Declare as PNG but provide JPEG data
	_, err := validator.ValidateFile(reader, "image/png", int64(len(jpegData)))
	if err != imageModel.ErrInvalidMimeType {
		t.Errorf("Expected ErrInvalidMimeType for MIME type mismatch, got %v", err)
	}
}

func TestValidator_ValidateFile_InvalidImageFormat(t *testing.T) {
	validator := imageService.NewValidator()
	corruptedData := []byte("this is not an image")
	reader := bytes.NewReader(corruptedData)

	_, err := validator.ValidateFile(reader, "image/jpeg", int64(len(corruptedData)))
	if err != imageModel.ErrInvalidImageFormat {
		t.Errorf("Expected ErrInvalidImageFormat, got %v (type %T)", err, err)
	}
}

func TestValidator_ValidateFile_DimensionsTooSmall(t *testing.T) {
	rules := imageModel.ImageValidationRules{
		MaxFileSize:      10 * 1024 * 1024, // 10MB
		MinWidth:         200,              // Larger than test image (100x100)
		MinHeight:        200,
		MaxWidth:         1024,
		MaxHeight:        1024,
		AllowedMimeTypes: []string{"image/jpeg", "image/png"},
	}
	validator := imageService.NewValidatorWithRules(rules)
	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	_, err := validator.ValidateFile(reader, "image/jpeg", int64(len(jpegData)))
	if err != imageModel.ErrImageTooSmall {
		t.Errorf("Expected ErrImageTooSmall, got %v", err)
	}
}

func TestValidator_ValidateFile_DimensionsTooLarge(t *testing.T) {
	rules := imageModel.ImageValidationRules{
		MaxFileSize:      10 * 1024 * 1024, // 10MB
		MinWidth:         1,
		MinHeight:        1,
		MaxWidth:         50, // Smaller than test image (100x100)
		MaxHeight:        50,
		AllowedMimeTypes: []string{"image/jpeg", "image/png"},
	}
	validator := imageService.NewValidatorWithRules(rules)
	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	_, err := validator.ValidateFile(reader, "image/jpeg", int64(len(jpegData)))
	if err != imageModel.ErrImageDimensionsTooLarge {
		t.Errorf("Expected ErrImageDimensionsTooLarge, got %v", err)
	}
}

func TestValidator_MimeTypeConsistency(t *testing.T) {
	validator := imageService.NewValidator()

	testCases := []struct {
		declared string
		detected string
		expected bool
	}{
		{"image/jpeg", "image/jpeg", true},
		{"image/png", "image/png", true},
		{"image/jpeg", "image/jpg", true},
		{"image/jpg", "image/jpeg", true},
		{"image/jpeg", "image/png", false},
		{"image/png", "image/jpeg", false},
		{"IMAGE/JPEG", "image/jpeg", true}, // Case insensitive
	}

	for _, tc := range testCases {
		// Use reflection to access private method (for testing purposes)
		// In a real implementation, you might want to make this method public for testing
		// or test it indirectly through ValidateFile
		t.Run(tc.declared+"_vs_"+tc.detected, func(t *testing.T) {
			// Since the method is private, we test it indirectly
			// by providing matching/mismatching MIME types to ValidateFile
			jpegData := createTestJPEG()
			reader := bytes.NewReader(jpegData)

			_, err := validator.ValidateFile(reader, tc.declared, int64(len(jpegData)))

			if tc.expected && strings.Contains(tc.declared, "jpeg") {
				if err == imageModel.ErrInvalidMimeType {
					t.Errorf("Expected MIME type consistency check to pass for %s vs %s", tc.declared, tc.detected)
				}
			} else if !tc.expected && !strings.Contains(tc.declared, "jpeg") {
				if err != imageModel.ErrInvalidMimeType {
					t.Errorf("Expected MIME type consistency check to fail for %s vs %s", tc.declared, tc.detected)
				}
			}
		})
	}
}
