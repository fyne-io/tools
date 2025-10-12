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
	"fyne.io/tools/cmd/fyne/internal/metadata"
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
		Aliases:   []string{"get", "i"},
		Usage:     "Packages and installs an application",
		ArgsUsage: "[remote[@version]]",
		Description: "The install command packages an application for the current platform and copies it\n" +
			"into the system location for applications by default.",
		Flags: []cli.Flag{
			stringFlags["target"](&i.os),
			stringFlags["dst"](&i.installDir),
			stringFlags["icon"](&i.icon),
			boolFlags["use-raw-icon"](&i.rawIcon),
			stringFlags["app-id"](&i.AppID),
			boolFlags["release"](&i.release),
			stringFlags["tags"](&i.tags),
			boolFlags["verbose"](&i.verbose),
		},
		Action: i.bundleAction,
	}
}

// Installer installs locally built Fyne apps.
type Installer struct {
	*appData
	installDir, srcDir, os string
	tags                   string
	Packager               *Packager
	release                bool
	verbose                bool
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
	flag.StringVar(&i.installDir, "dst", "", "A specific location to install to, rather than the OS default")
	flag.StringVar(&i.icon, "icon", "Icon.png", "The name of the application icon file")
	flag.StringVar(&i.AppID, "app-id", "", "For ios or darwin targets an app-id is required, for ios this must \nmatch a valid provisioning profile")
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
	arg := ctx.Args().Get(0)
	if arg == "" || strings.HasPrefix(arg, ".") {
		return i.installLocal(ctx)
	}

	return i.installRemote(ctx)
}

func (i *Installer) installLocal(ctx *cli.Context) error {
	if i.icon == "" {
		path := ctx.Args().Get(0)
		if path != "" {
			meta, err := metadata.LoadStandard(path)
			if err != nil {
				return err
			}
			i.srcDir = path
			i.icon = meta.Details.Icon
		}
	}

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

func getPackageAndBranch(s string) (string, string) {
	if pkg, branch, found := strings.Cut(s, "@"); found {
		return pkg, branch
	}
	return s, ""
}

func getLatestTag(repo string) (string, error) {
	cmd := exec.Command("git", "ls-remote", "-q", repo)
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}

	tag := ""
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}

		_, s, found := strings.Cut(fields[1], "refs/tags/")
		if !found || !strings.HasPrefix(s, "v") || strings.HasSuffix(s, "^{}") {
			continue
		}

		tag = s
	}

	return tag, nil
}

func getInstallBaseDir(path, pkg, root string) string {
	if len(pkg) <= len(root) || !strings.HasPrefix(pkg, root) || pkg == "" || root == "" {
		return path
	}
	return path + strings.TrimPrefix(pkg, root)
}

func (i *Installer) installRemote(ctx *cli.Context) error {
	pkg, branch := getPackageAndBranch(ctx.Args().Get(0))

	wd, _ := os.Getwd()
	defer func() {
		if wd != "" {
			os.Chdir(wd)
		}
	}()

	name := filepath.Base(pkg)
	temp, err := os.MkdirTemp("", fmt.Sprintf("fyne-install-%s-*", name))
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(temp)

	repo, err := vcs.RepoRootForImportPath(pkg, false)
	if err != nil {
		return fmt.Errorf("failed to look up source control for package: %w", err)
	}
	if repo.VCS.Name != "Git" {
		return errors.New("failed to find git repository: " + repo.VCS.Name)
	}

	if branch == "latest" {
		branch, err = getLatestTag(repo.Repo)
		if err != nil {
			return fmt.Errorf("failed to get latest tag: %v", err)
		}
		if i.verbose {
			fmt.Println("Latest tag:", branch)
		}
	}

	args := []string{"clone", repo.Repo, "--depth=1"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, temp)

	cmd := exec.Command("git", args...)
	if i.verbose {
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	path := getInstallBaseDir(temp, pkg, repo.Root)

	if !util.Exists(path) { // the error above may be ignorable, unless the path was not found
		return fmt.Errorf("path doesn't exist: %v", err)
	}

	if i.icon == "" {
		meta, err := metadata.LoadStandard(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to load metadata: %w", err)
		}
		if meta != nil {
			i.icon = filepath.Join(path, meta.Details.Icon)
		}
	}
	i.srcDir = path
	i.release = true
	if err := i.validate(); err != nil {
		return fmt.Errorf("failed to set up installer: %w", err)
	}

	return i.install()
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
	if i.verbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
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
	i.Packager.tags = i.tags
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
