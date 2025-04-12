package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"fyne.io/tools/cmd/fyne/internal/metadata"
	"fyne.io/tools/cmd/fyne/internal/templates"
)

func Init() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Initializes a new Fyne project",
		ArgsUsage: "[module-path]",
		Action:    initAction,
		Description: "Initializes a new Fyne project in the current directory, including\n" +
			"a go.mod, main.go, and FyneApp.toml file (unless existing).",
		Flags: []cli.Flag{
			stringFlags["app-id"](nil),
			stringFlags["name"](nil),
			stringFlags["icon"](nil),
			boolFlags["verbose"](nil),
		},
	}
}

func getAppID(modpath string) string {
	p := strings.Split(modpath, "/")
	if len(p) == 0 {
		return ""
	}

	d := strings.Split(p[0], ".")
	r := make([]string, len(p)+len(d)-1)
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

func initAction(ctx *cli.Context) error {
	modpath := ctx.Args().Get(0)
	appID := ctx.String("app-id")
	appName := ctx.String("name")
	icon := ctx.String("icon")
	verbose := ctx.Bool("verbose")

	if modpath == "" {
		modpath = "example"

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		if wd != "" {
			modpath = filepath.Base(wd)
		}
	}

	if appID == "" {
		appID = getAppID(modpath)
	}

	if appName == "" {
		appName = getAppName(modpath)
	}

	data := &metadata.FyneApp{}
	data.Details.ID = appID
	data.Details.Name = appName
	data.Details.Icon = icon
	data.Details.Version = "0.0.0"
	data.Migrations = map[string]bool{
		"fyneDo": true,
	}

	if err := checkFileOrDo("main.go", func() error {
		f, err := os.Create("main.go")
		if err != nil {
			return err
		}
		defer f.Close()
		return templates.HelloWorld.Execute(f, &data)
	}); err != nil {
		return err
	}

	if err := checkFileOrDo("go.mod", func() error {
		cmd := exec.Command("go", "mod", "init", modpath)
		if verbose {
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		}
		return cmd.Run()
	}); err != nil {
		return err
	}

	if err := checkFileOrDo("FyneApp.toml", func() error {
		return metadata.SaveStandard(data, ".")
	}); err != nil {
		return err
	}

	cmd := exec.Command("go", "mod", "tidy")
	if verbose {
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	fmt.Println("Your new app is ready. Run it directly with: go run .")

	return nil
}
