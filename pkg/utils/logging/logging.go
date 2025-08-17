package logging

import (
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/fatih/color"
	"github.com/m-mizutani/clog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/masq"
)

// Format represents the logging output format
type Format int

const (
	FormatConsole Format = iota + 1
	FormatJSON
)

var (
	defaultLogger = slog.Default()
	loggerMutex   sync.Mutex
)

// Default returns the default logger
func Default() *slog.Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	return defaultLogger
}

// SetDefault sets the default logger
func SetDefault(logger *slog.Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	defaultLogger = logger
	slog.SetDefault(logger)
}

// Quiet sets the logger to discard all output
func Quiet() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	SetDefault(logger)
	return logger
}

// New creates a new slog.Logger with the specified configuration
func New(w io.Writer, level slog.Level, format Format, stacktrace bool) *slog.Logger {
	// Configure masq filter for sensitive data
	filter := createMasqFilter()

	// Configure attribute hook based on stacktrace setting
	attrHook := clog.GoerrHook
	if !stacktrace {
		attrHook = goerrNoStacktrace
	}

	var handler slog.Handler
	switch format {
	case FormatConsole:
		handler = createConsoleHandler(w, level, filter, attrHook)
	case FormatJSON:
		handler = createJSONHandler(w, level, filter)
	default:
		panic(fmt.Sprintf("unsupported log format: %d", format))
	}

	return slog.New(handler)
}

// createMasqFilter creates a filter for masking sensitive data
func createMasqFilter() func([]string, slog.Attr) slog.Attr {
	return masq.New(
		masq.WithTag("secret"),
		masq.WithFieldPrefix("secret_"),
		masq.WithFieldPrefix("password_"),
		masq.WithFieldName("Authorization"),
		masq.WithFieldName("Token"),
		masq.WithFieldName("Password"),
		masq.WithFieldName("ApiKey"),
		masq.WithFieldName("Secret"),
	)
}

// createConsoleHandler creates a console handler with colors
func createConsoleHandler(w io.Writer, level slog.Level, filter func([]string, slog.Attr) slog.Attr, attrHook func([]string, slog.Attr) *clog.HandleAttr) slog.Handler {
	return clog.New(
		clog.WithWriter(w),
		clog.WithLevel(level),
		clog.WithReplaceAttr(filter),
		clog.WithAttrHook(attrHook),
		clog.WithColorMap(defaultColorMap()),
	)
}

// createJSONHandler creates a JSON handler
func createJSONHandler(w io.Writer, level slog.Level, filter func([]string, slog.Attr) slog.Attr) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: filter,
	})
}

// defaultColorMap returns the default color mapping for console output
func defaultColorMap() *clog.ColorMap {
	return &clog.ColorMap{
		Level: map[slog.Level]*color.Color{
			slog.LevelDebug: color.New(color.FgGreen, color.Bold),
			slog.LevelInfo:  color.New(color.FgCyan, color.Bold),
			slog.LevelWarn:  color.New(color.FgYellow, color.Bold),
			slog.LevelError: color.New(color.FgRed, color.Bold),
		},
		LevelDefault: color.New(color.FgBlue, color.Bold),
		Time:         color.New(color.FgWhite),
		Message:      color.New(color.FgHiWhite),
		AttrKey:      color.New(color.FgHiCyan),
		AttrValue:    color.New(color.FgHiWhite),
	}
}

// goerrNoStacktrace removes stacktrace from goerr errors for cleaner output
func goerrNoStacktrace(_ []string, attr slog.Attr) *clog.HandleAttr {
	goErr, ok := attr.Value.Any().(*goerr.Error)
	if !ok {
		return nil
	}

	// Extract error values without stacktrace
	var attrs []any
	for k, v := range goErr.Values() {
		attrs = append(attrs, slog.Any(k, v))
	}

	// Add the error message
	attrs = append(attrs, slog.String("message", goErr.Error()))

	// If there's a cause, add it
	if cause := goErr.Unwrap(); cause != nil {
		attrs = append(attrs, slog.Any("cause", cause))
	}

	newAttr := slog.Group(attr.Key, attrs...)
	return &clog.HandleAttr{
		NewAttr: &newAttr,
	}
}

// ErrAttr creates an error attribute for logging
func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}
