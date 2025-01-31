package commands

import (
	"github.com/urfave/cli/v2"
)

// Get returns the command which downloads and installs fyne applications.
func Get() *cli.Command {
	cmd := Install()
	cmd.Name = "get"
	cmd.Hidden = true

	return cmd
}
