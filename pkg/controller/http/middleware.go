package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/utils/errors"
)

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		ctxlog.From(r.Context()).Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
		)
	})
}

// panicRecoveryMiddleware recovers from panics
func panicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace for debugging
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				panicErr := goerr.New(fmt.Sprintf("panic recovered: %v", err), goerr.V("stack", string(buf[:n])))
				errors.Handle(r.Context(), panicErr)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// verifySlackSignature verifies Slack request signatures
func verifySlackSignature(verifier slack.PayloadVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				err = goerr.Wrap(err, "failed to read request body")
				errors.Handle(r.Context(), err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			// Restore body for next handler
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Get headers
			timestamp := r.Header.Get("X-Slack-Request-Timestamp")
			signature := r.Header.Get("X-Slack-Signature")

			// Verify signature
			if err := verifier.Verify(body, timestamp, signature); err != nil {
				err = goerr.Wrap(err, "slack signature verification failed", goerr.V("status", http.StatusUnauthorized))
				errors.Handle(r.Context(), err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *responseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}
