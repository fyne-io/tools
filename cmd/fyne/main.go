// Run a command line helper for various Fyne tools.
package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"

	"fyne.io/tools/cmd/fyne/internal/commands"
)

func main() {
	app := &cli.App{
		Name:        "fyne",
		Usage:       "A command line helper for various Fyne tools.",
		Description: "The fyne command provides tooling for fyne applications and to assist in their development.",
		Commands: []*cli.Command{
			commands.Init(),
			commands.Env(),
			commands.Build(),
			commands.Package(),
			commands.Release(),
			commands.Install(),
			commands.Serve(),
			commands.Translate(),
			commands.Version(),
			commands.Bundle(),
		},
	}

	info, ok := debug.ReadBuildInfo()
	if ok {
		app.Version = info.Main.Version
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
