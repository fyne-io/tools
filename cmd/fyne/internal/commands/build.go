package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mcuadros/go-version"
	"github.com/urfave/cli/v2"

	"fyne.io/fyne/v2"

	"fyne.io/tools/cmd/fyne/internal/goos"
	"fyne.io/tools/cmd/fyne/internal/metadata"
	"fyne.io/tools/cmd/fyne/internal/templates"
	"fyne.io/tools/cmd/fyne/internal/util"
)

// Partly based on https://gitlab.com/freedesktop-sdk/freedesktop-sdk/-/blob/master/include/flags.yml?ref_type=heads.
const (
	baseCFLAGSRegular = "-O2 -g -fexceptions -fasynchronous-unwind-tables -pipe"
	baseCFLAGSRelease = "-O3 -pipe"
)

// Names of the files that are generated in the source directory for a build.
const (
	pprofInitFileName    = "fyne_pprof.go"
	metadataInitFileName = "fyne_metadata_init.go"
)

// Builder generate the executables.
type Builder struct {
	*appData
	os, srcdir, target string
	goPackage          string
	release            bool
	pprof              bool
	pprofPort          int
	tags               []string
	tagsToParse        string
	verbose            bool

	customMetadata keyValueFlag

	goBin  string
	runner runner
}

// NewBuilder returns a command that can handle the build of GUI apps built using Fyne.
func NewBuilder() *Builder {
	return &Builder{appData: &appData{}}
}

// Build returns the cli command for building fyne applications
func Build() *cli.Command {
	b := NewBuilder()

	return &cli.Command{
		Name:        "build",
		Aliases:     []string{"b"},
		Usage:       "Builds an application",
		Description: "You can specify --target to define the OS to build for. The executable file will default to an appropriate name but can be overridden using -o.",
		Flags: []cli.Flag{
			stringFlags["target"](&b.os),
			stringFlags["src"](&b.srcdir),
			stringFlags["tags"](&b.tagsToParse),
			boolFlags["release"](&b.release),
			stringFlags["output"](&b.target),
			boolFlags["pprof"](&b.pprof),
			intFlags["pprof-port"](&b.pprofPort),
			genericFlags["metadata"](&b.customMetadata),
			boolFlags["verbose"](&b.verbose),
		},
		Action: func(ctx *cli.Context) error {
			argCount := ctx.Args().Len()
			if argCount > 0 {
				if argCount != 1 {
					return errors.New("incorrect amount of path provided")
				}
				b.goPackage = ctx.Args().First()
			}

			return b.Build()
		},
	}
}

// Build parse the tags and start building
func (b *Builder) Build() error {
	if b.srcdir != "" {
		b.srcdir = pkgUtil.EnsureAbsPath(b.srcdir)
		dirStat, err := os.Stat(b.srcdir)
		if err != nil {
			return err
		}
		if !dirStat.IsDir() {
			return errors.New("specified source directory is not a valid directory")
		}
	}
	if b.tagsToParse != "" {
		b.tags = util.SplitComma(b.tagsToParse)
	}
	b.Release = b.release
	b.CustomMetadata = b.customMetadata.m

	return b.build()
}

func (b *Builder) build() error {
	osTarget := b.os
	if osTarget == "" {
		osTarget = targetOS()
	}

	if goos.IsMobile(osTarget) {
		// a mobile application is created by gomobile, a plain go build would
		// silently produce a binary for this machine instead
		return fmt.Errorf("the build command cannot target %s, mobile packages are created with \"fyne package --target %s\"",
			osTarget, osTarget)
	}

	b.updateGoExecutable()

	srcdir, err := b.computeSrcDir()
	if err != nil {
		return err
	}

	b.updateToDefaultIconIfNotSet(srcdir)

	if b.verbose {
		exe := relPath(b.exePath(osTarget))
		if dir := relDir(b.srcdir); dir == "." {
			fmt.Println("Building", exe, "for", osTarget)
		} else {
			fmt.Println("Building", exe, "for", osTarget, "in", dir)
		}
	}

	if b.pprof {
		close, err := injectPprofFile(srcdir, b.pprofPort)
		if err != nil {
			fyne.LogError("Failed to inject pprof file, omitting pprof", err)
		} else if close != nil {
			if b.verbose {
				fmt.Println("Injecting pprof file", filepath.Join(relDir(srcdir), pprofInitFileName),
					"(removed after the build)")
			}
			defer close()
		}
	}

	close, err := injectMetadataIfPossible(srcdir, b.appData, createMetadataInitFile)
	if err != nil {
		fyne.LogError("Failed to inject metadata init file, omitting metadata", err)
	} else if close != nil {
		if b.verbose {
			fmt.Println("Injecting metadata file", filepath.Join(relDir(srcdir), metadataInitFileName),
				"(removed after the build)")
		}
		defer close()
	}

	args := []string{"build"}
	env := os.Environ()

	ldFlags := extractLdflagsFromGoFlags()
	if osTarget == goos.Windows {
		ldFlags += " -H=windowsgui"
	}

	if b.release {
		ldFlags += " -s -w"
		args = append(args, "-trimpath")
	}

	if len(ldFlags) > 0 {
		args = append(args, "-ldflags", strings.TrimSpace(ldFlags))
	}

	if b.target != "" {
		args = append(args, "-o", b.target)
	}

	if !goos.IsWeb(osTarget) {
		env = append(env, "CGO_ENABLED=1") // in case someone is trying to cross-compile...
		b.applyCAndLDFlags(&env, osTarget)
	} else {
		env = append(env, "CGO_ENABLED=0") // CGO is not available in WebAssembly
	}

	// handle build tags
	tags := b.tags
	if b.release {
		tags = append(tags, "release")
	}
	if ok, set := b.Migrations["fyneDo"]; ok && set {
		tags = append(tags, "migrated_fynedo")
	}
	if len(tags) > 0 {
		args = append(args, "-tags", util.JoinComma(tags))
	}

	if b.goPackage != "" {
		args = append(args, b.goPackage)
	}

	if osTarget != goos.IOS && osTarget != goos.Android && !goos.IsWeb(osTarget) {
		env = append(env, "GOOS="+osTarget)
	} else if goos.IsWASM(osTarget) {
		env = append(env, "GOARCH=wasm")
		env = append(env, "GOOS=js")
	}

	b.runner.setDir(b.srcdir)
	b.runner.setEnv(env)
	if b.verbose {
		fmt.Println("Running", util.JoinSpace(append([]string{b.goBin}, args...)))
	}
	out, err := b.runner.runOutput(args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
	}
	return err
}

func (b *Builder) computeSrcDir() (string, error) {
	if b.goPackage == "" || b.goPackage == "." {
		return b.srcdir, nil
	}

	srcdir := b.srcdir
	if strings.HasPrefix(b.goPackage, "."+string(os.PathSeparator)) ||
		strings.HasPrefix(b.goPackage, ".."+string(os.PathSeparator)) {
		srcdir = filepath.Join(srcdir, b.goPackage)
	} else if strings.HasPrefix(b.goPackage, string(os.PathSeparator)) {
		srcdir = b.goPackage
	} else {
		return "", fmt.Errorf("unrecognized go package: %s", b.goPackage)
	}
	return srcdir, nil
}

func (b *Builder) updateToDefaultIconIfNotSet(srcdir string) {
	if b.icon == "" {
		defaultIcon := filepath.Join(srcdir, "Icon.png")
		if pkgUtil.Exists(defaultIcon) {
			b.icon = defaultIcon
		}
	}
}

func injectPprofFile(srcdir string, port int) (func(), error) {
	pprofInitFilePath := filepath.Join(srcdir, pprofInitFileName)
	pprofInitFile, err := os.Create(pprofInitFilePath)
	if err != nil {
		return func() {}, err
	}
	defer pprofInitFile.Close()

	pprofInfo := struct {
		Port int
	}{
		Port: port,
	}

	err = templates.FynePprofInit.Execute(pprofInitFile, pprofInfo)
	if err != nil {
		os.Remove(pprofInitFilePath)
		return func() {}, err
	}

	return func() { os.Remove(pprofInitFilePath) }, nil
}

func (b *Builder) updateGoExecutable() {
	goBin := os.Getenv("GO")
	if goBin == "" {
		goBin = "go"
	}
	b.goBin = goBin
	if b.runner != nil {
		return
	}
	b.runner = newCommand(b.goBin)
}

// exePath returns the path of the executable that this build will create,
// following the naming that the go tool applies to the output.
func (b *Builder) exePath(osTarget string) string {
	if b.target != "" {
		return b.target
	}

	name := b.goPackage
	if name == "" || name == "." {
		name = calculateExeName(absDir(b.srcdir), osTarget)
	} else {
		// the go tool names the output after the package that is built
		name = filepath.Base(name)
		if osTarget == goos.Windows && !strings.HasSuffix(name, ".exe") {
			name += ".exe"
		}
	}

	return filepath.Join(absDir(b.srcdir), name)
}

// absDir returns the absolute path of dir, defaulting to the current directory.
func absDir(dir string) string {
	if dir != "" {
		return dir
	}
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// relDir returns the path of dir relative to the current directory, which is
// the short form that is easiest to recognise when reading build output.
// Directories outside of the current one, such as install destinations, are
// returned as absolute paths.
func relDir(dir string) string {
	abs := absDir(dir)

	wd, err := os.Getwd()
	if err != nil {
		return abs
	}

	rel, err := filepath.Rel(wd, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return abs
	}
	return rel
}

// relPath returns the path of a created file or directory like relDir does,
// but names a file in the current directory as "./name" to make it clear that
// it is written there.
func relPath(path string) string {
	rel := relDir(path)
	// a path that has no directory part is a file in the current one, note
	// that relDir may return a path that uses a foreign separator or none at
	// all, so ask for its directory instead of looking for a separator
	if rel != "." && filepath.Dir(rel) == "." {
		return "./" + rel
	}

	return rel
}

func (b *Builder) applyCAndLDFlags(env *[]string, os string) {
	cflags := []string{baseCFLAGSRegular}
	if b.release {
		cflags[0] = baseCFLAGSRelease
	}

	arch := targetArch()
	ccVer, err := ccVersionAndDefaultFlags()
	if err != nil {
		fyne.LogError("failed to get compiler version and default flags", err)
		ccVer = ccVersion()
	}
	cflagsHardening := hardeningCFlagsLookup(ccVer, os, arch)
	if cflagsHardening != "" {
		cflags = append(cflags, cflagsHardening)
	}

	ldflags := []string{}
	switch os {
	case goos.Linux:
		ldflags = append(ldflags, hardeningLDFLAGSLinux)
	case goos.Darwin:
		ldflags = append(ldflags, hardeningLDFLAGSDarwin)

		cflags = append(cflags, "-mmacosx-version-min=10.13")
		ldflags = append(ldflags, "-mmacosx-version-min=10.13")
	}

	switch arch {
	case "arm64":
		cflags = append(cflags, "-mbranch-protection=bti+pac-ret")
	}

	appendEnv(env, "CGO_CFLAGS", util.JoinSpace(cflags))
	appendEnv(env, "CGO_LDFLAGS", util.JoinSpace(ldflags))
}

const maxIconSize = 512

func createMetadataInitFile(srcdir string, app *appData) (func(), error) {
	data, err := metadata.LoadStandard(srcdir)
	if err == nil {
		// When icon path specified in metadata file, we should make it relative to metadata file
		if data.Details.Icon != "" {
			data.Details.Icon = pkgUtil.MakePathRelativeTo(srcdir, data.Details.Icon)
		}

		app.mergeMetadata(data)
	}

	metadataInitFilePath := filepath.Join(srcdir, metadataInitFileName)
	metadataInitFile, err := os.Create(metadataInitFilePath)
	if err != nil {
		return func() {}, err
	}
	defer metadataInitFile.Close()

	app.ResGoString = "nil"
	if app.icon != "" {
		res, err := fyne.LoadResourceFromPath(app.icon)
		if err != nil {
			fyne.LogError("Unable to load metadata icon file "+app.icon, err)
			return func() { os.Remove(metadataInitFilePath) }, err
		}

		res = metadata.ScaleIcon(res, maxIconSize)

		// The return type of fyne.LoadResourceFromPath is always a *fyne.StaticResource.
		app.ResGoString = res.(*fyne.StaticResource).GoString()
	}

	err = templates.FyneMetadataInit.Execute(metadataInitFile, app)
	if err != nil {
		fyne.LogError("Error executing metadata template", err)
	}

	return func() { os.Remove(metadataInitFilePath) }, err
}

func injectMetadataIfPossible(srcdir string, app *appData,
	createMetadataInitFile func(srcdir string, app *appData) (func(), error),
) (func(), error) {
	fyneGoModVersion, err := getFyneGoModVersion(srcdir)
	if err != nil {
		return nil, err
	}

	fyneGoModVersion = normaliseVersion(fyneGoModVersion)
	fyneGoModVersionConstraint := version.NewConstrainGroupFromString(">=2.2")
	if fyneGoModVersion != "master" && !fyneGoModVersionConstraint.Match(fyneGoModVersion) {
		return nil, nil
	}

	fyneGoModVersionAtLeast2_3 := version.NewConstrainGroupFromString(">=2.3")
	if fyneGoModVersionAtLeast2_3.Match(fyneGoModVersion) {
		app.VersionAtLeast2_3 = true
	}
	fyneGoModVersionAtLeast2_6 := version.NewConstrainGroupFromString(">=2.6")
	if fyneGoModVersionAtLeast2_6.Match(fyneGoModVersion) {
		app.VersionAtLeast2_6 = true
	}

	return createMetadataInitFile(srcdir, app)
}

func targetArch() string {
	archEnv, ok := os.LookupEnv("GOARCH")
	if ok {
		return archEnv
	}

	return runtime.GOARCH
}

func targetOS() string {
	osEnv, ok := os.LookupEnv("GOOS")
	if ok {
		return osEnv
	}

	return runtime.GOOS
}

func appendEnv(env *[]string, varName, value string) {
	for i := range *env {
		keyValue := strings.SplitN((*env)[i], "=", 2)

		if keyValue[0] == varName {
			(*env)[i] += " " + value
			return
		}
	}

	*env = append(*env, varName+"="+value)
}

const (
	goflagsEnvKey = "GOFLAGS"
)

func extractLdflagsFromGoFlags() string {
	goFlags := os.Getenv(goflagsEnvKey)

	ldFlags, goFlags := extractLdFlags(goFlags)
	if goFlags != "" {
		os.Setenv(goflagsEnvKey, goFlags)
	} else {
		os.Unsetenv(goflagsEnvKey)
	}

	return ldFlags
}

func extractLdFlags(goFlags string) (string, string) {
	if goFlags == "" {
		return "", ""
	}
	var ldflags, newGoFlags []string
	for _, flag := range strings.Fields(goFlags) {
		if strings.HasPrefix(flag, "-ldflags=") {
			ldflags = append(ldflags, strings.TrimPrefix(flag, "-ldflags="))
		} else {
			newGoFlags = append(newGoFlags, flag)
		}
	}
	return util.JoinSpace(ldflags), util.JoinSpace(newGoFlags)
}

func normaliseVersion(str string) string {
	if str == "master" {
		return str
	}

	if pos := strings.Index(str, "-0.20"); pos != -1 {
		str = str[:pos] + "-dev"
	}

	if pos := strings.Index(str, "-rc"); pos != -1 {
		str = str[:pos] + "-dev"
	}
	return version.Normalize(str)
}
