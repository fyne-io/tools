//go:build go1.24

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/tools/cmd/fyne/internal/commands/cross"
	"github.com/urfave/cli/v2"
)

type crossCmd struct {
	*appData
	cacheDir  string
	goPackage string
	goarch    string
	goos      string
	image     string
	ldFlags   cli.StringSlice
	noCache   bool
	output    string
	pull      bool
	release   bool
	srcdir    string
	tags      cli.StringSlice
	target    string
	verbose   bool
}

// Cross returns the command for cross-compiling Fyne applications using containers.
func Cross() *cli.Command {
	c := &crossCmd{appData: &appData{}}

	return &cli.Command{
		Name:    "cross",
		Aliases: []string{"x"},
		Usage:   "Cross-compile Fyne applications using containers",
		Flags: []cli.Flag{
			stringFlags["target"](&c.target),
			stringFlags["src"](&c.srcdir),
			boolFlags["release"](&c.release),
			stringFlags["output"](&c.output),
			&cli.StringFlag{
				Name:        "image",
				Usage:       "specify custom container image (overrides default)",
				Destination: &c.image,
			},
			&cli.BoolFlag{
				Name:        "pull",
				Usage:       "pull latest container image",
				Value:       false,
				Destination: &c.pull,
			},
			&cli.StringFlag{
				Name:        "cache-dir",
				Usage:       "directory for Go build and Zig compiler caches",
				Destination: &c.cacheDir,
				DefaultText: c.defaultCacheDir(),
				Value:       c.defaultCacheDir(),
			},
			&cli.BoolFlag{
				Name:        "no-cache",
				Usage:       "disable Go build and Zig compiler caching",
				Destination: &c.noCache,
			},
			&cli.StringSliceFlag{
				Name:        "ldflags",
				Usage:       "specify linker flags to pass to go build (can be used multiple times)",
				Destination: &c.ldFlags,
			},
			&cli.StringSliceFlag{
				Name:        "tags",
				Usage:       "specify build tags (can be used multiple times)",
				Destination: &c.tags,
			},
			boolFlags["verbose"](&c.verbose),
		},
		Action: func(ctx *cli.Context) error {
			return c.build(ctx.Context)
		},
		Before: func(ctx *cli.Context) error {
			argCount := ctx.Args().Len()
			if argCount > 0 {
				if argCount != 1 {
					return fmt.Errorf("incorrect amount of path provided")
				}
				c.goPackage = ctx.Args().First()
			}
			return c.validate()
		},
	}
}

// build executes the cross-compilation build process in a container.
// It sets up the Docker container engine, configures the go build command with
// all flags and options, and executes it inside the container.
func (c *crossCmd) build(ctx context.Context) error {
	// Set release flag in appData
	c.appData.Release = c.release

	// Create engine with target OS/arch (already validated in Before hook)
	engine, err := cross.NewEngine(c.goos, c.goarch)
	if err != nil {
		return fmt.Errorf("failed to create container engine: %w", err)
	}
	defer engine.Close()

	// Configure engine
	engine.WithImage(c.image)
	engine.WithVerbose(c.verbose)

	// Configure mounts
	engine.WithSourceMount(c.srcdir)

	// Setup caching if not disabled
	if !c.noCache {
		engine.WithCache(c.cacheDir)
	}

	// Enable debug stripping for release builds
	if c.release {
		engine.WithStripDebug()
	}

	// Pull image if requested
	if c.pull {
		fmt.Printf("Pulling image %s...\n", engine.Image())
		if err := engine.PullImage(ctx); err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
	}

	// Build go build command
	goCmd := cross.NewGoBuildCmd()
	goCmd.WithVerbose(c.verbose)
	goCmd.WithTrimpath(c.release)

	if c.output != "" {
		goCmd.WithOutput(c.output)
	}

	// Add ldflags
	for _, flag := range c.ldFlags.Value() {
		goCmd.WithLDFlag(flag)
	}

	// Add build tags
	if c.release {
		goCmd.WithTag("release")
	}
	if ok, set := c.appData.Migrations["fyneDo"]; ok && set {
		goCmd.WithTag("migrated_fynedo")
	}
	for _, tag := range c.tags.Value() {
		goCmd.WithTag(tag)
	}

	if c.goPackage != "" {
		goCmd.WithPackage(c.goPackage)
	}

	// Show configuration and command in verbose mode
	if c.verbose {
		fmt.Printf("Container configuration:\n%s\n\n", engine.Debug())
		fmt.Printf("Go build command: %s\n\n", goCmd.String())
	}

	fmt.Printf("Building for %s/%s in container...\n", c.goos, c.goarch)
	return engine.Exec(ctx, goCmd)
}

// computeSrcDir computes and returns the absolute path to the source directory
// based on the srcdir and goPackage fields. It handles relative paths starting
// with "./" or "../" by joining them with the source directory.
//
// NOTE: has been ported from internal/command/build.go
// TODO: refactor to have a common helper func?
func (c *crossCmd) computeSrcDir() (string, error) {
	if c.srcdir != "" {
		c.srcdir = util.EnsureAbsPath(c.srcdir)
		dirStat, err := os.Stat(c.srcdir)
		if err != nil {
			return "", err
		}
		if !dirStat.IsDir() {
			return "", fmt.Errorf("specified source directory is not a valid directory")
		}
	}

	srcdir, err := filepath.Abs(c.srcdir)
	if err != nil {
		return "", err
	}
	if c.goPackage == "" || c.goPackage == "." {
		return srcdir, nil
	}

	if strings.HasPrefix(c.goPackage, "."+string(os.PathSeparator)) ||
		strings.HasPrefix(c.goPackage, ".."+string(os.PathSeparator)) {
		srcdir = filepath.Join(srcdir, c.goPackage)
	}
	return srcdir, nil
}

// defaultCacheDir returns the default cache directory path.
// Tries $HOME/.cache/fyne-tools first, falls back to $TMPDIR/fyne-tools if that fails.
func (c *crossCmd) defaultCacheDir() string {
	// Try user cache directory first
	userCacheDir, err := os.UserCacheDir()
	if err == nil {
		return filepath.Join(userCacheDir, "fyne-tools")
	}

	// Fall back to temp directory
	return filepath.Join(os.TempDir(), "fyne-tools")
}

// validate performs all validation and setup before the build action.
// It parses the target OS/arch, computes and validates the source directory,
// and sets up the cache directory.
func (c *crossCmd) validate() error {
	// Parse target OS/arch
	goos, goarch, _ := strings.Cut(c.target, "/")
	if goarch == "" {
		goarch = targetArch()
	}
	if goos == "" {
		goos = targetOS()
	}
	c.goos = goos
	c.goarch = goarch

	// Compute and validate source directory
	srcdir, err := c.computeSrcDir()
	if err != nil {
		return err
	}
	c.srcdir = util.EnsureAbsPath(srcdir)

	// Set up cache directory
	if !c.noCache {
		if c.cacheDir == "" {
			c.cacheDir = c.defaultCacheDir()
		}
		c.cacheDir = util.EnsureAbsPath(c.cacheDir)
		if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory: %w", err)
		}
	}

	return nil
}
