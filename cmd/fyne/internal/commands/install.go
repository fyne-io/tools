package commands

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/tools/cmd/fyne/internal/mobile"

	"github.com/urfave/cli/v2"

	//lint:ignore SA1019 The recommended replacement does not solve the use-case
	"golang.org/x/tools/go/vcs"
)

// Install returns the cli command for installing fyne applications
func Install() *cli.Command {
	i := NewInstaller()

	return &cli.Command{
		Name:      "install",
		Usage:     "Packages and installs an application.",
		UsageText: "fyne install [options] [remote[@tag]]",
		Description: "The install command packages an application for the current platform and copies it\n" +
			"into the system location for applications. This can be overridden with installDir",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "target",
				Aliases:     []string{"os"},
				Usage:       "The mobile platform to target (android, android/arm, android/arm64, android/amd64, android/386, ios, iossimulator).",
				Destination: &i.os,
			},
			&cli.StringFlag{
				Name:        "installDir",
				Aliases:     []string{"o"},
				Usage:       "A specific location to install to, rather than the OS default.",
				Destination: &i.installDir,
			},
			&cli.StringFlag{
				Name:        "icon",
				Usage:       "The name of the application icon file.",
				Value:       "",
				Destination: &i.icon,
			},
			&cli.BoolFlag{
				Name:        "use-raw-icon",
				Usage:       "Skip any OS-specific icon pre-processing",
				Value:       false,
				Destination: &i.rawIcon,
			},
			&cli.StringFlag{
				Name:        "appID",
				Aliases:     []string{"id"},
				Usage:       "For Android, darwin, iOS and Windows targets an appID in the form of a reversed domain name is required, for ios this must match a valid provisioning profile",
				Destination: &i.AppID,
			},
			&cli.BoolFlag{
				Name:        "release",
				Usage:       "Enable installation in release mode (disable debug, etc).",
				Destination: &i.release,
			},
		},
		Action: i.bundleAction,
	}
}

// Installer installs locally built Fyne apps.
type Installer struct {
	*appData
	installDir, srcDir, os string
	Packager               *Packager
	release                bool
}

// NewInstaller returns a command that can install a GUI apps built using Fyne from local source code.
func NewInstaller() *Installer {
	return &Installer{appData: &appData{}}
}

// AddFlags adds the flags for interacting with the Installer.
//
// Deprecated: Access to the individual cli commands are being removed.
func (i *Installer) AddFlags() {
	flag.StringVar(&i.os, "os", "", "The mobile platform to target (android, android/arm, android/arm64, android/amd64, android/386, ios)")
	flag.StringVar(&i.installDir, "installDir", "", "A specific location to install to, rather than the OS default")
	flag.StringVar(&i.icon, "icon", "Icon.png", "The name of the application icon file")
	flag.StringVar(&i.AppID, "appID", "", "For ios or darwin targets an appID is required, for ios this must \nmatch a valid provisioning profile")
	flag.BoolVar(&i.release, "release", false, "Should this package be installed in release mode? (disable debug etc)")
}

// PrintHelp prints the help for the install command.
//
// Deprecated: Access to the individual cli commands are being removed.
func (i *Installer) PrintHelp(indent string) {
	fmt.Println(indent, "The install command packages an application for the current platform and copies it")
	fmt.Println(indent, "into the system location for applications. This can be overridden with installDir")
	fmt.Println(indent, "Command usage: fyne install [parameters]")
}

// Run runs the install command.
//
// Deprecated: A better version will be exposed in the future.
func (i *Installer) Run(args []string) {
	if len(args) != 0 {
		fyne.LogError("Unexpected parameter after flags", nil)
		return
	}

	err := i.validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	err = i.install()
	if err != nil {
		fyne.LogError("Unable to install application", err)
		os.Exit(1)
	}
}

func (i *Installer) bundleAction(ctx *cli.Context) error {
	if ctx.Args().Len() != 0 {
		return i.installRemote(ctx)
	}

	return i.installLocal(ctx)
}

func (i *Installer) installLocal(ctx *cli.Context) error {
	err := i.validate()
	if err != nil {
		return err
	}

	err = i.install()
	if err != nil {
		return err
	}

	return nil
}

func (i *Installer) installRemote(ctx *cli.Context) error {
	pkg := ctx.Args().Slice()[0]
	branch := ""

	if parts := strings.SplitN(pkg, "@", 2); len(parts) == 2 {
		pkg = parts[0]
		branch = parts[1]
	}

	wd, _ := os.Getwd()
	defer func() {
		if wd != "" {
			os.Chdir(wd)
		}
	}()

	name := filepath.Base(pkg)
	path, err := os.MkdirTemp("", fmt.Sprintf("fyne-get-%s-*", name))
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(path)

	repo, err := vcs.RepoRootForImportPath(pkg, false)
	if err != nil {
		return fmt.Errorf("failed to look up source control for package: %w", err)
	}
	if repo.VCS.Name != "Git" {
		return errors.New("failed to find git repository: " + repo.VCS.Name)
	}

	args := []string{"clone", repo.Repo, "--depth=1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, path)

	cmd := exec.Command("git", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	if !util.Exists(path) { // the error above may be ignorable, unless the path was not found
		return fmt.Errorf("path doesn't exist: %v", err)
	}

	if repo.Root != pkg {
		dir := strings.Replace(pkg, repo.Root, "", 1)
		path = filepath.Join(path, dir)
	}

	install := &Installer{appData: i.appData, installDir: i.installDir, srcDir: path, release: true}
	if err := install.validate(); err != nil {
		return fmt.Errorf("failed to set up installer: %w", err)
	}

	return install.install()
}

func (i *Installer) install() error {
	p := i.Packager

	if i.os != "" {
		if util.IsIOS(i.os) {
			return i.installIOS()
		} else if strings.Index(i.os, "android") == 0 {
			return i.installAndroid()
		}

		return errors.New("Unsupported target operating system \"" + i.os + "\"")
	}

	if i.installDir == "" {
		switch p.os {
		case "darwin":
			i.installDir = "/Applications"
		case "linux", "openbsd", "freebsd", "netbsd":
			i.installDir = "/" // the tarball contains the structure starting at usr/local
		case "windows":
			dirName := p.Name
			if filepath.Ext(p.Name) == ".exe" {
				dirName = p.Name[:len(p.Name)-4]
			}
			i.installDir = filepath.Join(os.Getenv("ProgramFiles"), dirName)
			err := runAsAdminWindows("mkdir", i.installDir)
			if err != nil {
				fyne.LogError("Failed to run as windows administrator", err)
				return err
			}
		default:
			return errors.New("Unsupported target operating system \"" + p.os + "\"")
		}
	}

	p.dir = i.installDir
	err := p.doPackage(nil)
	if err != nil {
		return err
	}

	return postInstall(i)
}

func (i *Installer) installAndroid() error {
	target := mobile.AppOutputName(i.os, i.Packager.Name, i.release)

	_, err := os.Stat(target)
	if os.IsNotExist(err) {
		err := i.Packager.doPackage(nil)
		if err != nil {
			return nil
		}
	}

	return i.runMobileInstall("adb", target, "install")
}

func (i *Installer) installIOS() error {
	target := mobile.AppOutputName(i.os, i.Packager.Name, i.release)

	// Always redo the package because the codesign for ios and iossimulator
	// must be different.
	if err := i.Packager.doPackage(nil); err != nil {
		return nil
	}

	switch i.os {
	case "ios":
		return i.runMobileInstall("ios-deploy", target, "--bundle")
	case "iossimulator":
		return i.installToIOSSimulator(target)
	default:
		return fmt.Errorf("unsupported install target: %s", target)
	}
}

func (i *Installer) runMobileInstall(tool, target string, args ...string) error {
	_, err := exec.LookPath(tool)
	if err != nil {
		return err
	}

	cmd := exec.Command(tool, append(args, target)...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func (i *Installer) validate() error {
	os := i.os
	if os == "" {
		os = targetOS()
	}
	i.Packager = &Packager{appData: i.appData, os: os, install: true, srcDir: i.srcDir}
	i.Packager.AppID = i.AppID
	i.Packager.icon = i.icon
	i.Packager.release = i.release
	return i.Packager.validate()
}

func (i *Installer) installToIOSSimulator(target string) error {
	cmd := exec.Command(
		"xcrun", "simctl", "install",
		"booted", // Install to the booted simulator.
		target)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Install to a simulator error: %s%s", out, err)
	}

	i.runInIOSSimulator()
	return nil
}

func (i *Installer) runInIOSSimulator() error {
	cmd := exec.Command("xcrun", "simctl", "launch", "booted", i.Packager.AppID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.Stderr.Write(out)
		return err
	}
	return nil
}
