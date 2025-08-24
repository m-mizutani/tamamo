package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/cli/config"
	"github.com/urfave/cli/v3"
)

// CmdGenerateConfig returns the generate-config command
func CmdGenerateConfig() *cli.Command {
	return &cli.Command{
		Name:    "generate-config",
		Aliases: []string{"g"},
		Usage:   "Generate configuration file templates",
		Commands: []*cli.Command{
			cmdGenerateLLMProviders(),
		},
	}
}

func cmdGenerateLLMProviders() *cli.Command {
	var (
		outputPath string
		force      bool
	)

	return &cli.Command{
		Name:    "llm-providers",
		Aliases: []string{"llm"},
		Usage:   "Generate LLM providers configuration template",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Usage:       "Output file path",
				Value:       "providers.yaml",
				Destination: &outputPath,
			},
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Usage:       "Overwrite existing file",
				Destination: &force,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger := ctxlog.From(ctx)

			// Check if file exists
			if _, err := os.Stat(outputPath); err == nil && !force {
				return goerr.New("file already exists, use --force to overwrite", goerr.V("path", outputPath))
			}

			// Generate the config file
			if err := config.GenerateConfigFile(outputPath); err != nil {
				return err
			}

			logger.Info("LLM providers configuration template generated successfully",
				"path", outputPath,
			)
			fmt.Printf("✅ LLM providers configuration template generated: %s\n", outputPath)
			fmt.Println("\nNext steps:")
			fmt.Println("1. Edit the file to customize providers and models")
			fmt.Println("2. Set up API keys via environment variables or CLI flags")
			fmt.Println("3. Use --llm-config flag to load the configuration")

			return nil
		},
	}
}
