package apperr

import "github.com/m-mizutani/goerr/v2"

// NotFound errors (HTTP 404)
var (
	ErrTagNotFound          = goerr.NewTag("not_found")
	ErrTagAgentNotFound     = goerr.NewTag("agent_not_found")
	ErrTagAgentNoImage      = goerr.NewTag("agent_no_image")
	ErrTagImageNotFound     = goerr.NewTag("image_not_found")
	ErrTagThumbnailNotFound = goerr.NewTag("thumbnail_not_found")
	ErrTagThreadNotFound    = goerr.NewTag("thread_not_found")
	ErrTagUserNotFound      = goerr.NewTag("user_not_found")
	ErrTagMessageNotFound   = goerr.NewTag("message_not_found")
)

// Validation errors (HTTP 400)
var (
	ErrTagValidation      = goerr.NewTag("validation")
	ErrTagInvalidInput    = goerr.NewTag("invalid_input")
	ErrTagInvalidFormat   = goerr.NewTag("invalid_format")
	ErrTagRequiredField   = goerr.NewTag("required_field")
	ErrTagInvalidFileType = goerr.NewTag("invalid_file_type")
	ErrTagImageTooLarge   = goerr.NewTag("image_too_large")
	ErrTagImageTooSmall   = goerr.NewTag("image_too_small")
	ErrTagCorruptedImage  = goerr.NewTag("corrupted_image")
)

// Permission errors (HTTP 401/403)
var (
	ErrTagUnauthorized = goerr.NewTag("unauthorized")
	ErrTagForbidden    = goerr.NewTag("forbidden")
	ErrTagExpiredToken = goerr.NewTag("expired_token")
)

// External service errors (HTTP 502/503)
var (
	ErrTagExternal  = goerr.NewTag("external")
	ErrTagSlackAPI  = goerr.NewTag("slack_api")
	ErrTagLLMError  = goerr.NewTag("llm_error")
	ErrTagFirestore = goerr.NewTag("firestore")
)

// System errors (HTTP 500)
var (
	ErrTagInternal              = goerr.NewTag("internal")
	ErrTagTimeout               = goerr.NewTag("timeout")
	ErrTagRateLimit             = goerr.NewTag("rate_limit")
	ErrTagImageProcessingFailed = goerr.NewTag("image_processing_failed")
	ErrTagImageRetrievalFailed  = goerr.NewTag("image_retrieval_failed")
)
