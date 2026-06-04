package commands

import (
	"github.com/urfave/cli/v2"

	crossCommand "fyne.io/tools/cmd/fyne/internal/cross/command"
)

func Cross() *cli.Command {
	return &cli.Command{
		Name:        "cross",
		Aliases:     []string{"x"},
		Usage:       "Cross-compiles a Fyne application",
		Subcommands: []*cli.Command{
			crossCommand.DarwinSDKExtract(),
			crossCommand.Darwin(),
			crossCommand.Linux(),
			crossCommand.Windows(),
			crossCommand.Android(),
			crossCommand.IOS(),
			crossCommand.FreeBSD(),
			crossCommand.Web(),
			crossCommand.Version(),
		},
	}
}

