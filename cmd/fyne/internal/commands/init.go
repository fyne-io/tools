package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/natefinch/atomic"
	"github.com/urfave/cli/v2"
)

const codeFmt = `package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID(%q)
	w := a.NewWindow("Hello World")

	w.SetContent(widget.NewLabel("Hello World!"))
	w.ShowAndRun()
}
`

const tomlFmt = `[Details]
Icon = "Icon.png"
Name = %q
ID = %q
Version = "0.0.1"
Build = 1
`

func Init() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Args:      true,
		Usage:     "Initializes a new Fyne project in the current directory, including a go.mod file and a FyneApp.toml",
		UsageText: "fyne init module-path [app-id [app-name]]",
		Action: func(ctx *cli.Context) error {
			modpath := ctx.Args().Get(0)
			appid := ctx.Args().Get(1)
			appname := ctx.Args().Get(2)

			// something like: github.com/bla/foo
			// or a local app: mysuperapp
			if modpath == "" {
				return cli.ShowSubcommandHelp(ctx)
			}

			if appid == "" {
				// get
				appid = "some.fake"
			}

			if appname == "" {
				//
				appname = "Some Fake"
			}

			// check for main.go
			fi, err := os.Stat("main.go")
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}

			if fi == nil {
				//create main.go
				code := fmt.Sprintf(codeFmt, appid)

				if err := atomic.WriteFile("main.go", strings.NewReader(code)); err != nil {
					return err
				}
				os.Chmod("main.go", 0644)
			}

			// check for go.mod
			fi, err = os.Stat("go.mod")
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}

			if fi == nil {
				cmd := exec.Command("go", "mod", "init", modpath)
				cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

				if err = cmd.Run(); err != nil {
					return fmt.Errorf("failed to run command: %v", err)
				}
			}

			// create toml
			fi, err = os.Stat("FyneApp.toml")
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if fi == nil {
				code := fmt.Sprintf(tomlFmt, appname, appid)

				if err := atomic.WriteFile("FyneApp.toml", strings.NewReader(code)); err != nil {
					return err
				}
				os.Chmod("FyneApp.toml", 0644)
			}

			// run go mod tidy
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

			if err = cmd.Run(); err != nil {
				return fmt.Errorf("failed to run command: %v", err)
			}

			return nil
		},
	}
}
