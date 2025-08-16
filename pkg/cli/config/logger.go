package config

import (
	"log/slog"
	"os"

	"github.com/m-mizutani/masq"
	"github.com/urfave/cli/v3"
)

type Logger struct {
	Level string
}

func (x *Logger) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "Log level (debug, info, warn, error)",
			Value:       "info",
			Sources:     cli.EnvVars("TAMAMO_LOG_LEVEL"),
			Destination: &x.Level,
		},
	}
}

func (x *Logger) Configure() (*slog.Logger, error) {
	var level slog.Level
	switch x.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,

		ReplaceAttr: masq.New(
			masq.WithTag("secret"),
		),
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger, nil
}
