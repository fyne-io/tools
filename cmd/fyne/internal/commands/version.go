package commands

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/lucor/goinfo/report"
	"github.com/urfave/cli/v2"
)

// Version returns the cli command for the program version.
func Version() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Shows version information for fyne",
		Action: func(_ *cli.Context) error {
			info, ok := debug.ReadBuildInfo()
			if !ok {
				return errors.New("could not retrieve version information (ensure module support is activated and build again)")
			}
			fmt.Println("fyne cli version:", info.Main.Version)

			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			ver := "(not found)"
			if wd, err := lookupDirWithGoMod(wd); err == nil {
				wdInfo := &report.GoMod{WorkDir: wd, Module: fyneModule}
				info2, err := wdInfo.Info()
				if err != nil {
					return err
				}

				if imported, ok := info2["imported"]; ok && imported.(bool) {
					ver = info2["version"].(string)
				}
			}

			fmt.Println("fyne library version:", ver)
			return nil
		},
	}
}
