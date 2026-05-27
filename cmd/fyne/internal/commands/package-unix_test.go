package commands

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"fyne.io/fyne/v2/theme"
	"fyne.io/tools/cmd/fyne/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestFormatDesktopFileList(t *testing.T) {
	assert.Equal(t, "", formatDesktopFileList([]string{}))
	assert.Equal(t, "One;", formatDesktopFileList([]string{"One"}))
	assert.Equal(t, "One;Two;", formatDesktopFileList([]string{"One", "Two"}))
}

func TestDesktopFileSource(t *testing.T) {
	tplData := unixData{
		Name: "Testing",
	}
	buf := &bytes.Buffer{}

	err := templates.DesktopFileUNIX.Execute(buf, tplData)
	assert.Nil(t, err)
	assert.False(t, strings.Contains(buf.String(), "[X-Fyne"))

	tplData.SourceRepo = "https://example.com"
	tplData.SourceDir = "cmd/name"

	err = templates.DesktopFileUNIX.Execute(buf, tplData)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(buf.String(), "[X-Fyne"))
	assert.True(t, strings.Contains(buf.String(), "Repo=https://example.com"))
	assert.True(t, strings.Contains(buf.String(), "Dir=cmd/name"))
}

func TestDesktopFileOpenWith(t *testing.T) {
	tplData := unixData{
		Name:      "Testing",
		Exec:      "tester",
		MimeTypes: "text/plain",
	}
	buf := &bytes.Buffer{}

	err := templates.DesktopFileUNIX.Execute(buf, tplData)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(buf.String(), "MimeType=text/plain"))
}

func TestDesktopFile(t *testing.T) {
	tplData := unixData{
		Name:           "Testing",
		StartupWMClass: "Testing",
	}
	buf := &bytes.Buffer{}

	err := templates.DesktopFileUNIX.Execute(buf, tplData)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(buf.String(), "StartupWMClass=Testing"))
}

const testAppTarFiles = `testapp/
testapp/Makefile
testapp/usr/
testapp/usr/local/
testapp/usr/local/bin/
testapp/usr/local/bin/testapp
testapp/usr/local/share/
testapp/usr/local/share/applications/
testapp/usr/local/share/applications/testapp.desktop
testapp/usr/local/share/pixmaps/
testapp/usr/local/share/pixmaps/testapp.png
`

func TestPackageLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping test that needs linux as host os")
		return
	}

	if os.Getenv("FYNE_TEST_PACKAGE_LINUX") != "yes" {
		t.Skip("skipping test that needs all Fyne libraries")
		return
	}

	basedir := t.TempDir()
	dir := filepath.Join(basedir, "testapp")

	assert.NoError(t, os.Mkdir(dir, 0o755), "create dir for test app")
	assert.NoError(t, os.Chdir(dir), "change into new app dir")
	assert.NoError(t, os.WriteFile("Icon.png", theme.FyneLogo().Content(), 0o644)) //lint:ignore SA1019 It is fine for our own use.

	app := &cli.App{
		Name: "fyne",
		Commands: []*cli.Command{
			Init(),
			Translate(), // init uses translate
			Package(),
		},
	}

	assert.NoError(t, app.Run([]string{"fyne", "init", "-v"}), "fyne init")
	assert.NoError(t, app.Run([]string{"fyne", "package", "-os", "linux"}), "fyne package")

	buf := &bytes.Buffer{}
	cmd := exec.Command("tar", "-tJf", "testapp.tar.xz")
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	assert.NoError(t, cmd.Run(), "show archive")
	assert.Equal(t, testAppTarFiles, buf.String())
}
