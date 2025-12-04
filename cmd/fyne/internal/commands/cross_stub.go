//go:build !go1.24

package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

type crossCmd struct{}

// Cross returns the command for cross-compiling Fyne applications using containers.
func Cross() *cli.Command {
	return &cli.Command{
		Name:    "cross",
		Aliases: []string{"x"},
		Usage:   "Cross-compile Fyne applications using containers",
		Flags:   []cli.Flag{},
		Action: func(ctx *cli.Context) error {
			return fmt.Errorf("cross command requires Go 1.24 or newer")
		},
	}
}
