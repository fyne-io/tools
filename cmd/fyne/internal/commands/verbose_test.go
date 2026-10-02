package commands

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"fyne.io/tools/cmd/fyne/internal/goos"
	"fyne.io/tools/cmd/fyne/internal/util"
)

// captureStdout returns everything the given function prints to stdout.
func captureStdout(t *testing.T, run func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	stdout := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = stdout }()

	run()

	require.NoError(t, writer.Close())
	out, err := io.ReadAll(reader)
	require.NoError(t, err)

	return string(out)
}

func Test_VerboseFlagIsAvailable(t *testing.T) {
	for name, cmd := range map[string]*cli.Command{
		"build":   Build(),
		"package": Package(),
		"install": Install(),
	} {
		var found bool
		for _, flag := range cmd.Flags {
			if slices.Contains(flag.Names(), "verbose") {
				found = true
			}
		}

		assert.True(t, found, "the %s command should offer a verbose flag", name)
	}
}

func Test_BuildVerboseOutput(t *testing.T) {
	t.Setenv("GO", "")

	useMockPkgUtil(t)
	utilExistsMock = func(string) bool { return false } // no icon in the source dir

	expected := []mockRunner{
		{
			expectedValue: expectedValue{
				args:  []string{"build", "-o", "myapp.wasm"},
				env:   []string{"GOARCH=wasm", "GOOS=js", "CGO_ENABLED=0"},
				osEnv: true,
				dir:   "myTest",
			},
			mockReturn: mockReturn{
				ret: []byte(""),
			},
		},
	}

	b := &Builder{
		appData: &appData{},
		os:      goos.WASM,
		srcdir:  "myTest",
		target:  "myapp.wasm",
		verbose: true,
	}
	buildTest := &testCommandRuns{runs: expected, t: t}
	b.runner = buildTest

	var err error
	out := captureStdout(t, func() {
		err = b.build()
	})

	assert.NoError(t, err)
	buildTest.verifyExpectation()
	assert.Contains(t, out, "Building ./myapp.wasm for wasm in myTest")
	assert.Contains(t, out, "Running go build -o myapp.wasm")
}

func Test_BuilderExePath(t *testing.T) {
	tests := []struct {
		name      string
		builder   Builder
		osTarget  string
		expectExe string
	}{
		{
			name:      "output flag wins",
			builder:   Builder{target: filepath.Join("bin", "custom")},
			osTarget:  goos.Linux,
			expectExe: filepath.Join("bin", "custom"),
		},
		{
			name:      "named package",
			builder:   Builder{goPackage: "./cmd/foo"},
			osTarget:  goos.Linux,
			expectExe: "foo",
		},
		{
			name:      "named package for windows",
			builder:   Builder{goPackage: "example.com/myapp"},
			osTarget:  goos.Windows,
			expectExe: "myapp.exe",
		},
		{
			name:      "source directory module",
			builder:   Builder{srcdir: "testdata/modules_app"},
			osTarget:  goos.Linux,
			expectExe: filepath.Join("testdata", "modules_app", "module"),
		},
		{
			name:      "source directory module for windows",
			builder:   Builder{srcdir: "testdata/modules_app"},
			osTarget:  goos.Windows,
			expectExe: filepath.Join("testdata", "modules_app", "module.exe"),
		},
		{
			name:      "current directory",
			builder:   Builder{},
			osTarget:  goos.Linux,
			expectExe: filepath.Base(absDir("")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exe := test.builder.exePath(test.osTarget)
			if !filepath.IsAbs(test.expectExe) {
				exe = relDir(exe) // the test cases are relative to the current directory
			}
			assert.Equal(t, test.expectExe, exe)
		})
	}
}

func Test_RelDir(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	assert.Equal(t, ".", relDir(""))                            // an unset directory is the current one
	assert.Equal(t, ".", relDir(wd))                            // ... which is named relatively
	assert.Equal(t, "sub", relDir("sub"))                       // a child of the current directory
	assert.Equal(t, "/etc", relDir("/etc"))                     // outside keeps the absolute path
	assert.Equal(t, filepath.Dir(wd), relDir(filepath.Dir(wd))) // above too
}

func Test_RelPath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	assert.Equal(t, "./myapp", relPath(filepath.Join(wd, "myapp"))) // a file next to us is named "./"
	assert.Equal(t, filepath.Join("sub", "myapp"), relPath(filepath.Join(wd, "sub", "myapp")))
	assert.Equal(t, "/etc/myapp", relPath("/etc/myapp")) // outside keeps the absolute path
}

func Test_InstallAndroidVerboseOutput(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "myapp.apk"), nil, util.FilePermDefault))

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer func() {
		require.NoError(t, os.Chdir(wd))
	}()

	t.Setenv("PATH", t.TempDir()) // make sure that no adb is found

	i := &Installer{
		appData: &appData{},
		os:      "android/arm64",
		verbose: true,
		Packager: &Packager{
			appData: &appData{Name: "myapp"},
		},
	}

	out := captureStdout(t, func() {
		err = i.installAndroid()
	})

	assert.Error(t, err) // adb is not available, so the install fails
	assert.Contains(t, out, "Using existing package myapp.apk")
	assert.Contains(t, out, "Installing myapp.apk")
}
