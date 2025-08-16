package http

import (
	"log/slog"
	"net/http"

	"github.com/m-mizutani/goerr"
)

// handleError handles errors and returns appropriate HTTP responses
func handleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Log the error with context
	slog.ErrorContext(r.Context(), "request error",
		"error", err,
		"path", r.URL.Path,
		"method", r.Method,
	)

	// Determine status code based on error type
	statusCode := http.StatusInternalServerError
	message := "Internal Server Error"

	// Check for specific error types
	if goErr, ok := err.(*goerr.Error); ok {
		// Could add specific error type handling here
		// For now, keep it simple
		_ = goErr // Prevent unused variable warning
	}

	http.Error(w, message, statusCode)
}
