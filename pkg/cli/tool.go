package cli

import (
	"github.com/m-mizutani/tamamo/pkg/cli/tools"
	"github.com/urfave/cli/v3"
)

func cmdTool() *cli.Command {
	return &cli.Command{
		Name:    "tool",
		Aliases: []string{"t"},
		Usage:   "Utility tools",
		Commands: []*cli.Command{
			tools.CmdGenerateConfig(),
		},
	}
}
