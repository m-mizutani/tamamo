package config

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/utils/logging"
	"github.com/m-mizutani/tamamo/pkg/utils/safe"
	"github.com/urfave/cli/v3"
)

// Logger holds the configuration for logging
type Logger struct {
	level      string
	format     string
	output     string
	quiet      bool
	stacktrace bool
}

// Flags returns CLI flags for logger configuration
func (x *Logger) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Category:    "logging",
			Aliases:     []string{"l"},
			Sources:     cli.EnvVars("TAMAMO_LOG_LEVEL"),
			Usage:       "Set log level [debug|info|warn|error]",
			Value:       "info",
			Destination: &x.level,
		},
		&cli.StringFlag{
			Name:        "log-format",
			Category:    "logging",
			Aliases:     []string{"f"},
			Sources:     cli.EnvVars("TAMAMO_LOG_FORMAT"),
			Usage:       "Set log format [console|json]",
			Value:       "console",
			Destination: &x.format,
		},
		&cli.StringFlag{
			Name:        "log-output",
			Category:    "logging",
			Aliases:     []string{"o"},
			Sources:     cli.EnvVars("TAMAMO_LOG_OUTPUT"),
			Usage:       "Set log output (create file other than '-', 'stdout', 'stderr')",
			Value:       "stdout",
			Destination: &x.output,
		},
		&cli.BoolFlag{
			Name:        "log-quiet",
			Category:    "logging",
			Aliases:     []string{"q"},
			Usage:       "Quiet mode (no log output)",
			Sources:     cli.EnvVars("TAMAMO_LOG_QUIET"),
			Destination: &x.quiet,
		},
		&cli.BoolFlag{
			Name:        "log-stacktrace",
			Category:    "logging",
			Aliases:     []string{"s"},
			Usage:       "Show stacktrace (only for console format)",
			Sources:     cli.EnvVars("TAMAMO_LOG_STACKTRACE"),
			Destination: &x.stacktrace,
			Value:       true,
		},
	}
}

// LogValue returns the logger configuration as a slog.Value for logging
func (x Logger) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("level", x.level),
		slog.String("format", x.format),
		slog.String("output", x.output),
		slog.Bool("quiet", x.quiet),
		slog.Bool("stacktrace", x.stacktrace),
	)
}

// Configure sets up logger and returns the configured logger
func (x *Logger) Configure() (*slog.Logger, error) {
	// Handle quiet mode
	if x.quiet {
		return logging.Quiet(), nil
	}

	// Parse log level
	level, err := x.parseLevel()
	if err != nil {
		return nil, err
	}

	// Parse log format
	format, err := x.parseFormat()
	if err != nil {
		return nil, err
	}

	// Open output writer
	output, closer, err := x.openOutput()
	if err != nil {
		return nil, err
	}
	// Note: We're not returning the closer as the interface doesn't support it
	// This could be improved in the future
	_ = closer

	// Create and set the logger
	logger := logging.New(output, level, format, x.stacktrace)
	logging.SetDefault(logger)

	return logger, nil
}

// parseLevel converts the level string to slog.Level
func (x *Logger) parseLevel() (slog.Level, error) {
	levelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	level, ok := levelMap[strings.ToLower(x.level)]
	if !ok {
		return slog.LevelInfo, goerr.New("invalid log level",
			goerr.V("level", x.level),
			goerr.V("valid_levels", []string{"debug", "info", "warn", "error"}),
		)
	}

	return level, nil
}

// parseFormat determines the log format to use
func (x *Logger) parseFormat() (logging.Format, error) {
	// Auto-detect format if not specified
	if x.format == "" {
		return x.autoDetectFormat(), nil
	}

	// Parse explicit format
	formatMap := map[string]logging.Format{
		"console": logging.FormatConsole,
		"json":    logging.FormatJSON,
	}

	format, ok := formatMap[strings.ToLower(x.format)]
	if !ok {
		return logging.FormatConsole, goerr.New("invalid log format",
			goerr.V("format", x.format),
			goerr.V("valid_formats", []string{"console", "json"}),
		)
	}

	return format, nil
}

// autoDetectFormat detects the appropriate format based on the terminal
func (x *Logger) autoDetectFormat() logging.Format {
	term := os.Getenv("TERM")
	if strings.Contains(term, "color") || strings.Contains(term, "xterm") {
		return logging.FormatConsole
	}
	return logging.FormatJSON
}

// openOutput opens the output writer based on configuration
func (x *Logger) openOutput() (io.Writer, func(), error) {
	switch strings.ToLower(x.output) {
	case "stdout", "-":
		return os.Stdout, func() {}, nil

	case "stderr":
		return os.Stderr, func() {}, nil

	default:
		// Open file for logging
		f, err := os.OpenFile(filepath.Clean(x.output), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return nil, func() {}, goerr.Wrap(err, "failed to open log file",
				goerr.V("path", x.output),
			)
		}

		closer := func() {
			safe.Close(context.Background(), f)
		}

		return f, closer, nil
	}
}
