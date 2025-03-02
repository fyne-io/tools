package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/lucor/goinfo"
	"github.com/lucor/goinfo/format"
	"github.com/lucor/goinfo/report"
	"github.com/urfave/cli/v2"
)

const fyneModule = "fyne.io/fyne/v2"

// Env returns the env command
func Env() *cli.Command {
	return &cli.Command{
		Name:    "env",
		Aliases: []string{"e"},
		Usage:   "Prints the Fyne module and environment information",
		Action: func(_ *cli.Context) error {
			workDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not get the path for the current working dir: %v", err)
			}

			workDir, err = lookupDirWithGoMod(workDir)
			if err != nil {
				return fmt.Errorf("failed to find go.mod: %v", err)
			}

			reporters := []goinfo.Reporter{
				&fyneReport{GoMod: &report.GoMod{WorkDir: workDir, Module: fyneModule}},
				&report.GoVersion{},
				&report.GoEnv{Filter: []string{"GOOS", "GOARCH", "CGO_ENABLED", "GO111MODULE"}},
				&report.OS{},
			}

			err = goinfo.Write(os.Stdout, reporters, &format.Text{})
			if err != nil {
				return err
			}

			return nil
		},
	}
}

// fyneReport defines a custom report for fyne
type fyneReport struct {
	*report.GoMod
}

// Info returns the collected info
func (r *fyneReport) Info() (goinfo.Info, error) {
	info, err := r.GoMod.Info()
	if err != nil {
		return info, err
	}
	// remove info for the report
	delete(info, "module")

	binfo, ok := debug.ReadBuildInfo()
	if !ok {
		info["cli_version"] = "could not retrieve version information (ensure module support is activated and build again)"
	} else {
		info["cli_version"] = binfo.Main.Version
	}

	return info, nil
}

// lookupDirWithGoMod takes a directory and checks for a go.mod file, traverses back towards the root,
// and returns the first directory with a match
func lookupDirWithGoMod(workDir string) (string, error) {
	for {
		fi, err := os.Stat(filepath.Join(workDir, "go.mod"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		if fi != nil {
			break
		}
		parentDir := filepath.Dir(workDir)
		if parentDir == workDir {
			return "", os.ErrNotExist
		}
		workDir = parentDir
	}
	return workDir, nil
}
