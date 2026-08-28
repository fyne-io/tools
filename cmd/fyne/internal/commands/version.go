package commands

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/lucor/goinfo/report"
	"github.com/urfave/cli/v2"
)

func getFyneGoModVersion(dir string) (string, error) {
	wd, err := lookupDirWithGoMod(dir)
	if err != nil {
		return "", err
	}

	wdInfo := &report.GoMod{WorkDir: wd, Module: fyneModule}
	if info, err := wdInfo.Info(); err != nil {
		return "", err
	} else if imported, ok := info["imported"]; ok && imported.(bool) {
		return info["version"].(string), nil
	}

	return "", fmt.Errorf("fyne version not found")
}

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
			if v, err := getFyneGoModVersion(wd); err == nil {
				ver = v
			}

			fmt.Println("fyne library version:", ver)
			return nil
		},
	}
}
