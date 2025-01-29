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
		Action:    initAction,
	}
}

func getAppID(modpath string) string {
	p := strings.Split(modpath, "/")
	if len(p) == 0 {
		return ""
	}

	d := strings.Split(p[0], ".")
	r := make([]string, len(p) + len(d) - 1)
	for n, e := range d {
		r[len(d)-n-1] = e
	}
	for n, e := range p {
		if n == 0 {
			continue
		}
		r[len(d)+n-1] = e
	}

	return strings.Join(r, ".")
}

func getAppName(modpath string) string {
	p := strings.Split(modpath, "/")
	if len(p) == 0 {
		return ""
	}

	if len(p) > 1 {
		return p[len(p)-1]
	}

	d := strings.Split(p[0], ".")

	return d[0]
}

func checkFileOrDo(file string, cb func() error) error {
	fi, err := os.Stat(file)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if fi != nil {
		return nil
	}

	return cb()
}

func checkFileOrCreate(file, content string) error {
	return checkFileOrDo(file, func() error {
		if err := atomic.WriteFile(file, strings.NewReader(content)); err != nil {
			return err
		}
		return os.Chmod(file, 0644)
	})
}

func initAction(ctx *cli.Context) error {
	modpath := ctx.Args().Get(0)
	appID := ctx.Args().Get(1)
	appName := ctx.Args().Get(2)

	if modpath == "" {
		return cli.ShowSubcommandHelp(ctx)
	}

	if appID == "" {
		appID = getAppID(modpath)
	}

	if appName == "" {
		appName = getAppName(modpath)
	}

	if err := checkFileOrCreate("main.go", fmt.Sprintf(codeFmt, appID)); err != nil {
		return err
	}

	if err := checkFileOrDo("go.mod", func() error {
		cmd := exec.Command("go", "mod", "init", modpath)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}); err != nil {
		return err
	}

	if err := checkFileOrCreate("FyneApp.toml", fmt.Sprintf(tomlFmt, appName, appID)); err != nil {
		return err
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	return nil
}
