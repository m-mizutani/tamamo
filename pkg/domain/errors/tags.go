package errors

import "github.com/m-mizutani/goerr/v2"

// Image-related error tags
var (
	// HTTP 404 errors
	ErrTagAgentNotFound     = goerr.NewTag("agent_not_found")
	ErrTagAgentNoImage      = goerr.NewTag("agent_no_image")
	ErrTagImageNotFound     = goerr.NewTag("image_not_found")
	ErrTagThumbnailNotFound = goerr.NewTag("thumbnail_not_found")

	// HTTP 400 errors - validation failures
	ErrTagInvalidFileType = goerr.NewTag("invalid_file_type")
	ErrTagImageTooLarge   = goerr.NewTag("image_too_large")
	ErrTagImageTooSmall   = goerr.NewTag("image_too_small")
	ErrTagCorruptedImage  = goerr.NewTag("corrupted_image")

	// HTTP 500 errors - system failures
	ErrTagImageProcessingFailed = goerr.NewTag("image_processing_failed")
	ErrTagImageRetrievalFailed  = goerr.NewTag("image_retrieval_failed")
)
