package apperr

import (
	"net/http"

	"github.com/m-mizutani/goerr/v2"
)

// HTTPStatusFromError returns the appropriate HTTP status code based on error tags
func HTTPStatusFromError(err error) int {
	switch {
	// 404 Not Found
	case goerr.HasTag(err, ErrTagNotFound),
		goerr.HasTag(err, ErrTagAgentNotFound),
		goerr.HasTag(err, ErrTagThreadNotFound),
		goerr.HasTag(err, ErrTagUserNotFound),
		goerr.HasTag(err, ErrTagMessageNotFound),
		goerr.HasTag(err, ErrTagImageNotFound),
		goerr.HasTag(err, ErrTagThumbnailNotFound),
		goerr.HasTag(err, ErrTagAgentNoImage):
		return http.StatusNotFound

	// 400 Bad Request
	case goerr.HasTag(err, ErrTagValidation),
		goerr.HasTag(err, ErrTagInvalidInput),
		goerr.HasTag(err, ErrTagInvalidFormat),
		goerr.HasTag(err, ErrTagRequiredField),
		goerr.HasTag(err, ErrTagInvalidFileType),
		goerr.HasTag(err, ErrTagImageTooLarge),
		goerr.HasTag(err, ErrTagImageTooSmall),
		goerr.HasTag(err, ErrTagCorruptedImage):
		return http.StatusBadRequest

	// 401 Unauthorized
	case goerr.HasTag(err, ErrTagUnauthorized),
		goerr.HasTag(err, ErrTagExpiredToken):
		return http.StatusUnauthorized

	// 403 Forbidden
	case goerr.HasTag(err, ErrTagForbidden):
		return http.StatusForbidden

	// 408 Request Timeout
	case goerr.HasTag(err, ErrTagTimeout):
		return http.StatusRequestTimeout

	// 429 Too Many Requests
	case goerr.HasTag(err, ErrTagRateLimit):
		return http.StatusTooManyRequests

	// 502 Bad Gateway
	case goerr.HasTag(err, ErrTagExternal),
		goerr.HasTag(err, ErrTagSlackAPI),
		goerr.HasTag(err, ErrTagLLMError),
		goerr.HasTag(err, ErrTagFirestore):
		return http.StatusBadGateway

	// 500 Internal Server Error (default)
	default:
		return http.StatusInternalServerError
	}
}
